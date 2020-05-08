package serving

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"github.com/smallnest/rpcx/client"
	rpcxLog "github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

type RpcConfig struct {
	RunMode string							`json:"run_mode" validate:"required,oneof=debug release"`
	RpcPort               	int				`json:"rpc_port" validate:"required,min=1,max=10000"`
	RpcPerSecondConnIdle  	int64			`json:"rpc_per_second_conn_idle" validate:"required,gte=1"`
	ReadTimeout           	time.Duration	`json:"read_timeout" validate:"required,gte=1"`
	WriteTimeout          	time.Duration	`json:"write_timeout" validate:"required,gte=1"`
	// tls
	SecretOpen 				bool   			`json:"secret_open"`
	ServerCert 				string 			`json:"server_cert" validate:"required,gte=1"`
	ServerKey  				string			`json:"server_key" validate:"required,gte=1"`
	ClientCert 				string 			`json:"client_cert" validate:"required,gte=1"`
	ClientKey  				string 			`json:"client_key" validate:"required,gte=1"`
	// etcd
	EtcdRpcBasePath 		string			`json:"etcd_rpc_base_path" validate:"required,gte=1"`
	EtcdEndPoints 			[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	// log
	Logger 				rpcxLog.Logger
}
/** 服务注册回调函数 **/
type RpcRegisterFunc func(server *server.Server)

var (
	_rpcConfig RpcConfig
	_rpc_testing bool = false
)

/**
 * 启动 服务server
 * registerFunc 服务注册回调函数
 */
func RpcServerServe(config RpcConfig, registerFunc RpcRegisterFunc) {
	if err := validate.ValidateParameter(config); err != nil {
		log.Fatal(err)
	}
	_rpcConfig = config

	rpcxLog.SetLogger(_rpcConfig.Logger)

	var s *server.Server
	if config.SecretOpen {
		//cert, err := tls.LoadX509KeyPair(config.App.RuntimeRootPath + config.App.TLSCerPath, config.App.RuntimeRootPath + config.App.TLSKeyPath)
		cert, err := tls.X509KeyPair([]byte(config.ServerCert), []byte(config.ServerKey))
		if err != nil {
			log.Fatal(err)
			return
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		s = server.NewServer(server.WithTLSConfig(tlsConfig))
	} else {
		s = server.NewServer(server.WithReadTimeout(config.ReadTimeout))
	}

	address := fmt.Sprintf(":%d", config.RpcPort)
	s.Plugins.Add(RpcLogger)
	RegisterPluginEtcd(s, address)
	RegisterPluginRateLimit(s)

	registerFunc(s)
	go func() {
		err := s.Serve("tcp", address)
		if err != nil {
			log.Fatalf("server start error : %v", err)
		}
	}()
}

/**
 * 注册插件，Etcd注册中心，服务发现
 */
func RegisterPluginEtcd(s *server.Server, serviceAddr string)  {
	if _rpc_testing == true {
		plugin := client.InprocessClient
		s.Plugins.Add(plugin)
		return
	}
	plugin := &serverplugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + serviceAddr,
		EtcdServers:    _rpcConfig.EtcdEndPoints,
		BasePath:       _rpcConfig.EtcdRpcBasePath,
		Metrics:        metrics.NewRegistry(),
		Services:       nil,
		UpdateInterval: time.Minute,
		Options:        nil,
	}
	err := plugin.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(plugin)
}

/**
 * 注册插件，限流器，限制客户端连接数
 */
func RegisterPluginRateLimit(s *server.Server)  {
	var fillSpeed float64 = 1.0 / float64(_rpcConfig.RpcPerSecondConnIdle)
	fillInterval := time.Duration(fillSpeed * math.Pow(10, 9))
	plugin := serverplugin.NewRateLimitingPlugin(fillInterval, _rpcConfig.RpcPerSecondConnIdle)
	s.Plugins.Add(plugin)
}

/**
 * 创建rpc调用客户端，基于Etcd服务发现
 */
func CreateRpcClient(serviceName string) client.XClient {
	option := client.DefaultOption
	option.Heartbeat = true
	option.HeartbeatInterval = time.Second
	option.ReadTimeout = _rpcConfig.ReadTimeout
	option.WriteTimeout = _rpcConfig.WriteTimeout
	if _rpcConfig.SecretOpen {
		option.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	xclient := client.NewXClient(serviceName, client.Failover, client.RoundRobin, *getDiscovery(serviceName), option)
	return xclient
}
var discoveryMap sync.Map
func getDiscovery(serviceName string) *client.ServiceDiscovery {
	if discovery, ok := discoveryMap.Load(serviceName); ok {
		return discovery.(*client.ServiceDiscovery)
	}
	var discovery client.ServiceDiscovery
	if _rpc_testing == true {
		discovery = client.NewInprocessDiscovery()
	} else {
		discovery = client.NewEtcdDiscovery(_rpcConfig.EtcdRpcBasePath, serviceName, _rpcConfig.EtcdEndPoints, nil)
	}
	discoveryMap.Store(serviceName, &discovery)
	return &discovery
}

type IRpcxLogger interface {
	Rpcx(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}
type rpclogger struct {
}
var RpcLogger = &rpclogger{}

func (this *rpclogger) PostReadRequest(ctx context.Context, r *protocol.Message, e error) error {
	logPrint("PostReadRequest", ctx, r, e)
	return nil
}

func (this *rpclogger)PreHandleRequest(ctx context.Context, r *protocol.Message) error {
	return nil
}
func (this *rpclogger) PreWriteResponse(ctx context.Context, req *protocol.Message, resp *protocol.Message) error {
	return nil
}

func (this *rpclogger) PostWriteResponse(ctx context.Context, req *protocol.Message, resp *protocol.Message, e error) error {
	logPrint("PostWriteResponse", ctx, resp, e)
	return nil
}

func (this *rpclogger) PreWriteRequest(ctx context.Context) error {
	return nil
}
func (this *rpclogger) PostWriteRequest(ctx context.Context, r *protocol.Message, e error) error {
	return nil
}

func logPrint(prefix string, ctx context.Context, msg *protocol.Message, e error)  {
	if e != nil {
		_rpcConfig.Logger.Errorf("[RPCX] %s error:%s", prefix, e.Error())
		return
	}

	var request_id uint64 = 0
	var user_name string = ""
	if requisition := foundation.GetRequisition(ctx); requisition != nil {
		request_id = requisition.RequestId
		user_name = requisition.UserName
	}
	payload := strings.ToValidUTF8(string(msg.Payload), ":")
	var info = fmt.Sprintf("%s request_id:%d | user_name:%s | service_call:%s.%s | metadata:%s | payload:%s ",
		prefix, request_id, user_name, msg.ServicePath, msg.ServiceMethod, msg.Metadata, payload)
	if logger, ok := _rpcConfig.Logger.(IRpcxLogger); ok {
		logger.Rpcx(info)
	} else {
		_rpcConfig.Logger.Infof("[RPCX] %s", info)
	}
}