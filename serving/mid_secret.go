package serving

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"sync"
	"time"
)

type TokenInfo struct {
	UserId uint64 			`json:"userId"`
	UserName string 		`json:"userName"`
	jwt.StandardClaims
}


/** 存储接口 **/
type ISecretStorage interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string)
}

func getSecretStorage() ISecretStorage {
	if _httpConfig.SecretStorage != nil {
		return _httpConfig.SecretStorage
	}
	return NewMemeoryStorage()
}

type ISecretManager interface {
	GenerateToken(userId uint64, userName string, isAdministrotor bool, JwtSecret string, JwtIssuer string, JwtExpiresTime time.Duration) (string, error)
	ParseToken(token string, JwtSecret string, JwtIssuer string) (*TokenInfo, error)
	CheckToken(token string, JwtSecret string, JwtIssuer string) error
	GetAesKey(userName string) (key string, err error)
	InterceptToken() gin.HandlerFunc
	InterceptRsa() gin.HandlerFunc
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
 * 生成token
 */
func (this *secretManagerImpl) GenerateToken(userId uint64, userName string, isAdministrotor bool, JwtSecret string, JwtIssuer string, JwtExpiresTime time.Duration) (string, error) {
	expireTime := time.Now().Add(JwtExpiresTime)
	iat := time.Now().Unix()
	//jti := _httpConfig.IdWorker.GetId()
	c := TokenInfo{
		UserId:			userId,
		UserName:       userName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    JwtIssuer,
			//Id:strconv.FormatInt(jti, 10),
			IssuedAt:iat,
			NotBefore: iat,
			Subject:"client",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	token, err := tokenClaims.SignedString([]byte(JwtSecret))
	if err != nil {
		return "", foundation.NewError(STATUS_CODE_AUTH_TOKEN_GENERATE_FAILED, STATUS_MSG_AUTH_TOKEN_GENERATE_FAILED)
	}
	if err := getSecretStorage().Set(userName, token, JwtExpiresTime); err != nil {
		return "", foundation.NewError(STATUS_CODE_AUTH_TOKEN_GENERATE_FAILED, STATUS_MSG_AUTH_TOKEN_GENERATE_FAILED)
	}
	if err := getSecretStorage().Set(userName, hex.EncodeToString(encrypt.GetAes().GenerateKey()), JwtExpiresTime); err != nil {
		return "", foundation.NewError(STATUS_CODE_AUTH_TOKEN_GENERATE_FAILED, STATUS_MSG_AUTH_TOKEN_GENERATE_FAILED)
	}
	return token, nil
}

/**
 * 解析token
 */
func (this *secretManagerImpl) ParseToken(token string, JwtSecret string, JwtIssuer string) (*TokenInfo, error) {
	if token == "" || len(token) == 0 {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_EMPTY, STATUS_MSG_AUTH_CHECK_TOKEN_EMPTY)
	}
	tokenClaims, err := jwt.ParseWithClaims(token, &TokenInfo{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims == nil {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_FAILED, STATUS_MSG_AUTH_CHECK_TOKEN_FAILED)
	}
	if !tokenClaims.Valid {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_FAILED, STATUS_MSG_AUTH_CHECK_TOKEN_FAILED)
	}
	claims, ok := tokenClaims.Claims.(*TokenInfo)
	if !ok {
		return nil, foundation.NewError(STATUS_CODE_AUTH_TYPE_ERROR, STATUS_MSG_AUTH_TYPE_ERROR)
	}
	savedToken, err := getSecretStorage().Get(claims.UserName)
	if err != nil {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_FAILED, STATUS_MSG_AUTH_CHECK_TOKEN_FAILED)
	}
	if savedToken != token {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_FAILED, STATUS_MSG_AUTH_CHECK_TOKEN_FAILED)
	}
	if claims.Issuer != JwtIssuer {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_FAILED, STATUS_MSG_AUTH_CHECK_TOKEN_FAILED)
	}
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, foundation.NewError(STATUS_CODE_AUTH_CHECK_TOKEN_TIMEOUT, STATUS_MSG_AUTH_CHECK_TOKEN_TIMEOUT)
	}
	return claims, nil
}

func (this *secretManagerImpl) CheckToken(token string, JwtSecret string, JwtIssuer string) error {
	if _, err := this.ParseToken(token, JwtSecret, JwtIssuer); err != nil {
		return err
	}
	return nil
}

func (this *secretManagerImpl) GetAesKey(userName string) (key string, err error) {
	return getSecretStorage().Get(userName)
}

/**
 * jwt拦截校验
 */
func (this *secretManagerImpl) InterceptToken() gin.HandlerFunc {
	handler := func(ctx *gin.Context) {
		g := Gin{ctx}
		// parse & check
		tokenInfo, err := this.ParseToken(ctx.GetHeader("Authorization"), _httpConfig.JwtSecret, _httpConfig.JwtIssuer)
		if err != nil {
			g.ResponseError(err)
			ctx.Abort()
			return
		}
		foundation.GetRequisition(ctx).UserId = tokenInfo.UserId
		foundation.GetRequisition(ctx).UserName = tokenInfo.UserName
		// next
		ctx.Next()
	}
	return handler
}



type SecretParams struct {
	Secret string	`json:"secret" validate:"required,base64"`
} 

/**
 * rsa拦截校验
 */
func (this *secretManagerImpl) InterceptRsa() gin.HandlerFunc {
	handler := func(ctx *gin.Context) {
		if _httpConfig.SecretOpen == false {
			ctx.Next()
			return
		}

		g := Gin{ctx}
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
		} else if jsonBytes, err = encrypt.GetRsa().Decrypt(_httpConfig.ServerPriKey, encrypted); err != nil { // decrypt
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err = encrypt.GetRsa().Verify(_httpConfig.ClientPubKey, jsonBytes, signDatas); err != nil { // sign verify
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
	return handler
}

/**
 * aes拦截校验
 */
func (this *secretManagerImpl) InterceptAes() gin.HandlerFunc {
	handler := func(ctx *gin.Context) {
		if _httpConfig.SecretOpen == false {
			ctx.Next()
			return
		}

		g := Gin{ctx}
		var code StatusCode = STATUS_CODE_SUCCESS
		var params SecretParams
		var key string
		var keyBytes []byte
		var encrypted []byte
		var jsonBytes []byte

		// params handle
		if err := g.Ctx.Bind(&params); err != nil { // bind
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err := validate.ValidateParameter(params); err != nil { // validate
			code = STATUS_CODE_INVALID_PARAMS
		} else if key, err = getSecretStorage().Get(foundation.GetRequisition(ctx).UserName); err != nil { // get key
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if encrypted, err = base64.StdEncoding.DecodeString(params.Secret); err != nil { // decode
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if keyBytes, err = hex.DecodeString(key); err != nil { // get key bytes
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if jsonBytes, err = encrypt.GetAes().DecryptCBC(encrypted, keyBytes); err != nil { // decrypt
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
		g.getRequisition().Key = keyBytes
		// next
		ctx.Next()
	}
	return handler
}




/**
* 获取内存存储实例
*/
func NewMemeoryStorage() ISecretStorage {
	return new(SecretMemeoryStorageImpl)
}

// 内存存储实现
type SecretMemeoryStorageImpl struct {
	memoryStorageMap sync.Map
}

func (this *SecretMemeoryStorageImpl) Set(key string, value interface{}, expiresTime time.Duration) error {
	this.memoryStorageMap.Store(key, value)
	// 定时删除
	select {
	case <- time.After(expiresTime):
		this.Delete(key)
	}
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
