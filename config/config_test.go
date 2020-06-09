package config

import (
	"fmt"
	"testing"
)

func TestSecret(t *testing.T) {

	var bts = []byte("hahaser_@1")
	fmt.Printf("%+v\n", bts)
	fmt.Printf("long of bytes : %d\n", len(bts))
}

func TestConfig(t *testing.T) {
	var cmdParams = &CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	}
	Setup("test", "salt", nil, cmdParams, func(appConfig *AppConfig) {
		put(t)
		if err := PutConfig(cfg, "test", salt); err != nil {
			t.Error(err)
		}
		fmt.Println("put success")
		if modules, err := GetAllModules(); err != nil {
			t.Error(err)
		} else {
			fmt.Println(modules)
		}
		//get(t)
		//delete(t)
		////get(t)
		//putWorkerId(t)
		//getWorkerId(t)
		//getAllWorkers(t)
		//deleteWorkerId(t)
		//getAllWorkers(t)
	})
}

func put(t *testing.T) {
	if err := PutConfig(cfg, module, salt); err != nil {
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
