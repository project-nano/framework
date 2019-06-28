package framework

import (
	"testing"
	"github.com/project-nano/sonar"
	"time"
	"net"
	"errors"
	"log"
)

type CoreEndpoint struct {
	EndpointService //base class
}

func (core *CoreEndpoint)OnMessageReceived(msg Message){
	log.Printf("<core> received message %d from %s", msg.GetID(), msg.GetSender())
}

func (core *CoreEndpoint)OnServiceConnected(name string, t ServiceType, remote string){
	log.Printf("<core> service %s connected, type %d", name, t)
}

func (core *CoreEndpoint)OnServiceDisconnected(name string, t ServiceType, gracefully bool){
	log.Printf("<core> service %s disconnected, type %d, gracefully: %t", name, t, gracefully)
}

func (core *CoreEndpoint)OnDependencyReady(){
	core.SetServiceReady()
}

func (core *CoreEndpoint)InitialEndpoint() error{
	log.Print("<core> initialed")
	return nil
}
func (core *CoreEndpoint)OnEndpointStarted() error{
	log.Print("<core> started")
	return nil
}
func (core *CoreEndpoint)OnEndpointStopped(){
	log.Print("<core> stopped")
}

type PeerEndpoint struct {
	EndpointService //base class
	EventChan chan bool
}

func (peer *PeerEndpoint)OnMessageReceived(msg Message){
	log.Printf("<peer> received message %d from %s", msg.GetID(), msg.GetSender())
}

func (peer *PeerEndpoint)OnServiceConnected(name string, t ServiceType, remote string){
	log.Printf("<peer> service %s connected, type %d", name, t)
	peer.EventChan <- true
}

func (peer *PeerEndpoint)OnServiceDisconnected(name string, t ServiceType, gracefully bool){
	log.Printf("<peer> service %s disconnected, type %d, gracefully %t", name, t, gracefully)
	peer.EventChan <- true
}

func (peer *PeerEndpoint)OnDependencyReady(){
	peer.SetServiceReady()
}
func (peer *PeerEndpoint)InitialEndpoint() error{
	log.Print("<peer> initialed")
	return nil
}
func (peer *PeerEndpoint)OnEndpointStarted() error{
	log.Print("<peer> started")
	return nil
}
func (peer *PeerEndpoint)OnEndpointStopped(){
	log.Print("<peer> stopped")
}

func discoverIPv4Address() (string, error){
	interfaceList, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var checkFlag = net.FlagMulticast | net.FlagPointToPoint | net.FlagUp
	for _, i := range interfaceList {
		if i.Flags&net.FlagLoopback != 0 {
			//ignore loopback
			continue
		}
		if i.Flags&checkFlag != 0 {
			addrs, err := i.Addrs()
			if err != nil{
				return "", err
			}
			if len(addrs) == 0{
				continue
			}
			for _, addr := range addrs{
				log.Printf("check %s", addr.String())
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil{
					return "", err
				}
				if ip.To4() != nil{
					return ip.String(), nil
				}
			}
		}
	}
	return "", errors.New("no interface available")
}

func Test_TwoEndpoint(t *testing.T){
	ListenAddress, err := discoverIPv4Address()
	if err != nil{
		t.Fatal(err)
	}
	t.Logf("local ip %s discovered", ListenAddress)
	endpoint1, err := CreateStubEndpoint(sonar.DefaultMulticastAddress, sonar.DefaultMulticastPort, sonar.DefaultDomain, ListenAddress)
	if err != nil{
		t.Fatal(err)
	}
	t.Log("core created")
	inf, err := InterfaceByAddress(ListenAddress)
	if err != nil{
		t.Fatal(err)
	}
	var core = CoreEndpoint{endpoint1}
	core.handler = &core
	if err = core.GenerateName(ServiceTypeCore, inf); err != nil{
		t.Fatal(err)
	}
	t.Logf("core name generated:%s", core.name)
	if err = core.Start(); err != nil{
		t.Fatal(err)
	}
	t.Log("core started")
	endpoint2, err := CreatePeerEndpoint(sonar.DefaultMulticastAddress, sonar.DefaultMulticastPort, sonar.DefaultDomain)
	if err != nil{
		t.Fatal(err)
	}
	t.Log("peer created")
	var peerChan = make(chan bool, 1)
	var peer = PeerEndpoint{endpoint2, peerChan}
	peer.handler = &peer
	if err = peer.GenerateName(ServiceTypeCell, inf); err != nil{
		t.Fatal(err)
	}
	t.Logf("peer name generated:%s", peer.name)
	if err = peer.Start();err != nil{
		t.Fatal(err)
	}

	{
		const connectTimeout = 3*time.Second
		//wait connected
		var timer = time.NewTimer(connectTimeout)
		select {
		case <- timer.C:
			//timeout
			t.Fatal("wait peer connect timeout")
		case <- peerChan:
			t.Log("peer connected")
		}
	}
	time.Sleep( 1 * time.Second)
	if err = core.Stop(); err != nil{
		t.Fatal(err)
	}
	t.Log("core stopped")
	{
		const disconnectTimeout = 30*time.Second
		//wait disconnect
		var timer = time.NewTimer(disconnectTimeout)
		select {
		case <- timer.C:
			//timeout
			t.Fatal("wait peer disconnect timeout")
		case <- peerChan:
			t.Log("peer disconnected")
		}
	}
	if err = peer.Stop(); err != nil{
		t.Fatal(err)
	}
	t.Log("peer stopped")
}