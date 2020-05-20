package config

import "testing"

func TestEnc(t *testing.T) {
	params := CmdParams{
		EtcdEndPoints:  []string{"127.0.0.1:2379", "127.0.0.1:32379"},
		EtcdConfigPath: "sean.tech/webkit/config/",
	}
	service_name := "user"
	ip := "192.168.1.20"
	salt := "Ammzijhwekn!l@20k35"
	secret := cmdEncrypt(params, service_name, ip, salt)
	cmdDecrypt(secret, service_name, salt)
}

func TestIpGet(t *testing.T) {
	GetLocalIP()
}