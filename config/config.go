package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
)

type AppConfig struct {
	Log 	*logging.LogConfig		`json:"log" validate:"required"`
	RsaOpen bool					`json:"rsa_open"`
	Rsa 	*gohttp.RsaConfig		`json:"rsa"`
	Http 	*gohttp.HttpConfig		`json:"http" validate:"required"`
	Rpc 	*gorpc.RpcConfig		`json:"rpc" validate:"required"`
	Mysql 	*database.MysqlConfig	`json:"mysql" validate:"required"`
	Redis 	*database.RedisConfig	`json:"redis" validate:"required"`
}

func (cfg *AppConfig) Validate() error {
	if err := validate.ValidateParameter(cfg); err != nil {
		return err
	}
	if cfg.RsaOpen {
		if err := validate.ValidateParameter(cfg.Rsa); err != nil {
			return err
		}
	}
	if cfg.Rpc.TlsOpen {
		if err := validate.ValidateParameter(cfg.Rpc.Tls); err != nil {
			return err
		}
	}
	if cfg.Http != nil && cfg.Rpc != nil {
		if cfg.Log.RunMode != cfg.Http.RunMode {
			return errors.New("runmode is not equal between log in http")
		}
		if cfg.Http.RunMode != cfg.Rpc.RunMode {
			return errors.New("runmode is not equal between http in rpc")
		}
	}
	return nil
}



type ConfigLoad func(appConfig *AppConfig)

/**
* command start * load config
* module, 服务程序所属模块
* testAddress, cmd:-ccaddress 测试cc rpc地址
* debugConfig，本地debug配置，当未指定cc时默认加载debug配置
*/
func Setup(module string, testAddress string, debugConfig *AppConfig, load ConfigLoad) {

	ccaddress_usage := "please use -ccaddress to pointing at configcenter rpc address."
	ccaddress := flag.String("ccaddress", testAddress, ccaddress_usage)
	flag.Parse()

	// when ccaddress set, load config from config center
	if ccaddress != nil && *ccaddress != "" {
		os.Stdout.Write([]byte("config load success from configcenter.\n"))
		load(ConfigCenterLoading(*ccaddress, module))
		return
	}
	// load config from local debug config
	if debugConfig == nil {
		panic(ccaddress_usage)
	}
	if err := debugConfig.Validate(); err != nil {
		panic("debug config validate error:" + err.Error())
	}
	os.Stdout.Write([]byte("config load success from local debugconfig.\n"))
	load(debugConfig)
}



const ConfigCenterServiceName = "sean,tech/webkit/configcenter"
const ConfigLoadMethodName = ".ConfigLoad"

type IConfigCenter interface {
	ConfigLoad(path string, config *AppConfig) error
}

func ConfigCernterServing(cc IConfigCenter, port int, whitelistips []string) {
	rpc.RegisterName(ConfigCenterServiceName, cc)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}
		var clientIp = conn.RemoteAddr().String()
		if ip, _, err := net.SplitHostPort(strings.TrimSpace(conn.RemoteAddr().String())); err == nil {
			clientIp = ip
		}
		if WhiteListIpsFitter(clientIp, whitelistips) == true {
			go rpc.ServeConn(conn)
		} else {
			conn.Close()
		}
	}
}

func WhiteListIpsFitter(clientIp string, whitelistips []string) bool {
	if whitelistips == nil {
		return true
	}
	for _, ip := range whitelistips {
		if clientIp == ip {
			return true
		}
	}
	return false
}

func ConfigCenterLoading(ccaddress, module string) *AppConfig {
	client, err := rpc.Dial("tcp", ccaddress)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var path = fmt.Sprintf("%s/%s", module, GetLocalIP())
	var config = new(AppConfig)
	err = client.Call(ConfigCenterServiceName+ConfigLoadMethodName, path, config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}


func GetIPs() (ips []string){
	addrs,err := net.InterfaceAddrs()
	if err != nil{
		//fmt.Println("get ip arr failed: ",err)
		return nil
	}
	for _,addr := range addrs{
		if ipnet,ok := addr.(*net.IPNet);ok && !ipnet.IP.IsLoopback(){
			if ipnet.IP.To4() != nil{
				ips = append(ips,ipnet.IP.String())
			}
		}
	}
	return ips
}

func GetLocalIP() string {
	addrs,err := net.InterfaceAddrs()
	if err != nil{
		return ""
	}
	for _,addr := range addrs{
		if ipnet,ok := addr.(*net.IPNet);ok && !ipnet.IP.IsLoopback(){
			if isLocal(ipnet.IP.To4()) {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func isLocal(ip4 net.IP) bool {
	if ip4 == nil {
		return false
	}
	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}




