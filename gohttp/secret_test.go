package gohttp

import (
	"fmt"
	"testing"
	"time"
)

func TestJwt(t *testing.T) {
	//default_file_path := "../../web_config.ini"
	//config.SetupFromLocal(default_file_path)
	//logging.Setup()
	//database.Setup(config.Mysql, config.Redis)
	//Setup(database.NewRedisManager(config.Redis))
	//token, err := GetJwtManager().GenerateToken(102899123, "yang", "yangpwd123", false)
	//if err != nil {
	//	t.Error(err)
	//}
	//key, err := GetSecretManager().GetAesKey(token)
	//if err != nil {
	//	t.Error(err)
	//}
	//fmt.Println(token + "-----------" + key)
}


type iTokenManager interface {
	GenerateToken(userId uint64, userName string, JwtSecret string, JwtIssuer string, JwtExpiresTime time.Duration) (string, error)
	CheckToken(token string, JwtSecret string, JwtIssuer string) error
}
func TestToken(t *testing.T) {
	var secret = "ahsjdadusba"
	var issuer = "sean.test"
	var tokenMgr iTokenManager = GetSecretManager()

	var expires = 30 * time.Second
	token, err := tokenMgr.GenerateToken(1230090123, "seantest1", secret, issuer, expires)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("token generate success!---" + token)

	if err := tokenMgr.CheckToken(token, secret, issuer); err != nil {
		t.Error(err)
		return
	}
	fmt.Println("token check success!")
}