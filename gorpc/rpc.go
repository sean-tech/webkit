package gorpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/docker/libkv/store"
	"github.com/rcrowley/go-metrics"
	"github.com/sean-tech/gokit/requisition"
	"github.com/sean-tech/gokit/validate"
	"github.com/smallnest/rpcx/client"
	rpcxLog "github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/smallnest/rpcx/share"
	"log"
	"math"
	"sync"
	"time"
)

type IRpcxLogger interface {
	Rpcx(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

// log
var _logger 	rpcxLog.Logger

type RpcConfig struct {
	RunMode string							`validate:"required,oneof=debug test release"`
	RpcPort               	int
	RpcPerSecondConnIdle  	int64			`json:"rpc_per_second_conn_idle" validate:"required,gte=1"`
	ReadTimeout           	time.Duration	`json:"read_timeout" validate:"required,gte=1"`
	WriteTimeout          	time.Duration	`json:"write_timeout" validate:"required,gte=1"`
	// token
	TokenSecret      		string        	`json:"token_secret" validate:"required,gte=1"`
	TokenIssuer      		string        	`json:"token_issuer" validate:"required,gte=1"`
	// tls
	TlsOpen					bool			`json:"tls_open"`
	Tls						*TlsConfig
	// whiteList
	WhiteListOpen 			bool			`json:"white_list_open"`
	WhiteListIps			[]string		`json:"white_list_ips"`
	// etcd
	EtcdEndPoints 			[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	EtcdRpcBasePath 		string			`json:"etcd_rpc_base_path" validate:"required,gte=1"`
	EtcdRpcUserName 		string			`json:"etcd_rpc_username" validate:"required,gte=1"`
	EtcdRpcPassword 		string			`json:"etcd_rpc_password" validate:"required,gte=1"`
}

type TlsConfig struct {
	CACert       			string 			`json:"ca_cert" validate:"required"`
	CACommonName 			string 			`json:"ca_common_name" validate:"required"`
	ServerCert   			string 			`json:"server_cert" validate:"required"`
	ServerKey    			string 			`json:"server_key" validate:"required"`
}

/** 服务注册回调函数 **/
type RpcRegisterFunc func(server *server.Server)

var (
	_config      RpcConfig
	_rpc_testing bool = false
)

/**
 * 启动 服务server
 * registerFunc 服务注册回调函数
 */
func ServerServe(config RpcConfig, logger rpcxLog.Logger, registerFunc RpcRegisterFunc) {
	// config validate
	if logger != nil {
		_logger = logger
		rpcxLog.SetLogger(_logger)
	}
	if err := validate.ValidateParameter(config); err != nil {
		log.Fatal(err)
	}
	if config.TlsOpen {
		if config.Tls == nil {
			log.Fatal("server rpc start error : secret is nil")
		}
		if err := validate.ValidateParameter(config.Tls); err != nil {
			log.Fatal(err)
		}
	}
	_config = config

	// server
	var s *server.Server
	if config.TlsOpen == false {
		s = server.NewServer(server.WithReadTimeout(config.ReadTimeout), server.WithWriteTimeout(config.WriteTimeout))
	} else {

		//cert, err := tls.LoadX509KeyPair(_config.ServerPemPath, _config.ServerKeyPath)
		cert, err := tls.X509KeyPair([]byte(config.Tls.ServerCert), []byte(config.Tls.ServerKey))
		if err != nil {
			log.Fatal(err)
			return
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM([]byte(_config.Tls.CACert))
		if !ok {
			panic("failed to parse root certificate")
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}
		s = server.NewServer(server.WithTLSConfig(tlsConfig), server.WithReadTimeout(config.ReadTimeout), server.WithWriteTimeout(config.WriteTimeout))
	}

	address := fmt.Sprintf(":%d", _config.RpcPort)
	registerPlugins(s, address)
	registerFunc(s)
	go func() {
		err := s.Serve("tcp", address)
		if err != nil {
			log.Fatalf("server start error : %v", err)
		}
	}()
}

func registerPlugins(s *server.Server, address string)  {
	s.Plugins.Add(ServerLogger)
	s.AuthFunc = serverAuth
	// white list
	if _config.WhiteListOpen {
		var wl = make(map[string]bool)
		for _, ip := range _config.WhiteListIps {
			wl[ip] = true
		}
		s.Plugins.Add(serverplugin.WhitelistPlugin{
			Whitelist:     wl,
			WhitelistMask: nil,
		})
	}
	
	RegisterPluginEtcd(s, address)
	RegisterPluginRateLimit(s)
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

	plugin := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + serviceAddr,
		EtcdServers:    _config.EtcdEndPoints,
		BasePath:       _config.EtcdRpcBasePath,
		Metrics:        metrics.NewRegistry(),
		Services:       nil,
		UpdateInterval: time.Minute,
		Options:        &store.Config{
			ClientTLS:         nil,
			TLS:               nil,
			ConnectionTimeout: 3 * time.Minute,
			Bucket:            "",
			PersistConnection: false,
			Username:          _config.EtcdRpcUserName,
			Password:          _config.EtcdRpcPassword,
		},
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
	var fillSpeed float64 = 1.0 / float64(_config.RpcPerSecondConnIdle)
	fillInterval := time.Duration(fillSpeed * math.Pow(10, 9))
	plugin := serverplugin.NewRateLimitingPlugin(fillInterval, _config.RpcPerSecondConnIdle)
	s.Plugins.Add(plugin)
}



var clientMap sync.Map
/**
 * 创建rpc调用客户端，基于Etcd服务发现
 */
func CreateClient(serviceName string) client.XClient {
	if c, ok := clientMap.Load(serviceName); ok {
		return c.(client.XClient)
	}
	option := client.DefaultOption
	option.Heartbeat = true
	option.HeartbeatInterval = time.Second
	option.ReadTimeout = _config.ReadTimeout
	option.WriteTimeout = _config.WriteTimeout
	if _config.TlsOpen {
		//cert, err := tls.LoadX509KeyPair(_config.ServerPemPath, _config.ServerKeyPath)
		cert, err := tls.X509KeyPair([]byte(_config.Tls.ServerCert), []byte(_config.Tls.ServerKey))
		if err != nil {
			if _logger != nil {
				_logger.Errorf("[RPCX]:unable to read cert.pem and cert.key : %s", err.Error())
			}
			goto OPTION_SECRET_SETED
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM([]byte(_config.Tls.CACert))
		if !ok {
			if _logger != nil {
				_logger.Errorf("[RPCX]:failed to parse root certificate : %s", err.Error())
			}
			goto OPTION_SECRET_SETED
		}
		option.TLSConfig = &tls.Config{
			RootCAs:            certPool,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: false,
			ServerName: serviceName + "." + _config.Tls.CACommonName,
		}
	}
OPTION_SECRET_SETED:
	xclient := client.NewXClient(serviceName, client.Failover, client.RoundRobin, newDiscovery(serviceName), option)
	xclient.GetPlugins().Add(ClientLogger)
	clientMap.Store(serviceName, xclient)
	return xclient
}

func newDiscovery(serviceName string) client.ServiceDiscovery {
	var discovery client.ServiceDiscovery
	var options = &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: 0,
		Bucket:            "",
		PersistConnection: false,
		Username:          _config.EtcdRpcUserName,
		Password:          _config.EtcdRpcPassword,
	}
	if _rpc_testing == true {
		discovery = client.NewInprocessDiscovery()
	} else {
		discovery = client.NewEtcdV3Discovery(_config.EtcdRpcBasePath, serviceName, _config.EtcdEndPoints, options)
	}
	return discovery
}

/**
 * call
 */
func Call(serviceName string, ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	client := CreateClient(serviceName)
	return ClientCall(client, ctx, serviceMethod, args, reply)
}

/**
 * client call
 */
func ClientCall(client client.XClient, ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {

	if ctx, err := clientAuth(client, ctx); err != nil {
		return err
	} else {
		return client.Call(ctx, serviceMethod, args, reply)
	}
}

/**
 * go
 */
func Go(serviceName string, ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *client.Call) (*client.Call, error) {
	client := CreateClient(serviceName)
	return ClientGo(client, ctx, serviceMethod, args, reply, done)
}

/**
 * client go
 */
func ClientGo(client client.XClient, ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *client.Call) (*client.Call, error) {
	if ctx, err := clientAuth(client, ctx); err != nil {
		return nil, err
	} else {
		return client.Go(ctx, serviceMethod, args, reply, done)
	}
}

/**
 * client auth
 */
func clientAuth(client client.XClient, ctx context.Context) (context.Context, error) {
	var req_id uint64 = 0; var user_id uint64 = 0; var user_name = ""
	if req := requisition.GetRequisition(ctx); req != nil {
		req_id = req.RequestId
		user_id = req.UserId
		user_name = req.UserName
	}
	var expiresTime = (_config.ReadTimeout + _config.WriteTimeout) * 2
	if token, err := generateToken(req_id, user_id, user_name, _config.TokenSecret, _config.TokenIssuer, expiresTime); err != nil {
		return ctx, err
	} else {
		client.Auth(token)
	}

	if ctx == nil {
		ctx = requisition.NewRequestionContext(context.Background())
	}
	ctx = context.WithValue(context.Background(), share.ReqMetaDataKey, make(map[string]string))
	return ctx, nil
}

/**
 * server token auth
 */
func serverAuth(ctx context.Context, req *protocol.Message, token string) error {

	if tokenInfo, err := parseToken(token, _config.TokenSecret, _config.TokenIssuer); err != nil {
		return err
	} else if tokenInfo.RequestId <= 0 {
		return errors.New("invalid request_id in token")
	}
	return nil
}





