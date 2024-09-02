package timer

import (
	"fmt"
	"time"
)

type TimerTask struct {
	TimerInfo
	Interval int64 `json:"interval"`
}

// Start start task
func (tt *TimerTask) start() {
	if !tt.isRun {
		return
	}

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	if tt.State == Timer_state_init || tt.State == Timer_state_stop {
		tt.State = Timer_state_run
		startTickTask(tt)
	}
}

func (tt *TimerTask) runOnce() error {
	err := tt.handler(tt.context)
	return err
}

func NewTickTimer(taskID string, isRun bool, dueTime int64, interval int64, handler TimerHandleFunc, taskData interface{}) (Timer, error) {
	context := new(TimerContext)
	context.TimerID = taskID
	context.TimerData = taskData

	task := new(TimerTask)
	task.timerID = context.TimerID
	task.TaskType = Timer_type_tick
	task.isRun = isRun
	task.handler = handler
	task.DueTime = dueTime
	task.Interval = interval
	task.State = Timer_state_init
	task.context = context
	return task, nil
}

func startTickTask(task *TimerTask) {
	handler := func() {
		defer func() {
			if err := recover(); err != nil {
				//task.taskService.Logger().Debug(task.TaskID, " loop handler recover error => ", err)
				if task.taskService.ExceptionHandler != nil {
					task.taskService.ExceptionHandler(task.Context(), fmt.Errorf("%v", err))
				}
				//goroutine panic, restart cron task
				startTickTask(task)
			}
		}()
		if task.taskService != nil && task.taskService.OnBeforeHandler != nil {
			task.taskService.OnBeforeHandler(task.Context())
		}
		var err error
		if !task.Context().IsEnd {
			err = task.handler(task.Context())
		}
		if err != nil {
			if task.taskService != nil && task.taskService.ExceptionHandler != nil {
				task.taskService.ExceptionHandler(task.Context(), err)
			}
		}

		if task.taskService != nil && task.taskService.OnEndHandler != nil {
			task.taskService.OnEndHandler(task.Context())
		}
	}
	dofunc := func() {
		task.TimeTicker = time.NewTicker(time.Duration(task.Interval) * time.Millisecond)
		handler()
		for {
			<-task.TimeTicker.C
			handler()
		}
	}
	//等待设定的延时毫秒
	if task.DueTime > 0 {
		go time.AfterFunc(time.Duration(task.DueTime)*time.Millisecond, dofunc)
	} else {
		go dofunc()
	}

}
