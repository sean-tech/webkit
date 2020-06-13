package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/validate"
	"net"
)

type CmdParams struct {
	EtcdEndPoints 		[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	EtcdConfigPath 		string			`json:"etcd_config_path" validate:"required,gte=1"`
	EtcdConfigUserName 	string			`json:"etcd_config_user_name" validate:"required,gte=1"`
	EtcdConfigPassword 	string			`json:"etcd_config_password" validate:"required,gte=1"`
}

var _params *CmdParams

func CmdEncrypt(params CmdParams, module, salt string) string {
	var err error
	if err = validate.ValidateParameter(params); err != nil {
		panic(err)
	}
	var data []byte
	if data, err = json.Marshal(params); err != nil {
		panic(err)
	}

	var originkey = fmt.Sprintf("%s@%s", module, salt)
	md5Value := encrypt.GetMd5().Encode([]byte(originkey))
	//fmt.Println(md5Value)
	key := generateKey([]byte(md5Value))

	var encryptData []byte
	if encryptData, err = encrypt.GetAes().EncryptCBC(data, key); err != nil {
		panic(err)
	}
	secret := base64.StdEncoding.EncodeToString(encryptData)
	//fmt.Println(secret)
	return secret
}

func cmdDecrypt(secret string, module, salt string) *CmdParams {
	var params = &CmdParams{}
	// encrypt parse
	var encryptData []byte
	var err error
	if encryptData, err = base64.StdEncoding.DecodeString(secret); err != nil {
		panic(err)
	}
	// decrypt
	var decryptData []byte
	var originkey = fmt.Sprintf("%s@%s", module, salt)
	md5Value := encrypt.GetMd5().Encode([]byte(originkey))
	//fmt.Println(md5Value)
	key := generateKey([]byte(md5Value))
	if decryptData, err = encrypt.GetAes().DecryptCBC(encryptData, key); decryptData == nil || err != nil {
		panic(err)
	}
	if err := json.Unmarshal(decryptData, params); err != nil {
		panic(err)
	}

	return params
}

func configEncrypt(cfg *AppConfig, module, salt string) (string, error) {
	if _params == nil {
		panic(errors.New("etcd client init with nil cmd params"))
	}

	var data []byte; var err error
	if data, err = json.Marshal(cfg); err != nil {
		return "", err
	}

	var originkey = fmt.Sprintf("%s/%s@%s", _params.EtcdConfigPath, module, salt)
	md5Value := encrypt.GetMd5().Encode([]byte(originkey))
	//fmt.Println(md5Value)
	key := generateKey([]byte(md5Value))

	var encryptData []byte
	if encryptData, err = encrypt.GetAes().EncryptCBC(data, key); err != nil {
		panic(err)
	}
	secret := base64.StdEncoding.EncodeToString(encryptData)
	//fmt.Println(secret)
	return secret, nil
}

func ConfigDecrypt(secret string, module, salt string) (cfg *AppConfig, err error) {
	if _params == nil {
		panic(errors.New("etcd client init with nil cmd params"))
	}

	var encryptData []byte
	if encryptData, err = base64.StdEncoding.DecodeString(secret); err != nil {
		return nil, err
	}

	var decryptData []byte
	var originkey = fmt.Sprintf("%s/%s@%s", _params.EtcdConfigPath, module, salt)
	md5Value := encrypt.GetMd5().Encode([]byte(originkey))
	//fmt.Println(md5Value)
	key := generateKey([]byte(md5Value))
	if decryptData, err = encrypt.GetAes().DecryptCBC(encryptData, key); decryptData == nil || err != nil {
		return nil, err
	}

	cfg = new(AppConfig)
	if err := json.Unmarshal(decryptData, cfg); err != nil {
		panic(err)
	}
	//fmt.Println(cfg)
	return cfg, nil
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 32)
	copy(genKey, key)
	for i := 32; i < len(key); {
		for j := 0; j < 32 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}


func GetLocalIP() (ips []string){
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