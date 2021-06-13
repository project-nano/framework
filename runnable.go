package framework

import "errors"

type RoutineController interface {
	SetStopping()
	IsStopping() bool
	GetNotifyChannel() chan bool
	NotifyExit()
}

type SimpleRunner struct {
	notifyChan chan bool
	exitChan   chan bool
	exitFlag   bool
	routine    func(RoutineController)
	isRunning  bool
}

func CreateSimpleRunner (r func(RoutineController)) (runner *SimpleRunner){
	var notifyChan = make(chan bool, 1)
	var exitChan = make(chan bool, 1)
	runner = &SimpleRunner{
		notifyChan: notifyChan,
		exitChan:   exitChan,
		exitFlag:   false,
		routine:    r,
		isRunning:  false,
	}
	return runner
}

func (runner *SimpleRunner) Start() error{
	if runner.isRunning{
		return errors.New("already started")
	}
	go runner.routine(runner)
	runner.isRunning = true
	return nil
}

func (runner *SimpleRunner) Stop() error{
	if !runner.isRunning{
		return errors.New("already stopped")
	}
	runner.notifyChan <- true
	<- runner.exitChan
	runner.isRunning = false
	return nil
}

func (runner *SimpleRunner) IsRunning() bool{
	return runner.isRunning
}

func (runner *SimpleRunner) SetStopping(){
	runner.exitFlag = true
}
func (runner *SimpleRunner) IsStopping() bool{
	return runner.exitFlag
}

func (runner *SimpleRunner) GetNotifyChannel() chan bool{
	return runner.notifyChan
}

func (runner *SimpleRunner) NotifyExit(){
	runner.exitChan <- true
}
