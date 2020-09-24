package config

import (
	"encoding/json"
	"fmt"
	"github.com/sean-tech/gokit/fileutils"
	"github.com/sean-tech/webkit/logging"
	"io/ioutil"
	"net"
	"net/rpc"
	"testing"
)

func TestSecret(t *testing.T) {

	var bts = []byte("hahaser_@1")
	fmt.Printf("%+v\n", bts)
	fmt.Printf("long of bytes : %d\n", len(bts))
}

func TestLoadJson(t *testing.T) {
	jsonBytes, err := ioutil.ReadFile("./config.json")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(jsonBytes))
	var appConfig = new(AppConfig)
	if err := json.Unmarshal(jsonBytes, appConfig); err != nil {
		t.Error(err)
	} else {
		fmt.Println(appConfig)
	}
}

type CC struct {

}

func (this *CC) AppConfigLoad(worker *Worker, config *AppConfig) error {
	fmt.Println("module is ", worker.Module, "ip is ", worker.Ip)
	//panic("ha ha i am panicing")
	*config = AppConfig{
		Log:     &logging.LogConfig{
			RunMode:     "debug",
			LogSavePath: "/ahdsjadh",
			LogPrefix:   "zc",
		},
		Http:    nil,
		Rpc:     nil,
		Mysql:   nil,
		Redis:   nil,
		CE: 	 nil,
	}
	return nil
}

func TestRunServer(t *testing.T) {
	go ConfigCernterServing(new(CC), 1234, func(clientIp string) bool {
		if clientIp == "192.168.1.21" || clientIp == "172.20.10.2" {
			 return true
		}
		return false
	})
	select {

	}
}

func TestCallServer(t *testing.T) {
	config := configLoadFromCenter("172.20.10.2:1234", "webkittest", "user")
	fmt.Println(config)
}

func TestLocalIp(t *testing.T) {
	fmt.Println(GetIPs())
	fmt.Println(GetLocalIP())
}

func TestFilePahtJudge(t *testing.T) {
	if fileutils.CheckExist("/Users/sean/Desktop/") == true {
		fmt.Println("file exist")
	} else {
		fmt.Println("file not exist")
	}
}


func TestRpcServe(t *testing.T) {
	port := 1234
	cc := new(CC)
	rpc.RegisterName(ConfigCenterServiceName, &rcvr{cc: cc})
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			t.Fatal("Accept error:", err)
		}
		go serveconn(conn)
	}
}

func serveconn(conn net.Conn)  {
	rpc.ServeConn(conn)
}