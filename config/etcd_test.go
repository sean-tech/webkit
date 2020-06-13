package config

import (
	"fmt"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"testing"
	"time"
)

const module = "user"
const salt = "asdasdasdadzxczc"

func TestPut(t *testing.T) {
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

func TestDelete(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if err := DeleteConfig(module); err != nil {
		t.Error(err)
	}
	fmt.Println("delete success")
}

func TestGet(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	var cfg *AppConfig; var err error
	if cfg, err = GetConfig(module, salt); err != nil {
		t.Error(err)
	}
	fmt.Println("get success")
	fmt.Printf("%+v\n", cfg)
}

func TestGetAllModules(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if modules, err := GetAllModules(); err != nil {
		t.Error(err)
	} else {
		fmt.Println(modules)
	}
}

func TestPutWorkerId(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if err := PutWorkerId(1, module, "192.168.1.20"); err != nil {
		t.Error(err)
	}
	fmt.Println("workerid put success")
}

func TestGetWorkerId(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if workerId, err := GetWorkerId(module, "192.168.1.20"); err != nil {
		t.Error(err)
	} else {
		fmt.Println("workerid get success : ", workerId)
	}
}

func TestDeleteWorkerId(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if err := DeleteWorkerId(module, "192.168.1.20"); err != nil {
		t.Error(err)
	}
	fmt.Println("workerid delete success")
}

func TestGetAllWorkers(t *testing.T) {
	clientInit(&CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	})
	if workers, err := GetAllWorkers(module); err != nil {
		t.Error(err)
	} else {
		fmt.Println("all workers get success : ", workers)
	}
}


var testconfig = &AppConfig{
	Log: &logging.LogConfig{
		RunMode:     "debug",
		LogSavePath: "/Users/sean/Desktop/",
		LogPrefix:   "config",
	},
	RsaOpen: false,
	Rsa:     nil,
	Http:    &gohttp.HttpConfig{
		RunMode:      "debug",
		WorkerId:     3,
		HttpPort:     9022,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	},
	Rpc:     &gorpc.RpcConfig{
		RunMode:              "debug",
		RpcPort:              9021,
		RpcPerSecondConnIdle: 500,
		ReadTimeout:          60 * time.Second,
		WriteTimeout:         60 * time.Second,
		TokenSecret:          "th!@#isasd",
		TokenIssuer:          "/sean-tech/webkit/auth",
		TlsOpen:              false,
		Tls:                  nil,
		WhiteListOpen:        false,
		WhiteListIps:         nil,
		EtcdEndPoints:        []string{"127.0.0.1:2379"},
		EtcdRpcBasePath:      "/sean-tech/webkit/rpc",
		EtcdRpcUserName:      "root",
		EtcdRpcPassword:      "etcd.user.root.pwd",
	},
	Mysql:   &database.MysqlConfig{
		WorkerId:    3,
		Type:        "mysql",
		User:        "root",
		Password:    "admin2018",
		Hosts: 		 map[int]string{0:"127.0.0.1:3306"},
		Name:        "etcd_center",
		MaxIdle:     30,
		MaxOpen:     30,
		MaxLifetime: 200 * time.Second,
	},
	Redis:   &database.RedisConfig{
		Host:        "127.0.0.1:6379",
		Password:    "",
		MaxIdle:     30,
		MaxActive:   30,
		IdleTimeout: 200 * time.Second,
	},
}