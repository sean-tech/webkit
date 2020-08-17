package config

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strings"
)

const ConfigCenterServiceName = "sean,tech/webkit/configcenter"
const ConfigLoadMethodName = ".AppConfigLoad"

type Worker struct {
	Product string	`json:"product"`
	Module 	string 	`json:"module"`
	Ip 		string	`json:"ip"`
}

type IConfigCenter interface {
	AppConfigLoad(worker *Worker, appcfg *AppConfig) error
}

func ConfigCernterServing(cc IConfigCenter, port int, whitelistips []string) {
	rpc.RegisterName(ConfigCenterServiceName, cc)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		var clientIp = conn.RemoteAddr().String()
		if ip, _, err := net.SplitHostPort(strings.TrimSpace(conn.RemoteAddr().String())); err == nil {
			clientIp = ip
		}
		if whiteListIpsFitter(clientIp, whitelistips) == true {
			go rpc.ServeConn(conn)
		} else {
			conn.Close()
		}
	}
}

func whiteListIpsFitter(clientIp string, whitelistips []string) bool {
	if whitelistips == nil {
		return true
	}
	for _, ip := range whitelistips {
		if clientIp == ip {
			return true
		}
	}
	return false
}



func GetIPs() (ips []string){
	addrs,err := net.InterfaceAddrs()
	if err != nil{
		//fmt.Println("get ip arr failed: ",err)
		return nil
	}
	for _,addr := range addrs{
		if ipnet,ok := addr.(*net.IPNet);ok && !ipnet.IP.IsLoopback(){
			if ipnet.IP.To4() != nil{
				ips = append(ips,ipnet.IP.String())
			}
		}
	}
	return ips
}

func GetLocalIP() string {
	addrs,err := net.InterfaceAddrs()
	if err != nil{
		return ""
	}
	for _,addr := range addrs{
		if ipnet,ok := addr.(*net.IPNet);ok && !ipnet.IP.IsLoopback(){
			if isLocal(ipnet.IP.To4()) {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func isLocal(ip4 net.IP) bool {
	if ip4 == nil {
		return false
	}
	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}