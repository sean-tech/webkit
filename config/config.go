package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sean-tech/gokit/fileutils"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"log"
	"net/rpc"
	"os"
	"strings"
)

type AppConfig struct {
	Log 	*logging.LogConfig
	Http 	*gohttp.HttpConfig     	`json:"http" validate:"required"`
	Rpc 	*gorpc.RpcConfig       	`json:"rpc" validate:"required"`
	Mysql 	*database.MysqlConfig	`json:"mysql" validate:"required"`
	Redis 	*database.RedisConfig 	`json:"redis" validate:"required"`
	CE		*ConfigEtcd				`json:"ce" validate:"required"`
}

type ConfigEtcd struct {
	EtcdEndPoints 			[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	EtcdConfigBasePath 		string			`json:"etcd_rpc_base_path" validate:"required,gte=1"`
	EtcdConfigUserName 		string			`json:"etcd_rpc_username" validate:"required,gte=1"`
	EtcdConfigPassword 		string			`json:"etcd_rpc_password" validate:"required,gte=1"`
}

func (cfg *AppConfig) Validate() error {
	if err := validate.ValidateParameter(cfg); err != nil {
		return err
	}
	if cfg.Http.RsaOpen {
		if err := validate.ValidateParameter(cfg.Http.Rsa); err != nil {
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
* command start - load config
* module, 服务程序所属模块，程序写死不可更改
* httpPort,  cmd:-httpport  服务http端口号
* rpcPort,   cmd:-rpcport   服务rpc端口号
* logPath,   cmd:-logpath   日志输出文件路径（/结尾）
* localConfig，本地配置，当未指定cc时默认加载
* load, 配置加载成功后回调
 */
func Setup(module string, httpPort, rpcPort int, logPath string, localConfig *AppConfig, load ConfigLoad) {
	setup(module, httpPort, rpcPort, logPath, "", localConfig, load)
}

/**
* command start - load config
* module, 服务程序所属模块，程序写死不可更改
* httpPort,  cmd:-httpport  服务http端口号
* rpcPort,   cmd:-rpcport   服务rpc端口号
* logPath,   cmd:-logpath   日志输出文件路径（/结尾）
* ccAddress, cmd:-ccaddress 配置授权中心cca rpc地址
* localConfig，本地配置，当未指定cc时默认加载
* load, 配置加载成功后回调
*/
func setup(module string, httpPort, rpcPort int, logPath, ccAddress string,  localConfig *AppConfig, load ConfigLoad) {

	runmode_usage := "please use -runmode to pointing at runenv in:debug,test,release."
	runmode := flag.String("runmode", foundation.RUN_MODE_RELEASE, runmode_usage)

	httpport_usage := "please use -httpport to pointing at port of http."
	httpport := flag.Int("httpport", httpPort, httpport_usage)

	rpcport_usage := "please use -rpcport to pointing at port of rpc."
	rpcport := flag.Int("rpcport", rpcPort, rpcport_usage)

	logpath_usage := "please use -logpath to pointing at log save path."
	logpath := flag.String("logpath", logPath, logpath_usage)

	ccaddress_usage := "please use -ccaddress to pointing at configcenter rpc address."
	ccaddress := flag.String("ccaddress", ccAddress, ccaddress_usage)

	flag.Parse()

	*runmode = strings.ReplaceAll(*runmode, " ", "")
	*logpath = strings.ReplaceAll(*logpath, " ", "")
	*ccaddress = strings.ReplaceAll(*ccaddress, " ", "")

	// runmode validate
	if runmode == nil || *runmode == "" {
		panic(runmode_usage)
	}
	if *runmode != foundation.RUN_MODE_DEBUG && *runmode != foundation.RUN_MODE_TEST && *runmode != foundation.RUN_MODE_RELEASE {
		panic("runmode not right," + runmode_usage)
	}
	os.Stdout.Write([]byte(fmt.Sprintf("app in %s starting...\n", *runmode)))

	// port validate
	if *httpport < 1 && *httpport > 10000 {
		panic("http port should set between 1 and 10000")
	}
	if *rpcport < 1 && *rpcport > 10000 {
		panic("rpc port should set between 1 and 10000")
	}
	os.Stdout.Write([]byte(fmt.Sprintf("http port %d, rpc port %d starting...\n", *httpport, *rpcport)))

	// logpath validate
	if logpath == nil || *logpath == "" {
		panic(logpath_usage)
	}
	if strings.HasSuffix(*logpath, "/") == false {
		*logpath = *logpath + "/"
	}
	if fileutils.CheckExist(*logpath) == false {
		panic("log file save path error:" + logpath_usage)
	}

	// appcfg load
	var appcfg *AppConfig; var loadinfo string
	if ccaddress != nil && *ccaddress != "" { // load config from config center，when ccaddress set
		appcfg = configLoadFromCenter(*ccaddress, module)
		loadinfo = "config load from configcenter finished.\n"
	} else if localConfig == nil {
		panic(ccaddress_usage)
	} else { // load config from local debug config
		appcfg = localConfig
		loadinfo = "config load from local finished.\n"
	}

	// appcfg set & validate
	appcfg.Http.RunMode = *runmode
	appcfg.Http.HttpPort = *httpport
	appcfg.Rpc.RunMode = *runmode
	appcfg.Rpc.RpcPort = *rpcport
	appcfg.Log = &logging.LogConfig{
		RunMode:     *runmode,
		LogSavePath: *logpath,
		LogPrefix:   module,
	}
	if err := appcfg.Validate(); err != nil {
		panic("config validate error:" + err.Error())
	}
	os.Stdout.Write([]byte(loadinfo))
	load(appcfg)
}

func configLoadFromCenter(ccaddress, module string) *AppConfig {
	os.Stdout.Write([]byte("config loading from configcenter...\n"))
	client, err := rpc.Dial("tcp", ccaddress)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var worker = &Worker{
		Module: module,
		Ip:     GetLocalIP(),
	}
	var config = new(AppConfig)
	err = client.Call(ConfigCenterServiceName+ConfigLoadMethodName, worker, config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}