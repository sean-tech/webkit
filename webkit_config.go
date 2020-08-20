package webkit

import (
	"github.com/sean-tech/webkit/config"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"time"
)

var testconfig = &config.AppConfig{
	Log: &logging.LogConfig{
		RunMode:     "debug",
		LogSavePath: "/Users/sean/Desktop/",
		LogPrefix:   "webkit",
	},
	Http:    &gohttp.HttpConfig{
		RunMode:      "debug",
		WorkerId:     3,
		HttpPort:     9022,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		RsaOpen: false,
		Rsa:     nil,
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
		Registry: &gorpc.EtcdRegistry{
			EtcdEndPoints:        []string{"127.0.0.1:2379"},
			EtcdRpcBasePath:      "/sean-tech/webkit/rpc",
			EtcdRpcUserName:      "root",
			EtcdRpcPassword:      "etcd.user.root.pwd",
		},
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
	CE: &config.ConfigEtcd{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigBasePath: "/sean-tech/webkit/config",
		EtcdConfigUserName: "module_config_user",
		EtcdConfigPassword: "module_config_user_pwd",
	},
}