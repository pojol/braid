package timer

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

const (
	Timer_state_init = "ts_init"
	Timer_state_run  = "ts_run"
	Timer_state_stop = "ts_stop"
)

const (
	Timer_type_tick = "tt_tick"
	Timer_type_cron = "tt_cron"
)

const (
	DefaultPeriod     = time.Second //默认执行周期
	defaultTimeLayout = "2006-01-02 15:04:05"
)

type TimerContext struct {
	TimerID   string
	TimerData interface{} //用于当前Task全局设置的数据项
	Message   interface{} //用于每次Task执行上下文消息传输
	IsEnd     bool        //如果设置该属性为true，则停止当次任务的后续执行，一般用在OnBegin中
}

type TimerHandleFunc func(*TimerContext) error
type ExceptionHandleFunc func(*TimerContext, error)

type Timer interface {
	TimerID() string
	Context() *TimerContext
	start()
	stop()
	runOnce() error
	SetTimerService(service *TimerService)
}

type TimerInfo struct {
	timerID     string
	isRun       bool
	taskService *TimerService
	mutex       sync.RWMutex
	TimeTicker  *time.Ticker
	TaskType    string
	handler     TimerHandleFunc
	context     *TimerContext
	State       string
	DueTime     int64 //开始任务的延迟时间（以毫秒为单位），如果<=0则不延迟
}

// Stop stop task
func (task *TimerInfo) stop() {
	if !task.isRun {
		return
	}
	if task.State == Timer_state_run {
		task.TimeTicker.Stop()
		task.State = Timer_state_stop
	}
}

func (task *TimerInfo) TimerID() string {
	return task.timerID
}

func (task *TimerInfo) Context() *TimerContext {
	return task.context
}

func (task *TimerInfo) SetTimerService(service *TimerService) {
	task.taskService = service
}

func (task *TimerInfo) runOnce() error {
	err := task.handler(task.context)
	return err
}

type TimerService struct {
	taskMap          map[string]Timer
	taskMutex        *sync.RWMutex
	handlerMap       map[string]TimerHandleFunc
	handlerMutex     *sync.RWMutex
	ExceptionHandler ExceptionHandleFunc
	OnBeforeHandler  TimerHandleFunc
	OnEndHandler     TimerHandleFunc
}

// 设置自定义异常处理方法
func (ts *TimerService) SetExceptionHandler(handler ExceptionHandleFunc) {
	ts.ExceptionHandler = handler
}

func (ts *TimerService) SetOnBeforeHandler(handler TimerHandleFunc) {
	ts.OnBeforeHandler = handler
}

func (ts *TimerService) SetOnEndHandler(handler TimerHandleFunc) {
	ts.OnEndHandler = handler
}

func Init() *TimerService {
	service := new(TimerService)
	service.taskMutex = new(sync.RWMutex)
	service.taskMap = make(map[string]Timer)
	service.handlerMutex = new(sync.RWMutex)
	service.handlerMap = make(map[string]TimerHandleFunc)

	// 默认异常处理
	service.ExceptionHandler = func(ctx *TimerContext, err error) {
		stack := string(debug.Stack())
		fmt.Println("taskid", ctx.TimerID, " error", err.Error(), " stack", stack)
	}
	return service
}

func (ts *TimerService) StartAllTask() {
	for _, v := range ts.taskMap {
		v.start()
	}
}

func (ts *TimerService) StopAllTask() {
	for _, v := range ts.taskMap {
		v.stop()
	}
}

func (ts *TimerService) NewTickTask(taskID string, isRun bool, dueTime int64, interval int64, handler TimerHandleFunc, taskData interface{}) (Timer, error) {
	task, err := NewTickTimer(taskID, isRun, dueTime, interval, handler, taskData)
	if err != nil {
		return task, err
	}
	task.SetTimerService(ts)
	ts.AddTask(task)
	return task, nil
}

func (service *TimerService) AddTask(t Timer) {
	service.taskMutex.Lock()
	service.taskMap[t.TimerID()] = t
	service.taskMutex.Unlock()
}

func (service *TimerService) RemoveTask(taskID string) {
	service.taskMutex.Lock()
	delete(service.taskMap, taskID)
	service.taskMutex.Unlock()
}

func (ts *TimerService) StopTask(taskID string) {
	if _, ok := ts.taskMap[taskID]; !ok {
		return
	}
	ts.taskMap[taskID].stop()
}

func (ts *TimerService) StartTask(taskID string) {
	if _, ok := ts.taskMap[taskID]; !ok {
		return
	}
	ts.taskMap[taskID].start()
}

func (ts *TimerService) Stop(name string) error {

	if v, ok := ts.taskMap[name]; ok {
		v.stop()
		return nil
	}

	return fmt.Errorf("not find task id %v", name)

}
