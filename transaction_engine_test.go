package framework

import (
	"time"
	"testing"
	"fmt"
)

const (
	KeyTaskID = 1
)

var resultChan = make(chan uint, 1 << 12)

type WaitExecutor struct {
	t *testing.T
}

func (executor *WaitExecutor)Execute(id SessionID, message Message, incoming chan Message, terminate chan bool) error{
	taskID, _ := message.GetUInt(KeyTaskID)
	//executor.t.Logf("[%08X] task %d begin wait", id, taskID)
	time.Sleep(1000 * time.Millisecond)
	//executor.t.Logf("[%08X] task %d wait finish", id, taskID)
	resultChan <- taskID
	return nil
}

func TestTransactionEngine_WaitExecutor(t *testing.T) {
	engine, err := CreateTransactionEngine()
	if err != nil{
		t.Fatalf("create engine fail: %s", err.Error())
	}
	const (
		testMessage = 1
		id1 = 1
		id2 = 2
	)
	engine.RegisterExecutor(testMessage, &WaitExecutor{t})

	if err = engine.Start(); err != nil{
		t.Fatalf("start engine fail: %s", err.Error())
	}
	t.Log("engine started")

	task1, _ := CreateJsonMessage(testMessage)
	task1.SetUInt(KeyTaskID, id1)

	task2, _ := CreateJsonMessage(testMessage)
	task2.SetUInt(KeyTaskID, id2)

	engine.InvokeTask(task1)
	time.Sleep(50*time.Millisecond)
	engine.InvokeTask(task2)

	var expected = []uint{id1, id2}
	var executed = make([]uint, 0)
	var timer = time.NewTimer(5*time.Second)
	for len(executed) < 2{
		select {
		case <- timer.C:
			t.Fatal("execute timeout")
		case id := <-resultChan:
			executed = append(executed, id)
		}
	}
	if len(executed) != len(expected){
		t.Fatal("unexpect queue length")
	}
	for i := 0; i < len(expected); i++{
		if expected[i] != executed[i]{
			t.Fatalf("unexpect task %d at index %d", executed[i], i)
		}
	}
	if err = engine.Stop(); err != nil{
		t.Fatalf("stop engine fail: %s", err.Error())
	}
	t.Log("engine stopped")
}

type TimeoutExecutor struct {
	Timeout time.Duration
}

func (executor *TimeoutExecutor)Execute(id SessionID, message Message, incoming chan Message, terminate chan bool) error{
	taskID, _ := message.GetUInt(KeyTaskID)
	var timer = time.NewTimer(executor.Timeout)
	select {
		case <- timer.C:
			resultChan <- taskID
	}
	return nil
}

func TestTransactionEngine_SessionTimeout(t *testing.T) {
	engine, err := CreateTransactionEngine()
	if err != nil{
		t.Fatalf("create engine fail: %s", err.Error())
	}
	const (
		testMessage = 2
		timeout = 1000*time.Millisecond
		taskCount = 5
	)
	engine.RegisterExecutor(testMessage, &TimeoutExecutor{Timeout:timeout})

	if err = engine.Start(); err != nil{
		t.Fatalf("start engine fail: %s", err.Error())
	}
	var expected []uint
	t.Log("engine started")
	for taskID := 0; taskID < taskCount; taskID++{
		msg, _ :=  CreateJsonMessage(testMessage)
		msg.SetUInt(KeyTaskID, uint(taskID))
		expected = append(expected, uint(taskID))
		engine.InvokeTask(msg)
		time.Sleep(50*time.Millisecond)
	}

	var executed = make([]uint, 0)
	var timer = time.NewTimer(6*time.Second)
	for len(executed) < taskCount{
		select {
		case <- timer.C:
			t.Fatal("execute timeout")
		case id := <-resultChan:
			executed = append(executed, id)
		}
	}
	if len(executed) != len(expected){
		t.Fatal("unexpect queue length")
	}
	for i := 0; i < len(expected); i++{
		if expected[i] != executed[i]{
			t.Fatalf("unexpect task %d at index %d", executed[i], i)
		}
	}
	if err = engine.Stop(); err != nil{
		t.Fatalf("stop engine fail: %s", err.Error())
	}
	t.Log("engine stopped")
}


type NopeExecutor struct {

}

func (executor *NopeExecutor)Execute(id SessionID, message Message, incoming chan Message, terminate chan bool) error{
	taskID, _ := message.GetUInt(KeyTaskID)
	resultChan <- taskID
	return nil
}

func TestTransactionEngine_NopeExecutor(t *testing.T) {
	engine, err := CreateTransactionEngine()
	if err != nil{
		t.Fatalf("create engine fail: %s", err.Error())
	}
	const (
		testMessage = 3
		taskCount = 1 << 11
	)
	engine.RegisterExecutor(testMessage, &NopeExecutor{})

	if err = engine.Start(); err != nil{
		t.Fatalf("start engine fail: %s", err.Error())
	}
	var expected []uint
	t.Log("engine started")
	for taskID := 0; taskID < taskCount; taskID++{
		msg, _ :=  CreateJsonMessage(testMessage)
		msg.SetUInt(KeyTaskID, uint(taskID))
		expected = append(expected, uint(taskID))
		engine.InvokeTask(msg)
		time.Sleep(5*time.Millisecond)
	}

	var executed = make([]uint, 0)
	var timer = time.NewTimer(6*time.Second)
	for len(executed) < taskCount{
		select {
		case <- timer.C:
			t.Fatal("execute timeout")
		case id := <-resultChan:
			executed = append(executed, id)
		}
	}
	if len(executed) != len(expected){
		t.Fatal("unexpect queue length")
	}
	for i := 0; i < len(expected); i++{
		if expected[i] != executed[i]{
			t.Fatalf("unexpect task %d at index %d", executed[i], i)
		}
	}
	if err = engine.Stop(); err != nil{
		t.Fatalf("stop engine fail: %s", err.Error())
	}
	t.Logf("engine stopped, %d task executed", len(executed))
}

type OneStepExecutor struct {

}

const (
	FirstMessage = iota
	SecondMessage
)

func (executor *OneStepExecutor)Execute(id SessionID, message Message, incoming chan Message, terminate chan bool) error{
	taskID, _ := message.GetUInt(KeyTaskID)
	var timer = time.NewTimer(5*time.Second)
	select {
	case received := <- incoming:
		if received.GetID() != FirstMessage{
			return fmt.Errorf("unexpected message %d", received.GetID())
		}
		resultChan <- taskID
	case <- timer.C:
		return fmt.Errorf("wait message for task %d timeout", taskID)
	}
	return nil
}