package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/logging"
	"github.com/sean-tech/webkit/config"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/smallnest/rpcx/server"
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"
	"time"
)

const (
	SERVICE_AUTH = "Auth"
)

var debugConfig = &config.AppConfig{
	RsaOpen: false,
	Rsa:     nil,
	Http:    &gohttp.HttpConfig{
		RunMode:      "debug",
		WorkerId:     3,
		HttpPort:     9012,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	},
	Rpc:     &gorpc.RpcConfig{
		RunMode:              "debug",
		RpcPort:              9011,
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
		EtcdRpcBasePath:      "/sean-tech/webkit/auth/rpc",
		EtcdRpcUserName:      "root",
		EtcdRpcPassword:      "etcd.user.root.pwd",
	},
	Mysql:   nil,
	Redis:   &database.RedisConfig{
		Host:        "127.0.0.1:6379",
		Password:    "",
		MaxIdle:     30,
		MaxActive:   30,
		IdleTimeout: 200 * time.Second,
	},
}

func TestAuthServer(t *testing.T) {

	logging.Setup(logging.LogConfig{
		LogSavePath: "/Users/sean/Desktop/",
		LogPrefix:   "auth",
	})
	// concurrent
	runtime.GOMAXPROCS(runtime.NumCPU())
	// config
	config.Setup("auth", "adzxcqwrz", debugConfig, func(appConfig *config.AppConfig) {
		// auth setup
		idWorker, _ := foundation.NewWorker(appConfig.Http.WorkerId)
		Setup(AuthConfig{
			TokenSecret:             "thisnand!abn",
			TokenIssuer:             "sean-tech/webkit/auth",
			RefreshTokenExpiresTime: 120 * time.Second,
			AccessTokenExpiresTime:  30 * time.Second,
		}, idWorker)
		// database start
		database.SetupRedis(*appConfig.Redis).Open()
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
	var url = "http://localhost:9012/api/v1/user/auth/new"
	var parameter = map[string]interface{}{
		"auth_code" : "this is auth code for validate",
		"uuid" : "kasnzncuhbajdjabdjazxc12345asd",
		"userid" : 12300901101,
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
		url = "http://localhost:9012/api/v1/user/auth/accesstoken"
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

		url = "http://localhost:9012/api/v1/user/auth/refresh"
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
