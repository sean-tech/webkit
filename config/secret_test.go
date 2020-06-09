package config

import (
	"fmt"
	"testing"
)

func TestCmdParamsEncrypt(t *testing.T) {

}

func TestEnc(t *testing.T) {

	var params = CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	}
	module := "user"
	salt := "Ammzijhwekn!l@20k35"
	ip := "192.168.1.20"
	secret := CmdEncrypt(params, module, ip, salt)
	p := cmdDecrypt(secret, module, salt)
	fmt.Println(p)
}

func TestConfigEnc(t *testing.T) {
	_params = &CmdParams{
		EtcdEndPoints:      []string{"127.0.0.1:2379"},
		EtcdConfigPath:     "/sean-tech/webkit/config",
		EtcdConfigUserName: "root",
		EtcdConfigPassword: "etcd.user.root.pwd",
	}
	module := "user"
	salt := "Ammzijhwekn!l@20k35"
	secret, err := configEncrypt(cfg, module, salt)
	if err != nil {
		t.Error(err)
	}
	if appConfig, err := ConfigDecrypt(secret, module, salt); err != nil {
		t.Error(err)
	} else {
		fmt.Println(appConfig)
	}
}

func TestIpGet(t *testing.T) {
	GetLocalIP()
}