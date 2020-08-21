package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/storage"
	"github.com/sean-tech/webkit/config"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"github.com/smallnest/rpcx/server"
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"
	"time"
)

const (
	SERVICE_AUTH = "Auth"
	SERVER_AUTH = "auth"
)

var authConfig = AuthConfig{
	WorkerId: 				 1,
	TokenSecret:             "thisnand!abn",
	TokenIssuer:             "sean-tech/webkit/auth",
	RefreshTokenExpiresTime: 120 * time.Second,
	AccessTokenExpiresTime:  30 * time.Second,
	AuthCode: "this is auth code for validate",
}

var testconfig = &config.AppConfig{
	Log: nil,
	Http:    &gohttp.HttpConfig{
		RunMode:      "debug",
		WorkerId:     3,
		HttpPort:     9002,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		RsaOpen: false,
		Rsa: nil,
	},
	Rpc:     &gorpc.RpcConfig{
		RunMode:              "debug",
		RpcPort:              9001,
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

func TestAuthServer(t *testing.T) {
	// concurrent
	runtime.GOMAXPROCS(runtime.NumCPU())
	// config
	config.Setup("webkittest", "auth", 9002, 9001, "/Users/sean/Desktop/", testconfig, func(appConfig *config.AppConfig) {
		// log start
		logging.Setup(*appConfig.Log)
		// database start
		//database.SetupRedis(*appConfig.Redis).Open()
		// auth setup
		//Setup(authConfig, database.Redis())
		Setup(authConfig, storage.Hash())
		// service start
		gorpc.ServerServe(*appConfig.Rpc, logging.Logger(), RegisterService)
		// server start
		gohttp.HttpServerServe(*appConfig.Http, logging.Logger(), RegisterApi)
	})
}

func RegisterService(server *server.Server)  {
	server.RegisterName(SERVICE_AUTH, Service(), "")
}

func RegisterApi(engine *gin.Engine)  {

	apiv1 := engine.Group("api/v1/user/auth/")
	{
		apiv1.POST("new", Api().NewAuth)
		apiv1.POST("refresh", Api().AuthRefresh)
		apiv1.POST("accesstoken", Api().AccessTokenAuth)
	}
}

func TestNewAuth(t *testing.T) {
	var url = "http://localhost:9002/api/v1/user/auth/new"
	var parameter = map[string]interface{}{
		"auth_code" : "this is auth code for validate",
		"uuid" : "kasnzncuhbajdjabdjazxc12345asd",
		"userid" : 12300901101,
		"roleid" : 1001,
		"username" : "sean",
		"client" : "iOS",
	}
	jsonStr, err := json.Marshal(parameter)
	if err != nil {
		fmt.Printf("to json error:%v\n", err)
		return
	}
	fmt.Println("--------------new----------------")
	var resp map[string]interface{}
	if resp, err = post(url, jsonStr); err == nil {
		resp, _ = resp["data"].(map[string]interface{})
		fmt.Println("--------------access auth----------------")
		url = "http://localhost:9002/api/v1/user/auth/accesstoken"
		parameter = map[string]interface{}{
			"access_token" : resp["access_token"],
		}
		jsonStr, err = json.Marshal(parameter)
		if err != nil {
			fmt.Printf("to json error:%v\n", err)
			return
		}
		post(url, jsonStr)

		fmt.Println("--------------refresh----------------")

		url = "http://localhost:9002/api/v1/user/auth/refresh"
		parameter = map[string]interface{}{
			"refresh_token" : resp["refresh_token"],
			"access_token" : resp["access_token"],
		}
		jsonStr, err = json.Marshal(parameter)
		if err != nil {
			fmt.Printf("to json error:%v\n", err)
			return
		}
		post(url, jsonStr)
	} else {
		t.Error(err)
	}
}

func post(url string, jsonStr []byte) (response map[string]interface{}, err error) {
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	var resp *http.Response
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		fmt.Printf("resp error:%v\n", err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("status_ode:%d\nheader:%+v\nbody:%s\n", resp.StatusCode, resp.Header, body)
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response, nil
}