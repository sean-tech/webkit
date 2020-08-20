package webkit

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"github.com/sean-tech/webkit/auth"
	"github.com/sean-tech/webkit/config"
	"github.com/sean-tech/webkit/gohttp"
	"github.com/sean-tech/webkit/gorpc"
	"github.com/sean-tech/webkit/logging"
	"github.com/smallnest/rpcx/server"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"
	"testing"
)

const SERVICE_USER = "User"
const SERVICE_AUTH = "Auth"
const AUTH_METHOD_NEW_AUTH = "NewAuth"
const AUTH_METHOD_ACCESSTOKEN_AUTH = "AccessTokenAuth"
const AUTH_CODE = "this is auth code for validate"

type User struct {
	UserId uint64 `json:"user_id"`
	UserName string	`json:"username"`
}

type UserInfo struct {
	*User
	*auth.AuthResult
}

type UserLoginParameter struct {
	UserName string	`json:"username" validate:"required,gte=1"`
	Password string	`json:"password" validate:"required,md5"`
}

type UserGetParameter struct {
	UserId uint64 `json:"user_id" validate:"required,min=1"`
}

type IUserApi interface {
	UserLogin(ctx *gin.Context)
	UserGet(ctx *gin.Context)
}

type IUserService interface {
	UserLogin(ctx context.Context, parameter *UserLoginParameter, userInfo *UserInfo) error
	UserGet(ctx context.Context, parameter *UserGetParameter, user *User) error
}

type iUserDao interface {
	UserGetByUserNameAndPassword(userName, password string) (user *User, err error)
	UserGetByUserId(userId uint64) (user *User, err error)
}

var (
	_api         IUserApi
	_apiOnce     sync.Once
	_service     IUserService
	_serviceOnce sync.Once
	_dao         iUserDao
	_daoOnce     sync.Once
)

func Api() IUserApi {
	_apiOnce.Do(func() {
		_api = new(userApiImpl)
	})
	return _api
}

func Service() IUserService {
	_serviceOnce.Do(func() {
		_service = new(userServiceImpl)
	})
	return _service
}

func dao() iUserDao {
	_daoOnce.Do(func() {
		_dao = new(userDaoImpl)
	})
	return _dao
}

type userApiImpl struct {
}

func (this *userApiImpl) UserLogin(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter UserLoginParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var userInfo = new(UserInfo)
	if err := Service().UserLogin(ctx, &parameter, userInfo); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(userInfo)
}

func (this *userApiImpl) UserGet(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter UserGetParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var user = new(User)
	if err := Service().UserGet(ctx, &parameter, user); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(user)
}

type userServiceImpl struct {

}

func (this *userServiceImpl) UserLogin(ctx context.Context, parameter *UserLoginParameter, userInfo *UserInfo) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	if model_user, err := dao().UserGetByUserNameAndPassword(parameter.UserName, parameter.Password); err != nil {
		return err
	} else {
		userInfo.User = model_user
	}
	var authParameter = &auth.NewAuthParameter{
		AuthCode: AUTH_CODE,
		UUID:     "ajsknzjcbjqnjbfdjand",
		UserId:   userInfo.UserId,
		UserName: userInfo.UserName,
		Client:   "iOS",
	}
	var authResult = new(auth.AuthResult)
	if err := gorpc.Call(SERVICE_AUTH, ctx, AUTH_METHOD_NEW_AUTH, authParameter, authResult); err != nil {
		return err
	}
	userInfo.AuthResult = authResult
	return nil
}

func (this *userServiceImpl) UserGet(ctx context.Context, parameter *UserGetParameter, user *User) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	if model_user, err := dao().UserGetByUserId(parameter.UserId); err != nil {
		return err
	} else {
		*user = *model_user
	}
	return nil
}

type userDaoImpl struct {
}

func (this *userDaoImpl) UserGetByUserNameAndPassword(userName, password string) (user *User, err error) {
	return &User{
		UserId:   1230090101,
		UserName: "zhao",
	}, nil
}

func (this *userDaoImpl) UserGetByUserId(userId uint64) (user *User, err error) {
	return &User{
		UserId:   userId,
		UserName: "zhao",
	}, nil
}



func TestUserServer(t *testing.T) {
	// concurrent
	runtime.GOMAXPROCS(runtime.NumCPU())
	// config
	config.Setup("wktest", "user", 9022, 9021, "/Users/sean/Desktop/", testconfig, func(appConfig *config.AppConfig) {
		// log start
		logging.Setup(*appConfig.Log)
		// database start
		//database.SetupRedis(*appConfig.Redis).Open()
		// service start
		gorpc.ServerServe(*appConfig.Rpc, logging.Logger(), RegisterService)
		// server start
		gohttp.HttpServerServe(*appConfig.Http, logging.Logger(), RegisterApi)
	})
}

func RegisterService(server *server.Server)  {
	server.RegisterName(SERVICE_USER, Service(), "")
}

func RegisterApi(engine *gin.Engine)  {

	//serverPubKeyBytes, _ := ioutil.ReadFile("/Users/lyra/Desktop/Doc/安全方案/businessS/spubkey.pem")
	//serverPriKeyBytes, _ := ioutil.ReadFile("/Users/lyra/Desktop/Doc/安全方案/businessS/sprivkey.pem")
	//clientPubKeyBytes, _ := ioutil.ReadFile("/Users/lyra/Desktop/Doc/安全方案/businessC/cpubkey.pem")
	//var rsaHandler = gohttp.SecretManager().InterceptRsa(&gohttp.RsaConfig{
	//	ServerPubKey:     string(serverPubKeyBytes),
	//	ServerPriKey:     string(serverPriKeyBytes),
	//	ClientPubKey:     string(clientPubKeyBytes),
	//})

	var tokenHandler =  gohttp.InterceptToken(func(ctx context.Context, token string) (userId uint64, userName, role, key string, err error) {
		var parameter = &auth.AccessTokenAuthParameter{
			AccessToken: token,
		}
		var accessTokenItem = new(auth.TokenItem)
		if err := gorpc.Call(SERVICE_AUTH, ctx, AUTH_METHOD_ACCESSTOKEN_AUTH, parameter, accessTokenItem); err != nil {
			return 0, "", "", "", err
		}
		return accessTokenItem.UserId, accessTokenItem.UserName, accessTokenItem.Role, accessTokenItem.Key, nil
	})

	apiv1 := engine.Group("api/v1/user/")
	{
		apiv1.POST("login", Api().UserLogin)
		apiv1.POST("get", tokenHandler, gohttp.InterceptAes(), Api().UserGet)
	}
}



func TestClientUserCall(t *testing.T) {

	fmt.Println("--------------login----------------")
	var url = "http://localhost:9022/api/v1/user/login"
	var parameter = map[string]interface{}{
		"username" : "sean",
		"password" : encrypt.GetMd5().Encode([]byte("sean.pwd123")),
	}
	jsonStr, err := json.Marshal(parameter)
	if err != nil {
		fmt.Printf("to json error:%v\n", err)
		return
	}
	var resp map[string]interface{}


	if resp, err = post(url, jsonStr, ""); err == nil {
		resp, _ = resp["data"].(map[string]interface{})

		fmt.Println("--------------getuserinfo----------------")
		url = "http://localhost:9022/api/v1/user/get"
		parameter = map[string]interface{}{
			"user_id" : resp["user_id"],
		}
		jsonStr, err = json.Marshal(parameter)
		if err != nil {
			fmt.Printf("to json error:%v\n", err)
			return
		}

		var key, _ = hex.DecodeString(resp["key"].(string))
		secretData, err := encrypt.GetAes().EncryptCBC(jsonStr, key)
		if err != nil {
			t.Error(err)
		}
		parameter := map[string]string{"secret" : base64.StdEncoding.EncodeToString(secretData)}
		//if encryptData, err := base64.StdEncoding.DecodeString(parameter["secret"]); err != nil {
		//	t.Error(err)
		//} else if  j, err := encrypt.GetAes().DecryptCBC(encryptData, hex.DecodeString(key)); err != nil {
		//	t.Error(err)
		//} else {
		//	fmt.Println("j is ", string(j))
		//}

		jsonStr, err = json.Marshal(parameter)
		if err != nil {
			fmt.Printf("to json error:%v\n", err)
			return
		}
		if resp, err := post(url, jsonStr, resp["access_token"].(string)); err != nil {
			t.Error(err)
		} else {
			data, _ := resp["data"].(string)
			secretData, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				t.Error(err)
			}
			jsonStr, err := encrypt.GetAes().DecryptCBC(secretData, key)
			if err != nil {
				t.Error(err)
			}
			fmt.Println(string(jsonStr))
		}
	} else {
		t.Error(err)
	}
}

func post(url string, jsonStr []byte, token string) (response map[string]interface{}, err error) {
	var req *http.Request
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	var resp *http.Response
	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		fmt.Printf("resp error:%v\n", err)
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("status_ode:%d\nheader:%+v\nbody:%s\n", resp.StatusCode, resp.Header, body)
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func TestErrshow(t *testing.T) {
	err := foundation.NewError(nil, 1, "haha")
	fmt.Println(err.Error())
}