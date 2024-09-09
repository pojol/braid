package actor

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	trhreids "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/def"
	"github.com/redis/go-redis/v9"
)

type BlockLoader struct {
	BlockName string
	BlockType reflect.Type

	Ins      interface{}
	oldBytes []byte
}

// EntityLoader entity装载器
type EntityLoader struct {
	ID      string
	Loaders []BlockLoader
}

func BuildEntityLoader(id string, loaderTypes []reflect.Type) *EntityLoader {
	loaders := make([]BlockLoader, 0, len(loaderTypes))
	for _, loader := range loaderTypes {
		loaders = append(loaders, BlockLoader{
			BlockName: getTypeName(loader),
			BlockType: loader,
		})
	}
	return &EntityLoader{ID: id, Loaders: loaders}
}

func getTypeName(t reflect.Type) string {
	// 如果是指针类型，获取其元素类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 获取完整的类型名（包括包名）
	fullName := t.String()

	// 分割包名和类型名
	parts := strings.Split(fullName, ".")

	// 返回最后一部分作为类型名
	return parts[len(parts)-1]
}

func (loader *EntityLoader) Load(ctx context.Context) error {

	if len(loader.Loaders) == 0 {
		return def.ErrEntityLoadEntityLoadersEmpty(loader.ID)
	}

	var cmds []redis.Cmder

	cmds, err := trhreids.TxPipelined(ctx, "[EntityLoader.Load]", func(pipe redis.Pipeliner) error {
		for _, load := range loader.Loaders {
			key := fmt.Sprintf("{%s}_%s", loader.ID, load.BlockName)
			pipe.Get(ctx, key)
		}
		return nil
	})
	if err != nil { // 装载到空entity也是一个错误， 如果有一个 block 不存在，这边就会返回 err
		return err
	}

	var bytSlice [][]byte
	bytSlice, err = trhreids.GetCmdsByteSlice(cmds)
	if err != nil {
		return err
	}

	for idx, load := range loader.Loaders {

		protoMsg := reflect.New(load.BlockType.Elem()).Interface().(proto.Message)

		if len(bytSlice[idx]) == 0 {
			return fmt.Errorf("load block %s is not empty", load.BlockName)
		}

		// 反序列化 protobuf 数据
		if err := proto.Unmarshal(bytSlice[idx], protoMsg); err != nil {
			return def.ErrEntityLoadUnpack(loader.ID, load.BlockName)
		}

		loader.Loaders[idx].oldBytes = bytSlice[idx]
		loader.Loaders[idx].Ins = protoMsg
	}

	return nil
}

func (loader *EntityLoader) Unload(ctx context.Context) error {

	if len(loader.Loaders) == 0 {
		return def.ErrEntityLoadEntityLoadersEmpty(loader.ID)
	}

	_, err := trhreids.TxPipelined(ctx, "[EntityLoader.Unload]", func(pipe redis.Pipeliner) error {
		for idx, load := range loader.Loaders {
			if loader.Loaders[idx].Ins == nil {
				fmt.Println("unload", load.BlockName, "Ins is nil")
			}

			byt, err := proto.Marshal(loader.Loaders[idx].Ins.(proto.Message))
			if err != nil {
				return err
			}

			if !bytes.Equal(loader.Loaders[idx].oldBytes, byt) {
				key := fmt.Sprintf("{%s}_%s", loader.ID, load.BlockName)
				pipe.Set(ctx, key, byt, 0)
			}
		}
		return nil
	})

	return err

}

func (loader *EntityLoader) Sync(ctx context.Context) error {
	return loader.Unload(ctx)
}

func (loader *EntityLoader) Clean(ctx context.Context) error {

	_, err := trhreids.TxPipelined(ctx, "[EntityLoader.Clean]", func(pipe redis.Pipeliner) error {
		for _, load := range loader.Loaders {
			key := fmt.Sprintf("{%s}_%s", loader.ID, load.BlockName)
			pipe.Del(ctx, key)
		}
		return nil
	})

	return err
}

func (loader *EntityLoader) GetModule(typ reflect.Type) interface{} {
	for _, load := range loader.Loaders {
		if load.BlockType == typ {
			return load.Ins
		}
	}
	return nil
}

func (loader *EntityLoader) SetModule(typ reflect.Type, module interface{}) {

	flag := false

	for idx, load := range loader.Loaders {
		if load.BlockType == typ {
			loader.Loaders[idx].Ins = module
			flag = true
		}
	}

	if !flag {
		fmt.Println("set module not found", typ)
	}
}
