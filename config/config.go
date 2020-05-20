package config

import (
	"time"
)

type AppConfig struct {
	RunMode 				string			`json:"run_mode" validate:"required,oneof=debug test release"`
	WorkerId 				int64			`json:"worker_id" validate:"min=0"`
	HttpPort            	int				`json:"http_port" validate:"required,min=1000"`
	RpcPort               	int				`json:"rpc_port" validate:"required,min=1000"`
	RpcPerSecondConnIdle  	int64			`json:"rpc_per_second_conn_idle" validate:"required,gte=1"`
}

/**
 * 全局server统一配置
 */
type ServerConfig struct {
	// http token
	HttpToken 			*TokenConfig	`json:"http_token" validate:"required"`
	HttpCert			*HttpSecret		`json:"http_cert" validate:"required"`

	// rpc token
	RpcToken 			*TokenConfig	`json:"http_token" validate:"required"`
	// timeout
	TcpReadTimeout      time.Duration	`json:"read_timeout" validate:"required,gte=1"`
	TcpWriteTimeout     time.Duration	`json:"write_timeout" validate:"required,gte=1"`
	// etcd
	EtcdEndPoints 		[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	EtcdRpcBasePath 	string			`json:"etcd_rpc_base_path" validate:"required,gte=1"`
	EtcdRpcUserName 	string			`json:"etcd_rpc_username" validate:"required,gte=1"`
	EtcdRpcPassword 	string			`json:"etcd_rpc_password" validate:"required,gte=1"`
}

type TokenConfig struct {
	TokenSecret     	string      	`json:"token_secret" validate:"required,gte=1"`
	TokenIssuer     	string      	`json:"token_issuer" validate:"required,gte=1"`
	TokenExpiresTime 	time.Duration 	`json:"token_expires_time" validate:"required,gte=1"`
}

type HttpSecret struct {
	ServerPubKey 		string 			`json:"server_pub_key" validate:"required"`
	ServerPriKey		string 			`json:"server_pri_key" validate:"required"`
	ClientPubKey 		string 			`json:"client_pub_key" validate:"required"`
}

type RpcSecret struct {
	CACert       			string 			`json:"ca_cert" validate:"required"`
	CACommonName 			string 			`json:"ca_common_name" validate:"required"`
	ServerCert   			string 			`json:"server_cert" validate:"required"`
	ServerKey    			string 			`json:"server_key" validate:"required"`
}

//type HttpConfig struct {
//	RunMode 			string			`json:"run_mode" validate:"required,oneof=debug test release"`
//	WorkerId 			int64			`json:"worker_id" validate:"min=0"`
//	HttpPort            int				`json:"http_port" validate:"required,min=1,max=10000"`
//	ReadTimeout         time.Duration	`json:"read_timeout" validate:"required,gte=1"`
//	WriteTimeout        time.Duration	`json:"write_timeout" validate:"required,gte=1"`
//	// token
//	TokenSecret      	string        	`json:"token_secret" validate:"required,gte=1"`
//	TokenIssuer      	string        	`json:"token_issuer" validate:"required,gte=1"`
//	TokenExpiresTime 	time.Duration 	`json:"token_expires_time" validate:"required,gte=1"`

//	// secret
//	ServerPubKey 		string 			`json:"server_pub_key"`
//	ServerPriKey 		string 			`json:"server_pri_key"`
//	ClientPubKey 		string 			`json:"client_pub_key"`
//}

//type RpcConfig struct {
//	RunMode string							`json:"run_mode" validate:"required,oneof=debug test release"`
//	RpcPort               	int				`json:"rpc_port" validate:"required,min=1,max=10000"`
//	RpcPerSecondConnIdle  	int64			`json:"rpc_per_second_conn_idle" validate:"required,gte=1"`
//	ReadTimeout           	time.Duration	`json:"read_timeout" validate:"required,gte=1"`
//	WriteTimeout          	time.Duration	`json:"write_timeout" validate:"required,gte=1"`
//	// token
//	TokenSecret      		string        	`json:"token_secret" validate:"required,gte=1"`
//	TokenIssuer      		string        	`json:"token_issuer" validate:"required,gte=1"`
//	// tls
//	CACert       			string 			`json:"ca_cert"`
//	CACommonName 			string 			`json:"ca_common_name"`
//	ServerCert   			string 			`json:"server_cert"`
//	ServerKey    			string 			`json:"server_key"`
//	// etcd
//	EtcdEndPoints 			[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
//	EtcdRpcBasePath 		string			`json:"etcd_rpc_base_path" validate:"required,gte=1"`
//	EtcdRpcUserName 		string			`json:"etcd_rpc_username" validate:"required,gte=1"`
//	EtcdRpcPassword 		string			`json:"etcd_rpc_password" validate:"required,gte=1"`
//}

type MysqlConfig struct{
	Type 		string 			`json:"type" validate:"required,oneof=mysql"`
	User 		string			`json:"user" validate:"required,gte=1"`
	Password 	string			`json:"password" validate:"required,gte=1"`
	Hosts 		map[int]string	`json:"hosts" validate:"required,gte=1,dive,keys,min=0,endkeys,tcp_addr"`
	Name 		string			`json:"name" validate:"required,gte=1"`
	MaxIdle 	int				`json:"max_idle" validate:"required,min=1"`
	MaxOpen 	int				`json:"max_open" validate:"required,min=1"`
	MaxLifetime time.Duration	`json:"max_lifetime" validate:"required,gte=1"`
}

type RedisConfig struct{
	Host        string			`json:"host" validate:"required,tcp_addr"`
	Password    string			`json:"password" validate:"gte=0"`
	MaxIdle     int				`json:"max_idle" validate:"required,min=1"`
	MaxActive   int				`json:"max_active" validate:"required,min=1"`
	IdleTimeout time.Duration	`json:"idle_timeout" validate:"required,gte=1"`
}