package serving

import (
	"context"
	"flag"
	"fmt"
	"github.com/sean-tech/gokit/logging"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"log"
	"testing"
	"time"
)

type UserInfo struct {
	UserId int						`json:"user_id" validate:"required,min=1"`
	UserName string					`json:"user_name" validate:"required,eorp"`
}

type UserAddParameter struct {
	UserName string					`json:"user_name" validate:"required,eorp"`
	Password string					`json:"password" validate:"required,gte=1"`
}

type userServiceImpl struct {
}
var UserService = &userServiceImpl{}

func (this *userServiceImpl) UserAdd(ctx context.Context, parameter *UserAddParameter, user *UserInfo) error {
	*user = UserInfo{
		UserId:   1230090123,
		UserName: parameter.UserName,
	}
	//user.UserId = 1230090123
	//user.UserName = parameter.UserName
	return nil
}

func TestRpcServer(t *testing.T) {
	logging.Setup(logging.LogConfig{
		LogSavePath:     "/Users/lyra/Desktop/",
		LogPrefix:       "rpctest",
	})
	_rpc_testing = true
	RpcServerServe(RpcConfig{
		RunMode:              "debug",
		RpcPort:              9001,
		RpcPerSecondConnIdle: 500,
		ReadTimeout:          60 * time.Second,
		WriteTimeout:         60 * time.Second,
		SecretOpen:           false,
		ServerCert:           "1",
		ServerKey:            "1",
		ClientCert:           "1",
		ClientKey:            "1",
		EtcdRpcBasePath:      "sean.tech/webkit/serving/rpc",
		EtcdEndPoints:        []string{"127.0.0.1:2379"},
		Logger:               logging.Logger(),
	}, func(server *server.Server) {
		server.RegisterName("User", new(userServiceImpl), "")
	})

	var user = new(UserInfo)
	client := CreateRpcClient("User")
	if err := client.Call(context.Background(), "UserAdd", &UserAddParameter{
		UserName: "1237757@qq.com",
		Password: "Aa123456",
	}, user); err != nil {
		fmt.Printf("err---%s", err.Error())
	} else  {
		fmt.Printf("user--%+v", user)
	}

	//signal
	//quit := make(chan os.Signal)
	//signal.Notify(quit, os.Interrupt)
	//<- quit
	//log.Println("Shutdown Server ...")
}


func TestParameter(t *testing.T) {
	var user = new(UserInfo)
	if err := UserService.UserAdd(context.Background(), &UserAddParameter{
		UserName: "1237757@qq.com",
		Password: "Aa123456",
	}, user); err != nil {
		fmt.Printf("err---%s", err.Error())
	} else  {
		fmt.Printf("user--%+v", user)
	}
}




type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

type Arith int

func (t *Arith) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func TestRpcst(t *testing.T) {
	flag.Parse()
	s := server.NewServer()
	addRegistryPlugin(s)

	s.RegisterName("Arith", new(Arith), "")

	go func() {
		s.Serve("tcp", ":9003")
	}()

	d := client.NewInprocessDiscovery()
	xclient := client.NewXClient("Arith", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	defer xclient.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	for i := 0; i < 100; i++ {
		reply := &Reply{}
		err := xclient.Call(context.Background(), "Mul", args, reply)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}

		log.Printf("%d * %d = %d", args.A, args.B, reply.C)
	}
}

func addRegistryPlugin(s *server.Server) {
	r := client.InprocessClient
	s.Plugins.Add(r)
}