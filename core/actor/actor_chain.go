package actor

import (
	"fmt"
	"reflect"

	"github.com/pojol/braid/router"
	"github.com/pojol/braid/router/msg"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type DefaultChain struct {
	Before  []EventHandler
	After   []EventHandler
	Handler EventHandler
	Script  *ScriptHandler // 添加脚本处理器
}

// ScriptHandler represents a script-based handler
type ScriptHandler struct {
	interpreter *interp.Interpreter
	scriptPath  string
}

// NewScriptHandler creates a new script handler
func NewScriptHandler(scriptPath string) (*ScriptHandler, error) {
	i := interp.New(interp.Options{
		GoPath: "/path/to/your/scripts", // 设置脚本路径
	})

	// 加载标准库
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, fmt.Errorf("failed to load stdlib: %w", err)
	}

	// 加载自定义符号（比如 Wrapper 等类型）
	if err := i.Use(map[string]map[string]reflect.Value{
		"github.com/pojol/braid/router/router": {
			"Wrapper": reflect.ValueOf((*msg.Wrapper)(nil)),
			"Message": reflect.ValueOf((*router.Message)(nil)),
			"Header":  reflect.ValueOf((*router.Header)(nil)),
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to load custom symbols: %w", err)
	}

	// 加载脚本文件
	if _, err := i.EvalPath(scriptPath); err != nil {
		return nil, fmt.Errorf("failed to load script %s: %w", scriptPath, err)
	}

	return &ScriptHandler{
		interpreter: i,
		scriptPath:  scriptPath,
	}, nil
}

// NewScriptHandlerFromString creates a new script handler from string content
func NewScriptHandlerFromString(content string) (*ScriptHandler, error) {
	i := interp.New(interp.Options{})

	// 只加载标准库
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, fmt.Errorf("failed to load stdlib: %w", err)
	}

	// 加载自定义符号
	if err := i.Use(map[string]map[string]reflect.Value{
		"github.com/pojol/braid/router/router": {
			"Wrapper": reflect.ValueOf((*msg.Wrapper)(nil)),
			"Message": reflect.ValueOf(&router.Message{}),
			"Header":  reflect.ValueOf(&router.Header{}),
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to load custom symbols: %w", err)
	}

	// 直接从字符串加载脚本
	if _, err := i.Eval(content); err != nil {
		return nil, fmt.Errorf("failed to evaluate script content: %w", err)
	}

	return &ScriptHandler{
		interpreter: i,
	}, nil
}

// Execute runs the script handler
func (sh *ScriptHandler) Execute(m *msg.Wrapper) error {
	v, err := sh.interpreter.Eval("Execute")
	if err != nil {
		return fmt.Errorf("failed to get Execute function: %w", err)
	}

	fn, ok := v.Interface().(func(*msg.Wrapper) error)
	if !ok {
		return fmt.Errorf("invalid Execute function signature")
	}

	return fn(m)
}

func (c *DefaultChain) Execute(mw *msg.Wrapper) error {
	var err error

	for _, before := range c.Before {
		err = before(mw)
		if err != nil {
			goto ext
		}
	}

	if c.Script != nil {
		err = c.Script.Execute(mw)
	} else {
		err = c.Handler(mw)
	}
	if err != nil {
		goto ext
	}

	for _, after := range c.After {
		err = after(mw)
		if err != nil {
			goto ext
		}
	}

ext:
	return err
}
