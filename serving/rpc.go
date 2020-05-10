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
	"net"
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
	s.Plugins.Add(ServerLogger)
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



var clientMap sync.Map
/**
 * 创建rpc调用客户端，基于Etcd服务发现
 */
func CreateRpcClient(serviceName string) client.XClient {
	if c, ok := clientMap.Load(serviceName); ok {
		return c.(client.XClient)
	}
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
	xclient := client.NewXClient(serviceName, client.Failover, client.RoundRobin, newDiscovery(serviceName), option)
	xclient.GetPlugins().Add(ClientLogger)
	clientMap.Store(serviceName, xclient)
	return xclient
}

func newDiscovery(serviceName string) client.ServiceDiscovery {
	var discovery client.ServiceDiscovery
	if _rpc_testing == true {
		discovery = client.NewInprocessDiscovery()
	} else {
		discovery = client.NewEtcdDiscovery(_rpcConfig.EtcdRpcBasePath, serviceName, _rpcConfig.EtcdEndPoints, nil)
	}
	return discovery
}



type IRpcxLogger interface {
	Rpcx(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

func (this *serverlogger) Register(name string, rcvr interface{}, metadata string) error {
	this.register(rcvr, name, true)
	return nil
}

func (this *serverlogger) Unregister(name string) error {
	this.UnRegister(name)
	return nil
}

func (this *serverlogger) PostReadRequest(ctx context.Context, r *protocol.Message, e error) error {
	this.logPrint("PostReadRequest", ctx, r, MsgTypeReq, e)
	return nil
}

func (this *serverlogger)PreHandleRequest(ctx context.Context, r *protocol.Message) error {
	return nil
}
func (this *serverlogger) PreWriteResponse(ctx context.Context, req *protocol.Message, resp *protocol.Message) error {
	return nil
}

func (this *serverlogger) PostWriteResponse(ctx context.Context, req *protocol.Message, resp *protocol.Message, e error) error {
	this.logPrint("PostWriteResponse", ctx, resp, MsgTypeResp, e)
	return nil
}

func (this *serverlogger) PreWriteRequest(ctx context.Context) error {
	return nil
}
func (this *serverlogger) PostWriteRequest(ctx context.Context, r *protocol.Message, e error) error {
	return nil
}

func (this *serverlogger) logPrint(prefix string, ctx context.Context, msg *protocol.Message, msgType MsgType, e error)  {
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
	data := this.paylodConvert(ctx, msg, msgType)
	var info = fmt.Sprintf("%s request_id:%d | user_name:%s | service_call:%s.%s | metadata:%s | payload:%+v ",
		prefix, request_id, user_name, msg.ServicePath, msg.ServiceMethod, msg.Metadata, data)
	if logger, ok := _rpcConfig.Logger.(IRpcxLogger); ok {
		logger.Rpcx(info)
	} else {
		_rpcConfig.Logger.Infof("[RPCX] %s", info)
	}
}



type clientLogger struct {}
var ClientLogger = &clientLogger{}

func (this *clientLogger) DoPreCall(ctx context.Context, servicePath, serviceMethod string, args interface{}) error {
	return nil
}

// PostCallPlugin is invoked after the client calls a server.
func (this *clientLogger) DoPostCall(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, err error) error {
	if err != nil {
		_rpcConfig.Logger.Errorf("[RPCX] DoPostCall %s.%s args:%+v reply:%+v error:%s", servicePath, serviceMethod, args, reply, err.Error())
		return nil
	}
	var info = fmt.Sprintf("[RPCX] DoPostCall %s.%s args:%+v reply:%+v", servicePath, serviceMethod, args, reply)
	if logger, ok := _rpcConfig.Logger.(IRpcxLogger); ok {
		logger.Rpcx(info)
	} else {
		_rpcConfig.Logger.Infof("[RPCX] %s", info)
	}
	return nil
}

// ConnCreatedPlugin is invoked when the client connection has created.
func (this *clientLogger) ConnCreated(conn net.Conn) (net.Conn, error) {
	return conn, nil
}

// ClientConnectedPlugin is invoked when the client has connected the server.
func (this *clientLogger) ClientConnected(conn net.Conn) (net.Conn, error) {
	return conn, nil
}

// ClientConnectionClosePlugin is invoked when the connection is closing.
func (this *clientLogger) ClientConnectionClose(net.Conn) error {
	return nil
}

// ClientBeforeEncodePlugin is invoked when the message is encoded and sent.
func (this *clientLogger) ClientBeforeEncode(*protocol.Message) error {
	return nil
}

// ClientAfterDecodePlugin is invoked when the message is decoded.
func (this *clientLogger) ClientAfterDecode(*protocol.Message) error {
	return nil
}
