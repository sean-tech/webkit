package serving

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type secret_method string
const (
	key_request_id 					= "gohttp/key_request_id"
	key_ctx_requestion              = "gohttp/key_ctx_requestion"
	_ secret_method 				= ""
	secret_method_rsa               = "secret_method_rsa"
	secret_method_aes               = "secret_method_aes"
	secret_method_nouse             = "secret_method_nouse"
)

type CError interface {
	Code() int
	Msg() string
}

type HttpConfig struct {
	RunMode 			string			`json:"run_mode" validate:"required,oneof=debug test release"`
	WorkerId 			int64			`json:"worker_id" validate:"min=0"`
	HttpPort            int				`json:"http_port" validate:"required,min=1,max=10000"`
	ReadTimeout         time.Duration	`json:"read_timeout" validate:"required,gte=1"`
	WriteTimeout        time.Duration	`json:"write_timeout" validate:"required,gte=1"`
	// jwt
	JwtSecret 			string			`json:"jwt_secret" validate:"required,gte=1"`
	JwtIssuer 			string			`json:"jwt_issuer" validate:"required,gte=1"`
	JwtExpiresTime 		time.Duration	`json:"jwt_expires_time" validate:"required,gte=1"`
	// storage
	Logger       		IGinLogger    	`json:"logger" validate:"required"`
	SecretStorage 		ISecretStorage  `json:"secret_storage" validate:"required"`
	// secret
	SecretOpen			bool			`json:"secret_open"`
	ServerPubKey 		string 			`json:"server_pub_key"`
	ServerPriKey 		string 			`json:"server_pri_key"`
	ClientPubKey 		string 			`json:"client_pub_key"`
}
/** 服务注册回调函数 **/
type GinRegisterFunc func(engine *gin.Engine)

var (
	_httpConfig HttpConfig
	_idWorker   foundation.SnowId
)

/**
 * 启动 api server
 * handler: 接口实现serveHttp的对象
 */
func HttpServerServe(config HttpConfig, registerFunc GinRegisterFunc) {
	if err := validate.ValidateParameter(config); err != nil {
		log.Fatal(err)
	}
	_httpConfig = config
	_idWorker, _ = foundation.NewWorker(config.WorkerId)

	// gin
	gin.SetMode(config.RunMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(config.Logger.Writer(), os.Stdout)

	// engine
	//engine := gin.Default()
	engine := gin.New()
	engine.Use(gin.Recovery())
	//engine.StaticFS(config.Upload.FileSavePath, http.Dir(GetUploadFilePath()))
	engine.Use(func(ctx *gin.Context) {
		newRequestion(ctx)
		foundation.NewRequestion(ctx).RequestId = uint64(_idWorker.GetId())
		ctx.Set(key_request_id, foundation.GetRequisition(ctx).RequestId)
		ctx.Next()
	})
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 你的自定义格式
		return fmt.Sprintf("[GIN] %s request_id:%d | %s | \"%s %s %s %d %s \"%s\" %s\"\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Keys[key_request_id].(uint64),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	registerFunc(engine)
	// server
	s := http.Server{
		Addr:           fmt.Sprintf(":%d", config.HttpPort),
		Handler:        engine,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(fmt.Sprintf("Listen: %v\n", err))
		}
	}()
	// signal
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<- quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

type Gin struct {
	Ctx *gin.Context
}

/**
 * 服务信息
 */
type requisition struct {
	SecretMethod secret_method `json:"secretMethod"`
	Params       []byte        `json:"params"`
	Key          []byte        `json:"key"`
}

/**
 * 请求信息创建，并绑定至context上
 */
func newRequestion(ctx *gin.Context) *requisition {
	rq := &requisition{
		SecretMethod: secret_method_nouse,
		Params:       nil,
		Key:          nil,
	}
	ctx.Set(key_ctx_requestion, rq)
	return rq
}

/**
 * 信息获取，获取传输链上context绑定的用户请求调用信息
 */
func (g *Gin) getRequisition() *requisition {
	obj := g.Ctx.Value(key_ctx_requestion)
	if info, ok := obj.(*requisition); ok {
		return  info
	}
	return nil
}

/**
 * 参数绑定
 */
func (g *Gin) BindParameter(parameter interface{}) error {

	switch g.getRequisition().SecretMethod {
	case secret_method_nouse:
		if err := g.Ctx.Bind(parameter); err != nil {
			return foundation.NewError(STATUS_CODE_INVALID_PARAMS, err.Error())
		}
		g.LogRequestParam(parameter)
		return nil
	case secret_method_aes:
	case secret_method_rsa:
		if err := json.Unmarshal(g.getRequisition().Params, parameter); err != nil {
			return foundation.NewError(STATUS_CODE_INVALID_PARAMS, err.Error())
		}
		g.LogRequestParam(parameter)
		return nil
	}
	return nil
}

/**
 * 响应数据，成功，原数据转json返回
 */
func (g *Gin) ResponseData(data interface{}) {

	var code StatusCode = STATUS_CODE_SUCCESS
	switch g.getRequisition().SecretMethod {
	case secret_method_nouse:
		g.Response(code, code.Msg(), data, "")
		return
	case secret_method_aes:
		jsonBytes, _ := json.Marshal(data)
		if secretBytes, err := encrypt.GetAes().EncryptCBC(jsonBytes, g.getRequisition().Key); err == nil {
			g.LogResponseInfo(code, code.Msg(), jsonBytes, "")
			g.Response(code, code.Msg(), base64.StdEncoding.EncodeToString(secretBytes), "")
			return
		}
		g.Response(code, code.Msg(), data, "response data aes encrypt failed")
		return
	case secret_method_rsa:
		jsonBytes, _ := json.Marshal(data)
		if secretBytes, err := encrypt.GetRsa().Encrypt(_httpConfig.ClientPubKey, jsonBytes); err == nil {
			if signBytes, err := encrypt.GetRsa().Sign(_httpConfig.ServerPriKey, jsonBytes); err == nil {
				sign := base64.StdEncoding.EncodeToString(signBytes)
				g.LogResponseInfo(code, code.Msg(), jsonBytes, sign)
				g.Response(code, code.Msg(), base64.StdEncoding.EncodeToString(secretBytes), sign)
				return
			}
		}
		g.Response(code, code.Msg(), data, "response data rsa encrypt failed")
		return
	}
}

/**
 * 响应数据，自定义error
 */
func (g *Gin) ResponseError(err error) {
	if e, ok := err.(CError); ok {
		g.Response(StatusCode(e.Code()), e.Msg(), nil, "")
		return
	}
	g.Response(STATUS_CODE_FAILED, err.Error(), nil, "")
}

/**
 * 响应数据
 */
func (g *Gin) Response(statusCode StatusCode, msg string, data interface{}, sign string) {

	if g.getRequisition().SecretMethod == secret_method_nouse || statusCode != STATUS_CODE_SUCCESS {
		g.LogResponseInfo(statusCode, msg, data, sign)
	}
	g.Ctx.JSON(http.StatusOK, gin.H{
		"code" : statusCode,
		"msg" :  msg,
		"data" : data,
		"sign" : sign,
	})
	return
}


type IGinLogger interface {
	Writer() io.Writer
	Gin(v ...interface{})
}

func (g *Gin) LogRequestParam(parameter interface{}) {
	var requestion = foundation.GetRequisition(g.Ctx)
	if jsonBytes, ok := parameter.([]byte); ok {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | params:", string(jsonBytes), "\n")
	} else if jsonBytes, err := json.Marshal(parameter); err == nil {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | params:", string(jsonBytes), "\n")
	} else {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | params:", parameter, "\n")
	}
}

func (g *Gin) LogResponseInfo(statusCode StatusCode, msg string, data interface{}, sign string) {
	var requestion = foundation.GetRequisition(g.Ctx)
	if jsonBytes, ok := data.([]byte); ok {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | response code:", statusCode, " | msg:", msg, " | data:", string(jsonBytes), " | sign:", sign, "\n")
	} else if jsonBytes, err := json.Marshal(data); err == nil {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | response code:", statusCode, " | msg:", msg, " | data:", string(jsonBytes), " | sign:", sign, "\n")
	} else {
		_httpConfig.Logger.Gin("request_id:", requestion.RequestId, "user_name:", requestion.UserName, " | response code:", statusCode, " | msg:", msg, " | data:", data, " | sign:", sign, "\n")
	}
}











