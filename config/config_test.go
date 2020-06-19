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

func TestConfigLoadFromLocal(t *testing.T) {
	Setup(module, salt, testconfig, "", func(appConfig *AppConfig) {
		fmt.Println(appConfig)
	})
}

func TestConfigLoadFromLocalAndEtcdPut(t *testing.T) {
	var cmdParams = CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	}
	var config_etcd_info = CmdEncrypt(cmdParams, module, salt)
	Setup(module, salt, testconfig, config_etcd_info, func(appConfig *AppConfig) {
		fmt.Println(appConfig)
		if err := PutConfig(appConfig, "test", salt); err != nil {
			t.Error(err)
		}
		fmt.Println("put success")
	})
}

func TestConfig(t *testing.T) {
	put(t)
	var cmdParams = CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	}
	var config_etcd_info = CmdEncrypt(cmdParams, module, salt)
	Setup(module, salt, testconfig, config_etcd_info, func(appConfig *AppConfig) {
		if err := PutConfig(appConfig, "test", salt); err != nil {
			t.Error(err)
		}
		fmt.Println("put success")
		if modules, err := GetAllModules(); err != nil {
			t.Error(err)
		} else {
			fmt.Println(modules)
		}
		get(t)
		delete(t)
		//get(t)
		putWorkerId(t)
		getWorkerId(t)
		getAllWorkers(t)
		deleteWorkerId(t)
		getAllWorkers(t)
	})
}

func put(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if err := PutConfig(testconfig, module, salt); err != nil {
		t.Error(err)
	}
	fmt.Println("put success")
}

func delete(t *testing.T) {
	if err := DeleteConfig(module); err != nil {
		t.Error(err)
	}
	fmt.Println("delete success")
}

func get(t *testing.T) {
	var cfg *AppConfig; var err error
	if cfg, err = GetConfig(module, salt); err != nil {
		t.Error(err)
	}
	fmt.Println("get success")
	fmt.Printf("%+v\n", cfg)
}

func putWorkerId(t *testing.T) {
	if err := PutWorkerId(1, module, "192.168.1.20"); err != nil {
		t.Error(err)
	}
	fmt.Println("workerid put success")
}

func getWorkerId(t *testing.T) {
	if workerId, err := GetWorkerId(module, "192.168.1.20"); err != nil {
		t.Error(err)
	} else {
		fmt.Println("workerid get success : ", workerId)
	}
}

func deleteWorkerId(t *testing.T) {
	if err := DeleteWorkerId(module, "192.168.1.20"); err != nil {
		t.Error(err)
	}
	fmt.Println("workerid delete success")
}

func getAllWorkers(t *testing.T) {
	if workers, err := GetAllWorkers(module); err != nil {
		t.Error(err)
	} else {
		fmt.Println("all workers get success : ", workers)
	}
}

func TestJsonConvert(t *testing.T) {
	jsonBytes, err := json.Marshal(testconfig)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(jsonBytes))
}
