package config

import (
	"encoding/json"
	"fmt"
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

func (this *CC) ConfigLoad(path string, config *AppConfig) error {
	fmt.Println("path is ", path)
	*config = AppConfig{
		Log:     nil,
		RsaOpen: false,
		Rsa:     nil,
		Http:    nil,
		Rpc:     nil,
		Mysql:   nil,
		Redis:   nil,
	}
	return nil
}

func TestRunServer(t *testing.T) {
	ConfigCernterServing(new(CC), 1234, []string{"192.168.1.20"})
}

func TestCallServer(t *testing.T) {
	config := ConfigCenterLoading("192.168.1.20:1234", "user")
	fmt.Println(config)
}

func TestLocalIp(t *testing.T) {
	fmt.Println(GetIPs())
	fmt.Println(GetLocalIP())
}



