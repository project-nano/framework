package framework

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type DaemonizedService interface {
	Start() (output string, err error)
	Stop() (output string, err error)
	Snapshot() (output string, err error)
}

type asyncOperateResult struct {
	Error  string `json:"error,omitempty"`
	Output string `json:"output,omitempty"`
}

type ConfigGenerator func(workingPath string) (err error)
type ServiceGenerator func(workingPath string) (service DaemonizedService, err error)

var (
	pipFileName       string
	daemonizedService DaemonizedService
)

func ProcessDaemon(executeName string, configGenerator ConfigGenerator, serviceGenerator ServiceGenerator) {
	const (
		ValidArguesCount = 2
		StartCommand     = "start"
		StopCommand      = "stop"
		StatusCommand    = "status"
		HaltCommand      = "halt"
		SnapshotCommand  = "snap"
		LogPathName      = "log"
	)
	if len(os.Args) != ValidArguesCount {
		printUsage(executeName)
		return
	}
	workingPath, err := getWorkingPath()
	if err != nil {
		fmt.Printf("get working path fail: %s\n", err.Error())
		return
	}
	var command = os.Args[1]
	var pidFileName = filepath.Join(workingPath, fmt.Sprintf("%s.pid", executeName))
	if "" == pipFileName{
		pipFileName = filepath.Join(workingPath, fmt.Sprintf("%s.pip", executeName))
	}
	context := &daemon.Context{
		PidFileName: pidFileName,
		PidFilePerm: daemon.FILE_PERM,
	}

	daemon.AddCommand(daemon.StringFlag(&command, StopCommand), syscall.SIGTERM, onStopDaemon)
	daemon.AddCommand(daemon.StringFlag(&command, SnapshotCommand), syscall.SIGUSR1, onDaemonSnapshot)
	daemon.AddCommand(daemon.StringFlag(&command, HaltCommand), syscall.SIGKILL, nil)

	switch command {
	case StopCommand:
		if process, err := context.Search(); err != nil {
			fmt.Printf("%s is already stopped\n", executeName)
			return
		} else if !isRunning(process) {
			fmt.Printf("%s is already stopped (process %d not running)\n", executeName, process.Pid)
		} else {
			defer os.Remove(pipFileName)
			if err = createPipe(pipFileName); err != nil{
				fmt.Printf("open pipe fail: %s\n", err.Error())
				return
			}
			go daemon.SendCommands(process)

			msg, err := readPipe(pipFileName)
			if err != nil {
				fmt.Printf("stop %s fail: %s\n", executeName, err.Error())
			} else if "" != msg {
				fmt.Println(msg)
				fmt.Printf("stop %s success\n", executeName)
			} else {
				fmt.Printf("stop %s success\n", executeName)
			}
		}
		return
	case SnapshotCommand:
		var process *os.Process
		if process, err = context.Search(); err != nil {
			fmt.Printf("%s is already stopped\n", executeName)
			return
		} else if !isRunning(process) {
			fmt.Printf("%s is already stopped (process %d not running)\n", executeName, process.Pid)
		} else {
			defer os.Remove(pipFileName)
			if err = createPipe(pipFileName); err != nil{
				fmt.Printf("open pipe fail: %s\n", err.Error())
				return
			}
			go daemon.SendCommands(process)
			var msg string
			msg, err = readPipe(pipFileName)
			if err != nil {
				fmt.Printf("capture snapshot of %s fail: %s\n", executeName, err.Error())
			} else if "" != msg {
				fmt.Println(msg)
				fmt.Printf("capture snapshot of %s success\n", executeName)
			} else {
				fmt.Printf("capture snapshot of %s success\n", executeName)
			}
		}
		return
	case HaltCommand:
		if process, err := context.Search(); err != nil {
			fmt.Printf("%s is already stopped\n", executeName)
		} else if !isRunning(process) {
			fmt.Printf("%s is already stopped (process %d not running)\n", executeName, process.Pid)
		} else {
			daemon.SendCommands(process)
		}
		return
	case StatusCommand:
		process, err := context.Search()
		if err != nil {
			fmt.Printf("%s is stopped\n", executeName)
		} else if !isRunning(process) {
			fmt.Printf("%s is stopped (pid %d)\n", executeName, process.Pid)
		} else {
			fmt.Printf("%s is running, current pid %d\n", executeName, process.Pid)
		}
		return
	case StartCommand:
		if err := configGenerator(workingPath); err != nil {
			fmt.Printf("generate config fail: %s\n", err.Error())
			return
		}
		if process, err := context.Search(); err == nil {
			if isRunning(process) {
				fmt.Printf("%s is already running\n", executeName)
				return
			}
		}
		if _, err = context.Reborn(); err != nil {
			fmt.Printf("create daemon fail:%s\n", err.Error())
			return
		}
		//parent or child
		if !daemon.WasReborn() {
			//parent
			msg, err := readMessageFromPipe(pipFileName)
			if err != nil {
				fmt.Printf("start %s fail: %s\n", executeName, err.Error())
			} else {
				fmt.Println(msg)
				fmt.Printf("%s started\n", executeName)
			}
			return

		} else {
			//child
			defer os.Remove(pidFileName)
			var logPath = filepath.Join(workingPath, LogPathName)
			if err = redirectLog(executeName, logPath); err != nil {
				notifyErrorToPipe(pipFileName, err.Error())
				log.Printf("redirect log fail: %s", err.Error())
				return
			}
			daemonizedService, err = serviceGenerator(workingPath)
			if err != nil {
				log.Printf("generate service fail: %s", err.Error())
				notifyErrorToPipe(pipFileName, err.Error())
				return
			}
			msg, err := daemonizedService.Start()
			if err != nil {
				log.Printf("start service fail: %s", err.Error())
				notifyErrorToPipe(pipFileName, err.Error())
			} else {
				notifyMessageToPipe(pipFileName, msg)
				daemon.ServeSignals()
			}
			return
		}
	default:
		printUsage(executeName)
	}
}

func onStopDaemon(sig os.Signal) error {
	if nil == daemonizedService{
		log.Println("invalid daemon service")
		return daemon.ErrStop
	}
	if "" == pipFileName{
		log.Println("invalid pipe file")
		return daemon.ErrStop
	}
	msg, err := daemonizedService.Stop()
	if err != nil{
		log.Printf("stop service fail: %s", err.Error())
		notifyErrorToPipe(pipFileName, err.Error())
	}else{
		notifyMessageToPipe(pipFileName, msg)
	}
	return daemon.ErrStop
}

func onDaemonSnapshot(sig os.Signal) error {
	if nil == daemonizedService{
		log.Println("invalid daemon service")
		return daemon.ErrStop
	}
	if "" == pipFileName{
		log.Println("invalid pipe file")
		return daemon.ErrStop
	}
	msg, err := daemonizedService.Snapshot()
	if err != nil{
		log.Printf("invoke snapshot fail: %s", err.Error())
		notifyErrorToPipe(pipFileName, err.Error())
	}else{
		notifyMessageToPipe(pipFileName, msg)
	}
	return nil
}

func readMessageFromPipe(pipeName string) (message string, err error) {
	defer os.Remove(pipeName)
	if err = createPipe(pipeName);err != nil{
		return
	}
	message, err = readPipe(pipeName)
	return
}

func createPipe(pipeName string) (err error) {
	const (
		PipeFilePerm = 0600
	)
	if _, err = os.Stat(pipeName); !os.IsNotExist(err){
		os.Remove(pipeName)
	}
	if err = syscall.Mkfifo(pipeName, PipeFilePerm);err != nil{
		return
	}

	return
}

func readPipe(pipeName string) (message string, err error) {
	const (
		PipeFilePerm = 0600
	)
	var pipe *os.File
	pipe, err = os.OpenFile(pipeName, os.O_RDONLY, PipeFilePerm)
	if err != nil{
		return
	}
	defer pipe.Close()

	var data = make([]byte, 1<<10)
	var n int
	n, err = pipe.Read(data)
	if err != nil {
		return
	}
	var result asyncOperateResult
	if err = json.Unmarshal(data[:n], &result); err != nil {

		return "", fmt.Errorf("unmarshal fail: %s, data %s", err.Error(), data[:n])
	}
	if result.Error != ""{
		err = errors.New(result.Error)
	}else{
		message = result.Output
	}
	return
}

func notifyMessageToPipe(pipeName, message string) (err error) {
	const (
		PipeFilePerm = 0600
	)
	pip, err := os.OpenFile(pipeName, os.O_RDWR, PipeFilePerm)
	if err != nil{
		return
	}
	defer pip.Close()
	var result = asyncOperateResult{Output:message}
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil{
		return
	}
	_, err = pip.Write(data)
	return
}

func notifyErrorToPipe(pipeName, message string) (err error) {
	const (
		PipeFilePerm = 0600
	)
	pip, err := os.OpenFile(pipeName, os.O_RDWR, PipeFilePerm)
	if err != nil{
		return
	}
	defer pip.Close()
	var result = asyncOperateResult{Error:message}
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil{
		return
	}
	_, err = pip.Write(data)
	return
}

func getWorkingPath() (path string, err error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Dir(executable))
}

func isRunning(process *os.Process) bool {
	if err := process.Signal(syscall.Signal(0)); err == nil {
		return true
	} else {
		return false
	}
}

func redirectLog(executeName, logPathName string) (err error) {
	const (
		DefaultLogPathPerm = 0740
	)
	if _, err = os.Stat(logPathName); os.IsNotExist(err) {
		//create path
		err = os.Mkdir(logPathName, DefaultLogPathPerm)
		if err != nil {
			return err
		}
		fmt.Printf("log path %s created\n", logPathName)
	}
	var filename = fmt.Sprintf("%s/%s.log", logPathName, executeName)
	output, err := os.Create(filename)
	if err != nil {
		return err
	}
	fmt.Printf("log redirected to '%s'\n", filename)
	log.SetOutput(output)
	return nil
}

func printUsage(executeName string) {
	fmt.Printf("Usage: %s [start|stop|status|halt|snap]\n", executeName)
}
