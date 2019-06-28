package framework


type RoutineController interface {
	SetStopping()
	IsStopping() bool
	GetNotifyChannel() (chan bool)
	NotifyExit()
}

type SimpleRunner struct {
	notifyChan chan bool
	exitChan   chan bool
	exitFlag   bool
	routine    func(RoutineController)
}


func CreateSimpleRunner (r func(RoutineController)) (runner *SimpleRunner){
	var notifyChan = make(chan bool, 1)
	var exitChan = make(chan bool, 1)
	runner = &SimpleRunner{notifyChan, exitChan, false, r}
	return runner
}

func (runner *SimpleRunner)Start() error{
	go runner.routine(runner)
	return nil
}

func (runner *SimpleRunner)Stop() error{
	runner.notifyChan <- true
	<- runner.exitChan
	return nil
}

func (runner *SimpleRunner)SetStopping(){
	runner.exitFlag = true
}
func (runner *SimpleRunner)IsStopping() bool{
	return runner.exitFlag
}

func (runner *SimpleRunner)GetNotifyChannel() (chan bool){
	return runner.notifyChan
}

func (runner *SimpleRunner) NotifyExit(){
	runner.exitChan <- true
}
