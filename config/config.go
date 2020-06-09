package config

import (
	"errors"
	"flag"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/database"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"log"
)

type AppConfig struct {
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
		if cfg.Http.RunMode != cfg.Rpc.RunMode {
			return errors.New("runmode is not equal between http in rpc")
		}
	}
	return nil
}



type ConfigLoad func(appConfig *AppConfig)

/**
 * 初始化config，通过etcd注册中心
 */
func Setup(module, salt string, debugConfig *AppConfig, debugCmd *CmdParams, load ConfigLoad) {
	// runmode
	runmode_usage := "please use -runmode to pointing at runmode, incase 'debug', 'test' and 'release'."
	runmode := flag.String("runmode", "debug", runmode_usage)
	// secret
	secret_usage := "please use -secret to pointing at etcd secret info."
	secret := flag.String("secret", "", secret_usage)
	// parse
	flag.Parse()

	// parse value validate
	switch *runmode {
	case foundation.RUN_MODE_DEBUG:
		if debugCmd == nil {
			panic("cmd params is nil in debug env")
		}
		_params = debugCmd
		clientInit()
		log.Println("config load success in ", *runmode)
		load(debugConfig)

	case foundation.RUN_MODE_TEST:fallthrough
	case foundation.RUN_MODE_RELEASE:
		if secret == nil || *secret == "" {
			panic("secret for etcd cmd params is nil in test or release runmode")
		}
		_params = cmdDecrypt(*secret, module, salt)
		log.Println("config loading...")
		cfg := configLoad(module, salt)
		if *runmode != cfg.Http.RunMode {
			panic("runmode in cmd is not equal with config")
		}
		log.Println("config load success in ", *runmode)
		load(cfg)

	default:
		panic("runmode is wrong, not incase 'debug', 'test' or 'release'")
	}
}

func configLoad(module, salt string) (appConfig *AppConfig) {

	clientInit()

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
	}
	return appConfig
}
