package config

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/sean-tech/gokit/fileutils"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"io/ioutil"
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
 * configFilePath, cmd:-configfilepath 通过local json 文件加载配置
 * configEtcdInfo, cmd:-configetcdinfo 通过etcd注册中心加载配置
 */
func Setup(module, salt string, configFilePath string, configEtcdInfo string, load ConfigLoad) {

	// config file path
	config_file_path_usage := "please use -configfilepath to pointing at local config file."
	config_file_path := flag.String("configfilepath", configFilePath, config_file_path_usage)
	// etcdinfo
	config_etcd_info_usage := "please use -configetcdinfo to pointing at etcd config info secreted."
	config_etcd_info := flag.String("configetcdinfo", configEtcdInfo, config_etcd_info_usage)
	// parse
	flag.Parse()

	// when etcdinfo set, etcd client init
	if config_etcd_info != nil && *config_etcd_info != "" {
		params := cmdDecrypt(*config_etcd_info, module, salt)
		clientInit(params)
	}

	// load config from local json file
	if config_file_path != nil && *config_file_path != "" && fileutils.CheckExist(*config_file_path) == true {
		var appConfig = new(AppConfig)
		if jsonBytes, err := ioutil.ReadFile(*config_file_path); err != nil {
			panic(err)
		} else if  err := json.Unmarshal(jsonBytes, appConfig); err != nil {
			panic(err)
		} else {
			load(appConfig)
			return
		}
	}

	// load config from etcd with etcd info
	if config_etcd_info == nil  || *config_etcd_info == "" {
		panic("please use -configfilepath or -configetcdinfo to pointing at config load method")
	}
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
