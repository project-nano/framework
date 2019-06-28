package framework

import (
	"strconv"
	"fmt"
	"errors"
	"net"
	"strings"
)

func SelectEthernetInterface(description string, requireUpLink bool) (name string, err error) {
	interfaceList, err := net.Interfaces()
	if err != nil {
		return
	}
	var checkFlag net.Flags
	if requireUpLink{
		checkFlag = net.FlagMulticast | net.FlagPointToPoint | net.FlagUp
	}else{
		checkFlag = net.FlagMulticast | net.FlagPointToPoint
	}
	var options []string
	for _, i := range interfaceList {
		if i.Flags&net.FlagLoopback != 0 {
			//ignore loop back
			continue
		}
		if i.Flags&checkFlag != 0 {
			if strings.HasPrefix(i.Name, "e"){
				options = append(options, i.Name)
			}
		}
	}
	if 0 == len(options){
		return "", errors.New("no ethernet interface available")
	}else if 1 == len(options){
		return options[0], nil
	}
	//output select menu
	var selectMap = map[int]string{} //index to name
	for index, name := range options{
		selectMap[index] = name
	}
	var input string
	var exists bool
	for {
		for index, name := range selectMap{
			fmt.Printf("%d> %s\n", index, name)
		}
		fmt.Printf("enter index to select interface as %s: ", description)
		fmt.Scanln(&input)
		selected, err := strconv.Atoi(input)
		if err != nil{
			fmt.Printf("invalid input : %s", input)
			continue
		}
		name, exists = selectMap[selected]
		if !exists{
			fmt.Printf("invalid selection: %d", selected)
			continue
		}
		return name, nil
	}

}

func ChooseIPV4Address(description string) (address string, err error){
	options, err := searchIPv4Address()
	if err != nil{
		return
	}
	if 1 == len(options){
		address, err =  InputString(description, options[0])
	}else{
		var optionMap = map[string]string{}//index to ip
		for index, address := range options{
			key := strconv.Itoa(index)
			optionMap[key] = address
			fmt.Printf("%d> %s\n", index, address)
		}
		fmt.Printf("enter index to select address as %s, or input a new address: ", description)
		var input string
		fmt.Scanln(&input)
		if "" == input{
			err = errors.New("must input an address")
			return
		}
		var exists bool
		if address, exists = optionMap[input];!exists{
			address = input
		}
	}
	if nil != net.ParseIP(address){
		//valid ip format
		return address, nil
	}
	return "", fmt.Errorf("invalid address value '%s'", address)
}

func InputInteger(description string, defaultValue int) (value int, err error) {
	fmt.Printf("%s = %d (press enter to accept or input new value): ", description, defaultValue)
	var input string
	fmt.Scanln(&input)
	if "" == input{
		//default value
		value = defaultValue
		return
	}
	value, err = strconv.Atoi(input)
	if err != nil{
		err = fmt.Errorf("invalid input %s", input)
		return
	}
	return
}

func InputString(description, defaultValue string) (value string, err error){
	fmt.Printf("%s = '%s' (press enter to accept or input new value): ", description, defaultValue)
	var input string
	fmt.Scanln(&input)
	if "" == input{
		//default value
		value = defaultValue
	}else{
		value = input
	}
	if "" == value{
		err = errors.New("no empty string allowed")
	}
	return
}

func InputIPAddress(description, defaultValue string) (value string, err error){
	fmt.Printf("%s = '%s' (press enter to accept or input new value): ", description, defaultValue)
	var input string
	fmt.Scanln(&input)
	if "" == input{
		//default value
		value = defaultValue
	}else{
		value = input
	}
	if nil == net.ParseIP(value){
		err = fmt.Errorf("invalid address '%s'", value)
		return "", err
	}
	return
}

func InputMultiCastAddress(description, defaultValue string) (address string, err error)  {
	fmt.Printf("%s = '%s' (press enter to accept or input new value): ", description, defaultValue)
	var input string
	fmt.Scanln(&input)
	if "" == input{
		//default value
		address = defaultValue
	}else{
		address = input
	}
	var ip = net.ParseIP(address)
	if nil == ip{
		err = fmt.Errorf("invalid address '%s'", address)
		return
	}
	if !ip.IsMulticast(){
		err = fmt.Errorf("'%s' not a multicast address", address)
		return
	}
	return address, nil
}

func InputNetworkPort(description string, defaultValue int) (port int, err error) {
	const (
		MaxPort = 0xFFFF
	)
	fmt.Printf("%s = %d (press enter to accept or input new value): ", description, defaultValue)
	var input string
	fmt.Scanln(&input)
	if "" == input{
		//default value
		port = defaultValue
	}else{
		port, err = strconv.Atoi(input)
		if err != nil{
			err = fmt.Errorf("invalid input %s", input)
			return
		}
	}
	if port > MaxPort{
		err = fmt.Errorf("invalid network port %d", port)
	}
	return
}

func searchIPv4Address() (addresses []string, err error) {
	interfaceList, err := net.Interfaces()
	if err != nil {
		return
	}
	var checkFlag = net.FlagMulticast | net.FlagPointToPoint | net.FlagUp
	for _, i := range interfaceList {
		if i.Flags&net.FlagLoopback != 0 {
			//ignore loopback
			continue
		}
		if i.Flags&checkFlag != 0 {
			addrs, err := i.Addrs()
			if err != nil {
				return nil, err
			}
			if len(addrs) == 0 {
				continue
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					return nil, err
				}
				if ip.To4() != nil {
					addresses = append(addresses, ip.String())
				}
			}
		}
	}
	if 0 == len(addresses){
		return nil, errors.New("no interface available")
	}
	return
}