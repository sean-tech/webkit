package config

import (
	"encoding/json"
	"fmt"
	"github.com/sean-tech/gokit/fileutils"
	"io/ioutil"
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
	*config = AppConfig{
		Log:     nil,
		Http:    nil,
		Rpc:     nil,
		Mysql:   nil,
		Redis:   nil,
		CE: 	 nil,
	}
	return nil
}

func TestRunServer(t *testing.T) {
	ConfigCernterServing(new(CC), 1234, []string{"192.168.1.20"})
}

func TestCallServer(t *testing.T) {
	config := configLoadFromCenter("192.168.1.20:1234", "webkittest", "user")
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

