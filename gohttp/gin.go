package gohttp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/requisition"
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
	key_request_id 						= "gohttp/key_request_id"
	key_ctx_requestion              	= "gohttp/key_ctx_requestion"
	_ secret_method 					= ""
	secret_method_rsa secret_method 	= "secret_method_rsa"
	secret_method_aes secret_method    	= "secret_method_aes"
	secret_method_nouse secret_method  	= "secret_method_nouse"
)

type CError interface {
	Code() int
	Msg() string
}

type HttpConfig struct {
	RunMode 			string			`json:"-" validate:"required,oneof=debug test release"`
	WorkerId 			int64			`json:"worker_id" validate:"min=0"`
	HttpPort            int				`json:"-"`
	ReadTimeout         time.Duration	`json:"read_timeout" validate:"required,gte=1"`
	WriteTimeout        time.Duration	`json:"write_timeout" validate:"required,gte=1"`
	CorsAllow			bool			`json:"cors_allow"`
	CorsAllowOrigins	[]string		`json:"cors_allow_origins"`
	RsaOpen bool                   		`json:"rsa_open"`
	Rsa 	*RsaConfig					`json:"-"`
}

/** 服务注册回调函数 **/
type GinRegisterFunc func(engine *gin.Engine)

var (
	_config   	HttpConfig
	_idWorker 	foundation.SnowId
	_logger 	IGinLogger
)

/**
 * 启动 api server
 * handler: 接口实现serveHttp的对象
 */
func HttpServerServe(config HttpConfig, logger IGinLogger, registerFunc GinRegisterFunc) {
	if err := validate.ValidateParameter(config); err != nil {
		log.Fatal(err)
	}
	if config.RsaOpen {
		if config.Rsa == nil {
			log.Fatal("server http start error : secret is nil")
		}
		if err := validate.ValidateParameter(config.Rsa); err != nil {
			log.Fatal(err)
		}
	}
	_config = config
	_idWorker, _ = foundation.NewWorker(config.WorkerId)
	if logger != nil {
		_logger = logger
	}

	// gin
	gin.SetMode(config.RunMode)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(_logger.Writer(), os.Stdout)

	// engine
	//engine := gin.Default()
	engine := gin.New()
	engine.Use(gin.Recovery())
	//engine.StaticFS(config.Upload.FileSavePath, http.Dir(GetUploadFilePath()))
	engine.Use(func(ctx *gin.Context) {
		var lang = ctx.GetHeader("Accept-Language")
		if  requisition.SupportLanguage(lang) == false {
			lang = requisition.LanguageZh
		}
		newGinRequestion(ctx).Language = lang
		var requestId = uint64(_idWorker.GetId())
		requisition.NewRequestion(ctx).RequestId = requestId
		ctx.Set(key_request_id, requestId)
		ctx.Next()
	})
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 你的自定义格式
		if param.ErrorMessage == "" {
			return fmt.Sprintf("[GIN]%s requestid:%d clientip:%s method:%s path:%s code:%d\n",
				param.TimeStamp.Format("2006/01/02 15:04:05"),
				param.Keys[key_request_id].(uint64),
				param.ClientIP,
				param.Method,
				param.Path,
				param.StatusCode,
			)
		}
		return fmt.Sprintf("[GIN]%s requestid:%d clientip:%s method:%s path:%s code:%d errmsg:%s\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Keys[key_request_id].(uint64),
			param.ClientIP,
			param.Method,
			param.Path,
			param.StatusCode,
			param.ErrorMessage,
			)
	}))
	if config.CorsAllow {
		if config.CorsAllowOrigins != nil {
			corscfg := cors.DefaultConfig()
			corscfg.AllowOrigins = config.CorsAllowOrigins
			engine.Use(cors.New(corscfg))
		} else {
			engine.Use(cors.Default())
		}

	}
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
type ginRequisition struct {
	Language 	 string
	SecretMethod secret_method `json:"secretMethod"`
	Params       []byte        `json:"params"`
	Key          []byte        `json:"key"`
	Rsa			 *RsaConfig
}

/**
 * 请求信息创建，并绑定至context上
 */
func newGinRequestion(ctx *gin.Context) *ginRequisition {
	rq := &ginRequisition{
		SecretMethod: secret_method_nouse,
		Params:       nil,
		Key:          nil,
		Rsa: 		  nil,
	}
	ctx.Set(key_ctx_requestion, rq)
	return rq
}

/**
 * 信息获取，获取传输链上context绑定的用户请求调用信息
 */
func (g *Gin) getGinRequisition() *ginRequisition {
	obj := g.Ctx.Value(key_ctx_requestion)
	if info, ok := obj.(*ginRequisition); ok {
		return  info
	}
	return nil
}

/**
 * 参数绑定
 */
func (g *Gin) BindParameter(parameter interface{}) error {

	switch g.getGinRequisition().SecretMethod {
	case secret_method_nouse:
		if err := g.Ctx.Bind(parameter); err != nil {
			return foundation.NewError(err, STATUS_CODE_INVALID_PARAMS, err.Error())
		}
		g.LogRequestParam(parameter)
		return nil
	case secret_method_aes:fallthrough
	case secret_method_rsa:
		if err := json.Unmarshal(g.getGinRequisition().Params, parameter); err != nil {
			return foundation.NewError(err, STATUS_CODE_INVALID_PARAMS, err.Error())
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
	var code = STATUS_CODE_SUCCESS
	var msg = requisition.Msg(g.getGinRequisition().Language, code)

	switch g.getGinRequisition().SecretMethod {
	case secret_method_nouse:
		g.LogResponseInfo(code, msg, data, "")
		g.Response(code, msg, data, "")
		return
	case secret_method_aes:
		jsonBytes, _ := json.Marshal(data)
		if secretBytes, err := encrypt.GetAes().EncryptCBC(jsonBytes, g.getGinRequisition().Key); err == nil {
			g.LogResponseInfo(code, msg, jsonBytes, "")
			g.Response(code, msg, base64.StdEncoding.EncodeToString(secretBytes), "")
			return
		}
		g.LogResponseInfo(code, msg, data, "response data aes encrypt failed")
		g.Response(code, msg, data, "response data aes encrypt failed")
		return
	case secret_method_rsa:
		jsonBytes, _ := json.Marshal(data)
		if secretBytes, err := encrypt.GetRsa().Encrypt(g.getGinRequisition().Rsa.ClientPubKey, jsonBytes); err == nil {
			if signBytes, err := encrypt.GetRsa().Sign(g.getGinRequisition().Rsa.ServerPriKey, jsonBytes); err == nil {
				sign := base64.StdEncoding.EncodeToString(signBytes)
				g.LogResponseInfo(code, msg, jsonBytes, sign)
				g.Response(code, msg, base64.StdEncoding.EncodeToString(secretBytes), sign)
				return
			}
		}
		g.LogResponseInfo(code, msg, data, "response data rsa encrypt failed")
		g.Response(code, msg, data, "response data rsa encrypt failed")
		return
	}
}

/**
 * 响应数据，自定义error
 */
func (g *Gin) ResponseError(err error) {
	if e, ok := err.(requisition.IError); ok {
		e.SetLang(g.getGinRequisition().Language)
	}
	if e, ok := err.(foundation.IError); ok {
		g.LogResponseError(e.Code(), e.Msg(), e.Error())
		g.Response(e.Code(), e.Msg(), nil, "")
		return
	}
	g.LogResponseError(STATUS_CODE_FAILED, err.Error(), "")
	g.Response(STATUS_CODE_FAILED, err.Error(), nil, "")
}

/**
 * 响应数据
 */
func (g *Gin) Response(statusCode int, msg string, data interface{}, sign string) {
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
	if _logger == nil {
		return
	}
	var req = requisition.GetRequisition(g.Ctx)
	if jsonBytes, ok := parameter.([]byte); ok {
		_logger.Gin("requestid:", req.RequestId, " user_name:", req.UserName, " params:", string(jsonBytes), "\n")
	} else if jsonBytes, err := json.Marshal(parameter); err == nil {
		_logger.Gin("requestid:", req.RequestId, " user_name:", req.UserName, " params:", string(jsonBytes), "\n")
	} else {
		_logger.Gin("requestid:", req.RequestId, " user_name:", req.UserName, " params:", parameter, "\n")
	}
}

func (g *Gin) LogResponseInfo(code int, msg string, data interface{}, sign string) {
	if _logger == nil {
		return
	}
	var req = requisition.GetRequisition(g.Ctx)
	if jsonBytes, ok := data.([]byte); ok {
		_logger.Gin("requestid:", req.RequestId, " username:", req.UserName, " respcode:", code, " msg:", msg, " data:", string(jsonBytes), " sign:", sign, "\n")
	} else if jsonBytes, err := json.Marshal(data); err == nil {
		_logger.Gin("requestid:", req.RequestId, " username:", req.UserName, " respcode:", code, " msg:", msg, " data:", string(jsonBytes), " sign:", sign, "\n")
	} else {
		_logger.Gin("requestid:", req.RequestId, " username:", req.UserName, " respcode:", code, " msg:", msg, " data:", data, " sign:", sign, "\n")
	}
}

func (g *Gin) LogResponseError(code int, msg string, err string) {
	if _logger == nil {
		return
	}
	var req = requisition.GetRequisition(g.Ctx)
	_logger.Gin("requestid:", req.RequestId, " username:", req.UserName, " respcode:", code, " msg:", msg, " error:", err, "\n")
}

