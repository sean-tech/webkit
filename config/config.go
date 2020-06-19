package config

import (
	"errors"
	"flag"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"os"
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
 * 初始化config
 * configFilePath, cmd:-debugconfig 通过local json 文件加载配置
 * configEtcdInfo, cmd:-configetcdinfo 通过etcd注册中心加载配置
 */
func Setup(module, salt string, debugConfig *AppConfig, etcdConfig string, load ConfigLoad) {

	// config file path
	debug_config_usage := "please use -debugconfig, ture or false to pointing at whether use local debug config."
	debug_config_use := flag.Bool("debugconfig", true, debug_config_usage)
	// etcdinfo
	etcd_config_usage := "please use -etcdconfig to pointing at etcd config info secreted."
	etcd_config := flag.String("etcdconfig", etcdConfig, etcd_config_usage)
	// parse
	flag.Parse()

	// when etcdinfo set, etcd client init
	if etcd_config != nil && *etcd_config != "" {
		params := cmdDecrypt(*etcd_config, module, salt)
		clientInit(params)
	}
	// load config from local debug config
	if *debug_config_use == true {
		if debugConfig == nil {
			panic("debug config is nil when debugconfig used true")
		}
		if err := debugConfig.Validate(); err != nil {
			panic("debug config validate error:" + err.Error())
		}
		os.Stdout.Write([]byte("config load success with debug config.\n"))
		load(debugConfig)
		return
	}
	// load config from etcd
	if etcd_config == nil  || *etcd_config == "" {
		panic("please use -configfilepath or -configetcdinfo to pointing at config load method")
	}
	os.Stdout.Write([]byte("config load success with etcd config.\n"))
	load(configLoad(module, salt))
}

func configLoad(module, salt string) *AppConfig {
	var ips []string
	if ips = GetLocalIP(); ips == nil {
		panic("local ip got failed")
	}
	var workerId int64 = -1
	for _, ip := range ips {
		var err error
		if workerId, err = GetWorkerId(module, ip); err != nil {
			continue
		}
	}
	if workerId == -1 {
		panic("load workerid failed")
	}

	if appConfig, err := GetConfig(module, salt); err != nil {
		panic(err)
	} else {
		appConfig.Http.WorkerId = workerId
		appConfig.Mysql.WorkerId = workerId
		return appConfig
	}
}
