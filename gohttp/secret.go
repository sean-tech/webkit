package gohttp

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"log"
	"sync"
	"time"
)

type TokenParseFunc func(ctx context.Context, token string) (userId uint64, userName, key string, err error)

type SecretParams struct {
	Secret string	`json:"secret" validate:"required,base64"`
}

type RsaConfig struct {
	ServerPubKey 		string 			`json:"server_pub_key" validate:"required"`
	ServerPriKey		string 			`json:"server_pri_key" validate:"required"`
	ClientPubKey 		string 			`json:"client_pub_key" validate:"required"`
}

type ISecretManager interface {
	InterceptRsa(rsa *RsaConfig) gin.HandlerFunc
	InterceptToken(tokenParse TokenParseFunc) gin.HandlerFunc
	InterceptAes() gin.HandlerFunc
}

var (
	_secretManagerOnce sync.Once
	_secretManager *secretManagerImpl
)

func GetSecretManager() ISecretManager {
	_secretManagerOnce.Do(func() {
		_secretManager = &secretManagerImpl{}
	})
	return _secretManager
}

type secretManagerImpl struct {
}

/**
 * rsa拦截校验
 */
func (this *secretManagerImpl) InterceptRsa(rsa *RsaConfig) gin.HandlerFunc {
	if err := validate.ValidateParameter(rsa); err != nil {
		log.Fatal(err)
	}
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		g.getRequisition().Rsa = rsa

		var code StatusCode = STATUS_CODE_SUCCESS
		var params SecretParams
		var encrypted []byte
		var jsonBytes []byte
		var sign = ctx.GetHeader("sign")
		var signDatas, _ = base64.StdEncoding.DecodeString(sign)

		// params handle
		if err := g.Ctx.Bind(&params); err != nil { // bind
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err := validate.ValidateParameter(params); err != nil { // validate
			code = STATUS_CODE_INVALID_PARAMS
		} else if encrypted, err = base64.StdEncoding.DecodeString(params.Secret); err != nil { // decode
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if jsonBytes, err = encrypt.GetRsa().Decrypt(rsa.ServerPriKey, encrypted); err != nil { // decrypt
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err = encrypt.GetRsa().Verify(rsa.ClientPubKey, jsonBytes, signDatas); err != nil { // sign verify
			code = STATUS_CODE_SECRET_CHECK_FAILED
		}
		// code check
		if code != STATUS_CODE_SUCCESS {
			g.Response(code, code.Msg(),nil, "")
			ctx.Abort()
			return
		}
		g.getRequisition().SecretMethod = secret_method_rsa
		g.getRequisition().Params = jsonBytes
		// next
		ctx.Next()
	}
}

/**
 * token拦截校验
 */
func (this *secretManagerImpl) InterceptToken(tokenParse TokenParseFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		if userId, userName, key, err := tokenParse(ctx, ctx.GetHeader("Authorization")); err != nil {
			g.ResponseError(err)
			ctx.Abort()
			return
		} else {
			foundation.GetRequisition(ctx).UserId = userId
			foundation.GetRequisition(ctx).UserName = userName
			g.getRequisition().Key, _ = hex.DecodeString(key)
			// next
			ctx.Next()
		}
	}
}

/**
 * aes拦截校验
 */
func (this *secretManagerImpl) InterceptAes() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		var code StatusCode = STATUS_CODE_SUCCESS
		var params SecretParams
		var encrypted []byte
		var jsonBytes []byte

		// params handle
		if err := g.Ctx.Bind(&params); err != nil { // bind
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err := validate.ValidateParameter(params); err != nil { // validate
			code = STATUS_CODE_INVALID_PARAMS
		} else if encrypted, err = base64.StdEncoding.DecodeString(params.Secret); err != nil { // decode
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if jsonBytes, err = encrypt.GetAes().DecryptCBC(encrypted, g.getRequisition().Key); err != nil { // decrypt
			code = STATUS_CODE_SECRET_CHECK_FAILED
		}
		// code check
		if code != STATUS_CODE_SUCCESS {
			g.Response(code, code.Msg(),nil, "")
			ctx.Abort()
			return
		}

		g.getRequisition().SecretMethod = secret_method_aes
		g.getRequisition().Params = jsonBytes
		// next
		ctx.Next()
	}
}



/** 存储接口 **/
type ISecretStorage interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string)
}

/**
* 获取内存存储实例
*/
func NewMemeoryStorage() ISecretStorage {
	return new(SecretMemeoryStorageImpl)
}
var _memoryStorage = NewMemeoryStorage()

// 内存存储实现
type SecretMemeoryStorageImpl struct {
	memoryStorageMap sync.Map
}

func (this *SecretMemeoryStorageImpl) Set(key string, value interface{}, expiresTime time.Duration) error {
	this.memoryStorageMap.Store(key, value)
	// 定时删除
	go func(storage *SecretMemeoryStorageImpl, expiresTime time.Duration) {
		select {
		case <- time.After(expiresTime):
			storage.Delete(key)
		}
	}(this, expiresTime)
	return nil
}

func (this *SecretMemeoryStorageImpl) Get(key string) (value string, err error) {
	if tokenInter, ok := this.memoryStorageMap.Load(key); ok {
		return tokenInter.(string), nil
	}
	return "", errors.New("value for key " + key + "not exist")
}

func (this *SecretMemeoryStorageImpl) Delete(key string) {
	this.memoryStorageMap.Delete(key)
}
