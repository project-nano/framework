package framework

import (
	"github.com/project-nano/sonar"
	"net"
	"fmt"
	"errors"
	"github.com/xtaci/kcp-go"
	"log"
	"time"
)

//need overwriting
type ServiceHandler interface {
	OnMessageReceived(msg Message)
	OnServiceConnected(name string, t ServiceType, address string)
	OnServiceDisconnected(name string, t ServiceType, gracefully bool)
	OnDependencyReady()
	InitialEndpoint() error
	OnEndpointStarted() error
	OnEndpointStopped()
}

type MessageSender interface {
	SendMessage(msg Message, target string) error
	SendToSelf(msg Message) error
}

type EndpointService struct {
	isPeer              bool
	isReady             bool
	groupListener       *sonar.Listener
	groupPinger         *sonar.Pinger
	fixedListenAddress  string
	connectionListener  *kcp.Listener
	connectionMap       map[string]connEntry
	connEventChan       chan connEvent
	incomingMessageChan chan Message
	guardianNotifyChan  chan bool
	guardianFinishChan  chan bool
	status              serviceStatus
	submoduleChannel    map[string]chan Message
	listenAddress       string
	listenPort          int
	name                string
	serviceType         ServiceType
	domain              string
	groupAddress        string
	groupPort           int
	stubAvailable       bool
	recoveringStub      bool
	handler             ServiceHandler
}

const (
	ListenPortRangeStart    = 5600
	ListenPortRangeEnd      = ListenPortRangeStart + 200
	DefaultDataShards       = 10
	DefaultParityShards     = 3
	DefaultBufferSize       = 2 << 20 //8KiB
	DefaultMessageQueueSize = 1 << 10
)

type serviceStatus int

const (
	serviceStatusStopped  = iota
	serviceStatusStopping
	serviceStatusRunning
)

func CreateStubEndpoint(groupAddress string, groupPort int, domain, listenAddress string) (endpoint EndpointService, err error) {
	listenInterface, err := getInterfaceByAddress(listenAddress)
	if err != nil{
		return
	}
	listener, err := sonar.CreateListener(groupAddress, groupPort, domain, listenInterface)
	if err != nil {
		return endpoint, err
	}
	return EndpointService{isPeer: false, groupListener: listener, fixedListenAddress: listenAddress,
		domain: domain, groupAddress: groupAddress, groupPort: groupPort,
		status: serviceStatusStopped, submoduleChannel:map[string]chan Message{}, stubAvailable: false, recoveringStub: false}, nil
}

func CreatePeerEndpoint(groupAddress string, groupPort int, domain string) (endpoint EndpointService, err error) {
	pinger, err := sonar.CreatePinger(groupAddress, groupPort, domain)
	if err != nil {
		return endpoint, err
	}
	return EndpointService{isPeer: true, groupPinger: pinger, status: serviceStatusStopped, submoduleChannel:map[string]chan Message{},
		domain: domain, groupAddress: groupAddress, groupPort: groupPort, stubAvailable: false, recoveringStub: false}, nil
}

func (endpoint *EndpointService)RegisterSubmodule(name string, channel chan Message) error{
	if _, exists := endpoint.submoduleChannel[name];exists{
		return fmt.Errorf("submodule '%s' already exists", name)
	}
	endpoint.submoduleChannel[name] = channel
	return nil
}

func (endpoint *EndpointService) RegisterHandler(h ServiceHandler){
	endpoint.handler = h
}

func getInterfaceByAddress(address string) (i *net.Interface, err error){
	list, err := net.Interfaces()
	if err != nil{
		return
	}
	if 0 == len(list){
		err = errors.New("no interface available")
		return
	}
	for _, inf := range list{
		if addreses, err := inf.Addrs();err == nil{
			for _, target := range addreses{
				if ip, _, err := net.ParseCIDR(target.String());err == nil{
					if ip.String() == address{
						i = &inf
						return i, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("no interface has address '%s'", address)
}
//
//func selectDefaultInterface() (i net.Interface, err error) {
//	interfaceList, err := net.Interfaces()
//	if err != nil {
//		return i, err
//	}
//	var checkFlag = net.FlagMulticast | net.FlagPointToPoint | net.FlagUp
//	for _, i = range interfaceList {
//		if i.Flags&net.FlagLoopback != 0 {
//			//ignore loopback
//			continue
//		}
//		if i.Flags&checkFlag != 0 {
//			return i, nil
//		}
//	}
//	return i, errors.New("no interface available")
//}

func InterfaceByAddress(address string) (inf *net.Interface, err error){
	interfaceList, err := net.Interfaces()
	if err != nil {
		return
	}
	var checkFlag = net.FlagMulticast | net.FlagPointToPoint | net.FlagUp
	var addressList []net.Addr
	var interfaceIP net.IP
	for _, currentInterface := range interfaceList {
		if currentInterface.Flags&net.FlagLoopback != 0 {
			//ignore
			continue
		}
		if currentInterface.Flags&checkFlag != 0 {
			addressList, err = currentInterface.Addrs()
			if err != nil {
				return
			}
			if len(addressList) == 0 {
				//no address available
				continue
			}

			for _, addr := range addressList {
				interfaceIP, _, err = net.ParseCIDR(addr.String())
				if err != nil {
					return
				}
				if interfaceIP.To4() != nil {
					//v4
					if interfaceIP.String() == address{
						//equal
						return &currentInterface, nil
					}
				}
			}
		}
	}
	err = fmt.Errorf("no interface with address '%s'", address)
	return
}

func (endpoint *EndpointService) GenerateName(t ServiceType, i *net.Interface) error {
	const hexDigit = "0123456789abcdef"
	var prefix string
	switch t {
	case ServiceTypeCore:
		prefix = "Core"
		break
	case ServiceTypeCell:
		prefix = "Cell"
		break
	case ServiceTypeImage:
		prefix = "Image"
		break
	default:
		return fmt.Errorf("unsupported service type %d", t)
	}
	var buf []byte
	for _, b := range i.HardwareAddr {
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}
	endpoint.serviceType = t
	endpoint.name = fmt.Sprintf("%s_%s", prefix, string(buf))
	return nil

}

func (endpoint *EndpointService) isRunning() bool {
	return endpoint.status == serviceStatusRunning
}

func (endpoint *EndpointService) isStopped() bool {
	return endpoint.status == serviceStatusStopped
}

func (endpoint *EndpointService) Start() error {
	if !endpoint.isStopped() {
		return errors.New("endpoint not stopped")
	}
	if err := endpoint.handler.InitialEndpoint(); err != nil {
		return err
	}
	var err error
	if endpoint.isPeer {
		err = endpoint.startPeerService()
	} else {
		err = endpoint.startCoreService()
	}
	if err != nil {
		return err
	}
	endpoint.status = serviceStatusRunning
	return nil
}

func (endpoint *EndpointService) Stop() error {
	if !endpoint.isRunning() {
		return errors.New("endpoint not running")
	}
	endpoint.status = serviceStatusStopping
	endpoint.handler.OnEndpointStopped()
	if err := endpoint.connectionListener.Close(); err != nil {
		endpoint.status = serviceStatusStopped
		return err
	}
	endpoint.guardianNotifyChan <- true
	<-endpoint.guardianFinishChan
	close(endpoint.connEventChan)
	close(endpoint.incomingMessageChan)
	endpoint.status = serviceStatusStopped
	return nil
}

//all dependent service must ready before prepare local service
func (endpoint *EndpointService) AddDependency(dependencies []ServiceType) {
	panic("not implement")
}

func (endpoint *EndpointService) SetServiceReady() {
	panic("not implement")
}

func (endpoint *EndpointService) GetListenAddress() string{
	return endpoint.listenAddress
}

func (endpoint *EndpointService) GetListenPort() int{
	return endpoint.listenPort
}

func (endpoint *EndpointService) GetName() string{
	return endpoint.name
}

func (endpoint *EndpointService) GetDomain() string{
	return endpoint.domain
}

func (endpoint *EndpointService) GetServiceType() ServiceType{
	return endpoint.serviceType
}

func (endpoint *EndpointService) GetGroupAddress() string{
	return endpoint.groupAddress
}

func (endpoint *EndpointService) GetGroupPort() int{
	return endpoint.groupPort
}
//private functions
func (endpoint *EndpointService) startCoreService() error {
	listener, listenPort, err := selectAvailablePort(endpoint.fixedListenAddress)
	if err != nil {
		return err
	}
	if err = endpoint.groupListener.AddService(ServiceTypeStringCore, "kcp", endpoint.fixedListenAddress, listenPort); err != nil {
		return err
	}
	log.Printf("<endpoint> service %s published for %s:%d", endpoint.name, endpoint.fixedListenAddress, listenPort)
	if err = endpoint.groupListener.Start(); err != nil {
		return err
	}
	endpoint.listenAddress = endpoint.fixedListenAddress
	endpoint.listenPort = listenPort
	return endpoint.startRoutine(listener)
}

func (endpoint *EndpointService) startPeerService() error {
	const (
		DefaultQueryDuration = 5*time.Second
	)
	echo, err := endpoint.groupPinger.Query(DefaultQueryDuration)
	if err != nil {
		return err
	}
	//select first service
	for _, service := range echo.Services {
		if ServiceTypeStringCore != service.Type {
			log.Printf("<endpoint> warning:invalid service type `%s` in echo response", service.Type)
			continue
		}
		//create listener
		listener, listenPort, err := selectAvailablePort(echo.LocalAddress)
		if err != nil {
			return err
		}
		//start routine
		log.Printf("<endpoint> %s listen at %s:%d", endpoint.name, echo.LocalAddress, listenPort)
		endpoint.listenPort = listenPort
		endpoint.listenAddress = echo.LocalAddress
		if err = endpoint.startRoutine(listener); err != nil {
			return err
		}
		//connect service
		return endpoint.connectRemoteService(service.Address, service.Port)
	}
	return errors.New("no service available")
}

func (endpoint *EndpointService) SendMessage(msg Message, target string) error {
	if !endpoint.isRunning() {
		return errors.New("endpoint closed")
	}
	if target == endpoint.name{
		return endpoint.SendToSelf(msg)
	}
	//inner submodule first
	channel, exists := endpoint.submoduleChannel[target]
	if exists{
		channel <- msg
		return nil
	}
	entry, exists := endpoint.connectionMap[target]
	if !exists {
		return fmt.Errorf("invalid target '%s'", target)
	}
	entry.OutgoingChan <- msg
	return nil
}

func (endpoint *EndpointService) SendToSelf(msg Message) error {
	if "" == msg.GetSender(){
		msg.SetSender(endpoint.name)
	}
	endpoint.incomingMessageChan <- msg
	return nil
}

func (endpoint *EndpointService) startRoutine(listener *kcp.Listener) error {
	endpoint.connectionListener = listener
	endpoint.connEventChan = make(chan connEvent, DefaultMessageQueueSize)
	endpoint.incomingMessageChan = make(chan Message, DefaultMessageQueueSize)
	endpoint.guardianNotifyChan = make(chan bool, 1)
	endpoint.guardianFinishChan = make(chan bool, 1)
	endpoint.connectionMap = map[string]connEntry{}
	go endpoint.listenRoutine()
	go endpoint.guardianRoutine()
	go endpoint.mainRoutine()
	return endpoint.handler.OnEndpointStarted()
}

func (endpoint *EndpointService) listenRoutine() {
	//listen&accept
	for {
		session, err := endpoint.connectionListener.AcceptKCP()
		if err != nil {
			break
		}
		go endpoint.handleIncomingConnection(session)
	}
}

type connEvent struct {
	Event        connEventType
	Name         string
	Service      ServiceType
	Address      string
	Port         int
	Gracefully   bool
	Conn         *kcp.UDPSession
	OutgoingChan chan Message
	FinishChan   chan bool
}

type connEntry struct {
	Name          string //remote name
	Type          ServiceType
	Status        connectionStatus
	LastHeartBeat time.Time
	Session       *kcp.UDPSession
	OutgoingChan  chan Message
	FinishChan    chan bool
}

type connEventType int

const (
	ConnEventOpen      = iota
	ConnEventClose
	ConnEventHeartBeat
)

type connectionStatus int

//Service status
const (
	connStatusDisconnected = iota
	connStatusConnected
	connStatusReady
	connStatusLost
)

func (endpoint *EndpointService) guardianRoutine() {
	var exitFlag = false
	const (
		CheckInterval               = time.Second * 5
		KeepAliveInterval           = time.Second * 3
		LostThresholdInterval       = time.Second * 9
		DisconnectThresholdInterval = time.Second * 15
	)
	var checkTicker = time.NewTicker(CheckInterval)
	var keepAliveTicker = time.NewTicker(KeepAliveInterval)

	for !exitFlag {
		select {
		case <-endpoint.guardianNotifyChan:
			//log.Print("exit guardian routine...")
			exitFlag = true
			break
		case event := <-endpoint.connEventChan:
			switch event.Event {
			case ConnEventOpen:
				{
					if _, exists := endpoint.connectionMap[event.Name]; exists {
						log.Printf("<endpoint> connection to service '%s' already opened", event.Name)
						continue
					}
					endpoint.connectionMap[event.Name] = connEntry{event.Name, event.Service, connStatusConnected,
						time.Now(), event.Conn, event.OutgoingChan, event.FinishChan}
					log.Printf("<endpoint> new connection '%s' opened", event.Name)
					if !endpoint.stubAvailable && (ServiceTypeCore == event.Service){
						endpoint.stubAvailable = true
					}
					msg, err := CreateJsonMessage(ServiceConnectedEvent)
					if err != nil {
						log.Printf("<endpoint> create message fail:%s", err.Error())
						continue
					}
					msg.SetString(ParamKeyName, event.Name)
					msg.SetUInt(ParamKeyType, uint(event.Service))
					msg.SetString(ParamKeyAddress, event.Address)
					if err = endpoint.SendToSelf(msg); err != nil {
						log.Printf("<endpoint> notify connected event fail: %s", err.Error())
						continue
					}

				}
			case ConnEventClose:
				entry, exists := endpoint.connectionMap[event.Name]
				if !exists {
					log.Printf("<endpoint> service '%s' not exists", event.Name)
					continue
				}
				var serviceType = entry.Type
				delete(endpoint.connectionMap, event.Name)
				log.Printf("<endpoint> connection '%s' closed", event.Name)
				if endpoint.isRunning()&&(ServiceTypeCore == serviceType) && endpoint.isPeer {
					//todo: verify multiple stub
					endpoint.stubAvailable = false
					go endpoint.recoverStubService()
				}
				msg, err := CreateJsonMessage(ServiceDisconnectedEvent)
				if err != nil {
					log.Printf("<endpoint> create message fail:%s", err.Error())
					continue
				}
				msg.SetString(ParamKeyName, event.Name)
				msg.SetUInt(ParamKeyType, uint(serviceType))
				msg.SetBoolean(ParamKeyFlag, event.Gracefully)
				if err = endpoint.SendToSelf(msg); err != nil {
					log.Printf("<endpoint> notify disconnected event fail: %s", err.Error())
					continue
				}

			case ConnEventHeartBeat:
				entry, exists := endpoint.connectionMap[event.Name]
				if !exists {
					log.Printf("<endpoint> invalid service '%s' for heartbeat", event.Name)
					continue
				}
				entry.LastHeartBeat = time.Now()
				entry.Status = connStatusConnected
				endpoint.connectionMap[event.Name] = entry

			default:
				log.Printf("<endpoint> warning: invalid connection event type %d", event.Event)
			}
			break
			//keep alive
		case <-keepAliveTicker.C:
			keepAlive, err := CreateJsonMessage(ConnectionKeepAliveEvent)
			if err != nil {
				log.Printf("<endpoint> warning: build keep alive message fail: %s", err.Error())
				break
			}
			for name, entry := range endpoint.connectionMap {
				if entry.Status == connStatusConnected {
					//only send keep alive to connected serivce
					if err = endpoint.SendMessage(keepAlive, name); err != nil {
						log.Printf("<endpoint> warning: send keep alive to '%s' fail: %s", name, err.Error())
					}
				}
			}
			break
			//check timeout
		case <-checkTicker.C:
			var current = time.Now()
			for name, entry := range endpoint.connectionMap {
				if connStatusConnected == entry.Status {
					if entry.LastHeartBeat.Add(LostThresholdInterval).Before(current) {
						//timeout
						entry.Status = connStatusLost
						endpoint.connectionMap[name] = entry
						log.Printf("<endpoint> service '%s' marked to lost", name)
					}
				} else if connStatusLost == entry.Status {
					if entry.LastHeartBeat.Add(DisconnectThresholdInterval).Before(current) {
						//timeout
						entry.Status = connStatusDisconnected
						endpoint.connectionMap[name] = entry
						log.Printf("<endpoint> service '%s' marked to disconnect", name)
						if err := endpoint.disconnectRemoteService(name, entry); err != nil {
							log.Printf("<endpoint> try disconnect lost service '%s' fail: %s", name, err.Error())
						}
					}
				}
			}
			break
		}

	}
	checkTicker.Stop()
	keepAliveTicker.Stop()
	for name, entry := range endpoint.connectionMap {
		if err := endpoint.disconnectRemoteService(name, entry);err != nil{
			log.Printf("<endpoint> disconnect service '%s' fail when stop: %s", name, err.Error())
		}
	}
	endpoint.guardianFinishChan <- true

}

func (endpoint *EndpointService) mainRoutine() {
	//handle incoming message
	for msg := range endpoint.incomingMessageChan {
		if !endpoint.isRunning() {
			break
		}
		switch msg.GetID() {
		case ServiceAvailableEvent, ServiceReadyEvent, ServiceConnectedEvent, ServiceDisconnectedEvent:
			endpoint.handleSystemMessage(msg)
		default:
			endpoint.handler.OnMessageReceived(msg)
		}
	}
}


func (endpoint *EndpointService) handleSystemMessage(msg Message) {
	switch msg.GetID() {
	case ServiceConnectedEvent:
		serviceName, err := msg.GetString(ParamKeyName)
		if err != nil {
			log.Printf("<endpoint> get name fail:%s", err.Error())
			return
		}
		serviceType, err := msg.GetUInt(ParamKeyType)
		if err != nil {
			log.Printf("<endpoint> get type fail:%s", err.Error())
			return
		}
		remoteAddress, err := msg.GetString(ParamKeyAddress)
		if err != nil{
			log.Printf("<endpoint> get remote address fail: %s", err.Error())
			return
		}
		endpoint.handler.OnServiceConnected(serviceName, ServiceType(serviceType), remoteAddress)
		return
	case ServiceDisconnectedEvent:
		serviceName, err := msg.GetString(ParamKeyName)
		if err != nil {
			log.Printf("<endpoint> get name fail:%s", err.Error())
			return
		}
		serviceType, err := msg.GetUInt(ParamKeyType)
		if err != nil {
			log.Printf("<endpoint> get type fail:%s", err.Error())
			return
		}
		gracefully, _ := msg.GetBoolean(ParamKeyFlag)
		endpoint.handler.OnServiceDisconnected(serviceName, ServiceType(serviceType), gracefully)
		return
	}
}

func (endpoint *EndpointService) handleIncomingConnection(session *kcp.UDPSession) {
	//receiver
	//read remote service info
	//send local service info
	var remoteAddress = session.RemoteAddr().(*net.UDPAddr)
	serviceName, serviceType, err := receiveRemoteServiceInfo(session)
	if err != nil {
		session.Close()
		log.Printf("<endpoint> get service info fail:%s", err.Error())
		return
	}
	var outgoingChan = make(chan Message, DefaultMessageQueueSize)
	var finishChan = make(chan bool, 1)
	var remoteIP = remoteAddress.IP.String()
	log.Printf("<endpoint> new service '%s' (type %d) connected from %s:%d", serviceName, serviceType, remoteIP, remoteAddress.Port)
	endpoint.connEventChan <- connEvent{ConnEventOpen, serviceName, serviceType,
		remoteIP, remoteAddress.Port, false, session, outgoingChan, finishChan}
	//notify remote service
	if err = sendServiceInfo(session, endpoint.name, endpoint.serviceType); err != nil {
		session.Close()
		log.Printf("<endpoint> notify service info fail:%s", err.Error())
		return
	}
	//start routine
	go sessionServeRoutine(serviceName, session, endpoint.incomingMessageChan, outgoingChan, finishChan, endpoint.connEventChan)
}

func (endpoint *EndpointService) connectRemoteService(address string, port int) error {
	//sender:
	//send local service info
	//read remote service info
	var target = fmt.Sprintf("%s:%d", address, port)
	session, err := kcp.DialWithOptions(target, nil, DefaultDataShards, DefaultParityShards)
	if err != nil {
		return err
	}
	if err = sendServiceInfo(session, endpoint.name, endpoint.serviceType); err != nil {
		session.Close()
		return err
	}
	remoteName, remoteType, err := receiveRemoteServiceInfo(session)
	if err != nil {
		session.Close()
		return err
	}

	var outgoingChan = make(chan Message, DefaultMessageQueueSize)
	var finishChan = make(chan bool,1 )
	log.Printf("<endpoint> remote service '%s' (type %d/ address %s) connected", remoteName, remoteType, target)
	endpoint.connEventChan <- connEvent{ConnEventOpen, remoteName, remoteType,
		address, port, true, session, outgoingChan, finishChan}
	//start routine
	go sessionServeRoutine(remoteName, session, endpoint.incomingMessageChan, outgoingChan, finishChan, endpoint.connEventChan)
	return nil
}

func (endpoint *EndpointService) disconnectRemoteService(name string, entry connEntry) (err error) {
	event, err := CreateJsonMessage(ConnectionClosedEvent)
	if err != nil{
		return err
	}
	data, err := event.Serialize()
	if err != nil{
		return err
	}
	//send disconnect event
	if _, err = entry.Session.Write(data);err != nil{
		return err
	}
	if err = entry.Session.Close(); err != nil {
		return err
	}
	const (
		stopTimeout = 3*time.Second
	)
	timer := time.NewTimer(stopTimeout)
	select {
	case <- timer.C:
		err = errors.New("wait session routine finish timeout")
		return err
	case <- entry.FinishChan:
		//finished
	}
	return nil
}

func (endpoint *EndpointService) recoverStubService(){
	if endpoint.recoveringStub{
		log.Println("<endpoint> recovery already in processing")
		return
	}
	endpoint.recoveringStub = true
	const (
		retryInterval = 3*time.Second
		queryTimeout = 5*time.Second
	)
	defer func() {endpoint.recoveringStub = false}()
	var err error
	var echo sonar.Echo
	if err != nil{
		log.Printf("<endpoint> create recover pinger fail: %s", err.Error())
		return
	}
	for endpoint.isRunning(){
		time.Sleep(retryInterval)
		if endpoint.stubAvailable{
			log.Println("<endpoint> stub service already recovered")
			break
		}
		log.Println("<endpoint> try recover stub service...")
		pinger, err := sonar.CreatePinger(endpoint.groupAddress, endpoint.groupPort, endpoint.domain)
		echo, err = pinger.Query(queryTimeout)
		if err != nil{
			log.Printf("<endpoint> recover fail: %s", err.Error())
			continue
		}
		if 0 == len(echo.Services){
			log.Println("<endpoint> requery success, but no stub available")
			continue
		}
		var stub = echo.Services[0]
		err = endpoint.connectRemoteService(stub.Address, stub.Port)
		if err != nil{
			log.Printf("<endpoint> connect stub %s:%d fail: %s", stub.Address, stub.Port, err.Error())
			continue
		}
		log.Printf("<endpoint> new stub %s:%d recovered", stub.Address, stub.Port)
		break
	}

}

func selectAvailablePort(host string) (*kcp.Listener, int, error) {
	var address string
	for port := ListenPortRangeStart; port < ListenPortRangeEnd; port++ {
		address = fmt.Sprintf("%s:%d", host, port)
		listener, err := kcp.ListenWithOptions(address, nil, DefaultDataShards, DefaultParityShards)
		if err != nil {
			continue
		}
		return listener, port, nil
	}
	return nil, 0, fmt.Errorf("no port available in range %d ~ %d", ListenPortRangeStart, ListenPortRangeEnd)
}

func receiveRemoteServiceInfo(session *kcp.UDPSession) (string, ServiceType, error) {
	var buf = make([]byte, DefaultBufferSize)
	//recv connect open
	length, err := session.Read(buf)
	if err != nil {
		return "", 0, err
	}
	msg, err := MessageFromJson(buf[:length])
	if err != nil {
		return "", 0, err
	}
	if msg.GetID() != ConnectionOpenedEvent {
		return "", 0, fmt.Errorf("invalid message %d", msg.GetID())
	}
	serviceName, err := msg.GetString(ParamKeyName)
	if err != nil {
		return "", 0, errors.New("can not get service name")
	}
	serviceType, err := msg.GetUInt(ParamKeyType)
	if err != nil {
		return "", 0, errors.New("can not get service type")
	}
	return serviceName, ServiceType(serviceType), nil
}

func sendServiceInfo(session *kcp.UDPSession, serviceName string, serviceType ServiceType) error {
	notify, err := CreateJsonMessage(ConnectionOpenedEvent)
	if err != nil {
		return err
	}
	notify.SetString(ParamKeyName, serviceName)
	notify.SetUInt(ParamKeyType, uint(serviceType))
	packet, err := notify.Serialize()
	if err != nil {
		return err
	}
	_, err = session.Write(packet)
	return err
}

func sessionServeRoutine(remote string, session *kcp.UDPSession, incomingChan chan Message,
	outgoingChan chan Message, finishChan chan bool, eventChan chan connEvent) {
	//log.Printf("<endpoint> receive routine for '%s' started", remote)
	var gracefullyClose = false
	var buf = make([]byte, DefaultBufferSize)
	var sendStopChan = make(chan bool, 1)
	var sendExitChan = make(chan bool, 1)
	go sessionOutgoingRoutine(remote, session, outgoingChan, sendStopChan, sendExitChan)
	var bufStart, bufEnd = 0, 0
	for {
		//recv connect open
		count, err := session.Read(buf[bufStart:])
		if err != nil {
			log.Printf("<endpoint> warning: connection lost from %s : %s", remote, err.Error())
			break
		}
		bufEnd = bufStart + count
		if bufEnd > DefaultBufferSize{
			bufStart = 0
			log.Printf("<endpoint> warning: discard cached data because buffer overflow from %s", remote)
			continue
		}
		msg, err := MessageFromJson(buf[:bufEnd])
		if err != nil {
			//need cache
			bufStart = bufEnd
			log.Printf("<endpoint> warning: cache %d byte(s) from %s", count, remote)
			//log.Printf("<endpoint> warning: parse message fail: %s, data(%d byte(s)): %s", err.Error(), count, buf[:count])
			continue
		}
		bufStart = 0
		if msg.GetID() == ConnectionKeepAliveEvent {
			eventChan <- connEvent{Event: ConnEventHeartBeat, Name: remote}
			continue
		}else if msg.GetID() == ConnectionClosedEvent{
			gracefullyClose = true
			log.Printf("<endpoint> connection closed by remote endpoint '%s'", remote)
			break
		}
		if "" == msg.GetSender() {
			msg.SetSender(remote)
		}
		incomingChan <- msg
	}
	//closing outgoing routine
	sendStopChan <- true
	<-sendExitChan
	//notify closed
	eventChan <- connEvent{Event: ConnEventClose, Name: remote, Gracefully: gracefullyClose}
	finishChan <- true
	//log.Printf("<endpoint> receive routine for '%s' stopped", remote)
}

func sessionOutgoingRoutine(remote string, session *kcp.UDPSession, outgoingChan chan Message,
	notify, stopped chan bool) {
	//log.Printf("<endpoint> send routine for '%s' started", remote)
	var exitFlag = false
	for !exitFlag {
		select {
		case msg := <-outgoingChan:
			data, err := msg.Serialize()
			if err != nil {
				log.Printf("<endpoint> serial outgoing message fail: %s", err.Error())
				break
			}
			if _, err = session.Write(data); err != nil {
				log.Printf("<endpoint> outgoing message to '%s' fail: %s", remote, err.Error())
				break
			}
			//log.Printf("debug:message send to '%s'", remote)
		case <-notify:
			exitFlag = true
		}
	}
	stopped <- true
	//log.Printf("<endpoint> send routine for '%s' stopped", remote)
}
