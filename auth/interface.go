package auth

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"strconv"
	"sync"
	"time"
)

type TokenItem struct {
	Id 			string		`json:"id"`
	ExpiresAt 	int64  		`json:"expires_at"`
	UUID 		string 		`json:"uuid"`
	SignedId	string		`json:"signed_id"`
	UserId 		uint64 		`json:"userId"`
	UserName 	string 		`json:"userName"`
	Token   	string		`json:"token"`
	Key     	string		`json:"key"`
	Client  	string		`json:"client"`
}

type NewAuthParameter struct {
	AuthCode 	string		`json:"auth_code" validate:"required,gte=1"`
	UUID 		string		`json:"uuid" validate:"required,gte=1"`
	UserId 		uint64		`json:"userid" validate:"required,gte=1"`
	UserName 	string		`json:"username" validate:"required,gte=1"`
	Client 		string		`json:"client" validate:"required,gte=1"`
}
type AuthRefreshParameter struct {
	RefreshToken 	string	`json:"refresh_token" validate:"required,gte=1"`
	AccessToken 	string	`json:"access_token" validate:"required,gte=1"`
}
type AuthResult struct {
	RefreshToken 	string	`json:"refresh_token" validate:"required,gte=1"`
	Key 			string	`json:"key" validate:"required,gte=1"`
	AccessToken 	string	`json:"access_token" validate:"required,gte=1"`
}

type AccessTokenAuthParameter struct {
	AccessToken 	string	`json:"access_token" validate:"required,gte=1"`
}



type IAuthApi interface {
	NewAuth(ctx *gin.Context)
	AuthRefresh(ctx *gin.Context)
	AccessTokenAuth(ctx *gin.Context)
}

type IAuthService interface {
	NewAuth(ctx context.Context, parameter *NewAuthParameter, result *AuthResult) error
	AuthRefresh(ctx context.Context, parameter *AuthRefreshParameter, result *AuthResult) error
	AccessTokenAuth(ctx context.Context, parameter *AccessTokenAuthParameter, accessTokenItem *TokenItem) error
}

type iAuthDao interface {
	SaveRefreshTokenItem(userName string, tokenItem *TokenItem) error
	GetRefreshTokenItem(userName string) (tokenItem *TokenItem, err error)
	GetKey(userName string) (key string, err error)
	DeleteRefreshTokenItem(userName string) error
	SaveAccessTokenItem(userName string, tokenItem *TokenItem) error
	GetAccessTokenItem(userName string) (tokenItem *TokenItem, err error)
	DeleteAccessTokenItem(userName string) error
	DeleteAllTokenItem() error
}

var (
	_api     IAuthApi
	_apiOnce     sync.Once
	_service IAuthService
	_serviceOnce sync.Once
	_dao     iAuthDao
	_daoOnce     sync.Once
)

func Api() IAuthApi {
	_apiOnce.Do(func() {
		_api = new(apiImpl)
	})
	return _api
}

func Service() IAuthService {
	_serviceOnce.Do(func() {
		_service = &serviceImpl{
			authCode: "this is auth code for validate",
		}
	})
	return _service
}

func dao() iAuthDao {
	_daoOnce.Do(func() {
		_dao = new(daoImpl)
	})
	return _dao
}



type IAuthStorage interface {
	// hash table set & get
	HashExists(key, field string) (bool, error)
	HashLen(key string) (int64, error)
	HashSet(key string, values ...interface{}) error
	HashGet(key, field string) (string, error)
	HashMSet(key string, values ...interface{}) error
	HashMGet(key string, fields ...string) ([]interface{}, error)
	HashDelete(key string, fields ...string) error
	HashKeys(key string) ([]string, error)
	HashVals(key string) ([]string, error)
	HashGetAll(key string) (map[string]string, error)
}

type AuthConfig struct {
	WorkerId 					int64			`json:"worker_id" validate:"min=0"`
	TokenSecret      			string        	`json:"token_secret" validate:"required,gte=1"`
	TokenIssuer      			string        	`json:"token_issuer" validate:"required,gte=1"`
	RefreshTokenExpiresTime 	time.Duration 	`json:"refresh_token_expires_time" validate:"required,gte=1"`
	AccessTokenExpiresTime 		time.Duration 	`json:"access_token_expires_time" validate:"required,gte=1"`
}

var (
	_config AuthConfig
	_idWorker foundation.SnowId
	_storage IAuthStorage
)

func Setup(config AuthConfig, storage IAuthStorage)  {
	if err := validate.ValidateParameter(config); err != nil {
		panic(err)
	}
	_config = config
	var err error
	if _idWorker, err = foundation.NewWorker(config.WorkerId); err != nil {
		panic("snowid worker init failed with id : " + strconv.FormatInt(config.WorkerId, 10))
	}
	if storage == nil {
		panic("auth storage is nil")
	}
	_storage = storage
}



const (
	// jwt token
	status_code_auth_code_wrong           = 811
	status_code_auth_token_empyt          = 801
	status_code_auth_token_checkfaild     = 802
	status_code_auth_token_timeout        = 803
	status_code_auth_token_generatefailed = 804
	status_code_auth_token_savefailed 	  = 805
	status_code_auth_token_typewrong      = 806
	status_code_auth_token_otherlogin     = 807
	status_code_auth_token_shouldrefresh  = 808
	status_code_auth_shouldnot_refresh 	  = 809
)

const (
	status_msg_auth_code_wrong           = "auth code 验证失败"
	status_msg_auth_token_empyt          = "Token为空，如未登录，请先登录"
	status_msg_auth_token_checkfaild     = "Token校验失败"
	status_msg_auth_token_timeout        = "Token已过期"
	status_msg_auth_token_generatefailed = "Token生成失败"
	status_msg_auth_token_savefailed	 = "Token持久化失败"
	status_msg_auth_token_typewrong      = "Token校验类型错误"
	status_msg_auth_token_otherlogin     = "您的账户已在其他设备登录，请注意账户安全"
	status_msg_auth_token_shouldrefresh  = "access token should refresh"
	status_msg_auth_shouldnot_refresh	 = "access token should not refresh"
)
