package framework

import (
	"testing"
	"time"
)

func Test_FunctionRunner(t *testing.T)  {

	var r = func(c RoutineController) {
		t.Log("routine started")
		for !c.IsStopping() {
			select {
			case <- c.GetNotifyChannel():
				c.SetStopping()
			}
		}
		t.Log("routine stoped")
		c.NotifyExit()
	}

	var runner = CreateSimpleRunner(r)
	var err = runner.Start()
	if err != nil{
		t.Fatalf("start runner fail: %s", err.Error())
	}
	var exitChan = make(chan bool, 1)
	go func() {
		runner.Stop()
		exitChan <- true
	}()
	var timer = time.NewTimer(3*time.Second)
	select{
	case <- timer.C:
		t.Fatal("stop runner timeout")
	case <- exitChan:
		t.Log("runner stopped")
	}
}

type stub struct {
	t *testing.T
}

func (s *stub)SomeMember(c RoutineController){
	s.t.Log("routine started")
	for !c.IsStopping() {
		select {
		case <- c.GetNotifyChannel():
			c.SetStopping()
		}
	}
	s.t.Log("routine stoped")
	c.NotifyExit()
}

func Test_MemberRunner(t *testing.T)  {
	var s = stub{t}

	var runner = CreateSimpleRunner(s.SomeMember)
	var err = runner.Start()
	if err != nil{
		t.Fatalf("start runner fail: %s", err.Error())
	}
	if err = runner.Start(); nil == err{
		t.Fatal("duplicate start success")
	}
	t.Logf("duplicate start fail: %s", err.Error())
	if !runner.IsRunning(){
		t.Fatal("runner not in running status")
	}
	t.Log("runner in running status")
	var exitChan = make(chan bool, 1)
	go func() {
		runner.Stop()
		exitChan <- true
	}()
	var timer = time.NewTimer(3*time.Second)
	select{
	case <- timer.C:
		t.Fatal("stop runner timeout")
	case <- exitChan:
		t.Log("runner stopped")
		if runner.IsRunning(){
			t.Fatal("runner in running status")
		}
		t.Log("runner not in running status")
		if err = runner.Stop(); nil == err{
			t.Fatal("duplicate stop success")
		}
		t.Logf("duplicate stop fail: %s", err.Error())
	}
	t.Log("member runner test: ok")
}
