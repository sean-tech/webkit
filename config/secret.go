package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/validate"
	"net"
)

type CmdParams struct {
	EtcdEndPoints 		[]string		`json:"etcd_end_points" validate:"required,gte=1,dive,tcp_addr"`
	EtcdConfigPath 		string			`json:"etcd_config_path" validate:"required,gte=1"`
	EtcdConfigUserName 	string
	EtcdConfigPassword 	string
}

func cmdEncrypt(params CmdParams, service_name string, ip string, salt string) string {
	var err error
	if err = validate.ValidateParameter(params); err != nil {
		panic(err)
	}
	var data []byte
	if data, err = json.Marshal(params); err != nil {
		panic(err)
	}

	var originkey = fmt.Sprintf("%s/%s/%s", service_name, ip, salt)
	md5Value := encrypt.GetMd5().Encrypt([]byte(originkey))
	fmt.Println(md5Value)
	key := generateKey([]byte(md5Value))

	var encryptData []byte
	if encryptData, err = encrypt.GetAes().EncryptCBC(data, key); err != nil {
		panic(err)
	}
	secret := base64.StdEncoding.EncodeToString(encryptData)
	fmt.Println(secret)
	return secret
}

func cmdDecrypt(secret string, service_name, salt string) *CmdParams {
	var params = &CmdParams{}

	var ips []string
	if ips = GetLocalIP(); ips == nil {
		panic("local ip got failed")
	}
	fmt.Printf("%+v", ips)

	var encryptData []byte
	var err error
	if encryptData, err = base64.StdEncoding.DecodeString(secret); err != nil {
		panic(err)
	}
	var decryptData []byte
	for _, ip := range ips {
		var originkey = fmt.Sprintf("%s/%s/%s", service_name, ip, salt)
		md5Value := encrypt.GetMd5().Encrypt([]byte(originkey))
		fmt.Println(md5Value)
		key := generateKey([]byte(md5Value))
		if decryptData, err = encrypt.GetAes().DecryptCBC(encryptData, key); decryptData == nil || err != nil {
			continue
		}
		if err := json.Unmarshal(decryptData, params); err == nil {
			break
		}
	}
	if params == nil {
		panic("decrypt failed")
	}
	fmt.Println(params)
	return params
}

func configEncrypt(EtcdConfigPath string, service_name string, ip string) {
	// etcdpath
	// endpoints
	// servicename

	//var originkey = fmt.Sprintf("%s/%s/%s", etcd_endpoints, etcd_config_path, service_name)
}

func configDecrypt(secret string,  service_name string)  {
	
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
		fmt.Println("get ip arr failed: ",err)
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