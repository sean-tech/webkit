package serving

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/logging"
	"github.com/sean-tech/gokit/validate"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type User struct {
	UserId int						`json:"user_id" validate:"required,min=1"`
	UserName string					`json:"user_name" validate:"required,eorp"`
	Email string					`json:"email" validate:"required,email"`
}

type GoodsPayInfoParameter struct {
	GoodsId int			`json:"goods_id" validate:"required,min=1"`
	GoodsName string	`json:"goods_name" validate:"required,gte=1"`
	GoodsAmount int		`json:"goods_amount" validate:"required,min=1"`
	Remark string 		`json:"remark" validate:"gte=0"`
}

type GoodsPayParameter struct {
	UserInfo *User					`json:"user_info" validate:"required"`
	Goods []*GoodsPayInfoParameter	`json:"goods" validate:"required,gte=1,dive,required"`
	GoodsIds []int				`json:"goods_ids" validate:"required,gte=1,dive,min=1"`
}

func TestGinServer(t *testing.T) {
	logging.Setup(logging.LogConfig{
		LogSavePath:     "/Users/lyra/Desktop/",
		LogPrefix:       "gintest",
	})
	// server start
	HttpServerServe(HttpConfig{
		RunMode:        "debug",
		WorkerId:       0,
		HttpPort:       8001,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		JwtSecret:      "webkit/serving/jwtsecret/token@20200427",
		JwtIssuer:      "sean.tech/webkit/user",
		JwtExpiresTime: 36 * time.Hour,
		SecretOpen:     false,
		ServerPubKey:   "",
		ServerPriKey:   "",
		ClientPubKey:   "",
		Logger:         logging.Logger(),
		SecretStorage:  NewMemeoryStorage(),
	}, RegisterApi)
}

func RegisterApi(engine *gin.Engine) {
	apiv1 := engine.Group("api/order/v1")
	{
		apiv1.POST("/bindtest", bindtest)
	}
}

func bindtest(ctx *gin.Context)  {
	date := ctx.Request.Header.Get("Date")
	fmt.Println(date)
	g := Gin{
		Ctx: ctx,
	}
	var parameter GoodsPayParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var payMoney float64 = 0
	if err := GoodsPay(ctx, &parameter, &payMoney); err != nil {
		g.ResponseError(err)
		return
	}
	var resp = make(map[string]string)
	resp["payMoney"] = fmt.Sprintf("%v", payMoney)
	g.ResponseData(resp)
}

func GoodsPay(ctx context.Context, parameter *GoodsPayParameter, payMoney *float64) error {
	err := validate.ValidateParameter(parameter)
	if err != nil {
		return foundation.NewError(STATUS_CODE_INVALID_PARAMS, err.Error())
	}
	*payMoney = 10.0
	return nil
}

func TestPostToGinServer(t *testing.T)  {
	var url = "http://localhost:8001/api/order/v1/bindtest"

	var user_info map[string]interface{} = make(map[string]interface{})
	user_info["user_id"] = 101
	user_info["user_name"] = "18922311056"
	user_info["email"] = "1028990481@qq.com"

	var goods1 map[string]interface{} = make(map[string]interface{})
	goods1["goods_id"] = 1001
	goods1["goods_name"] = "三只松鼠干果巧克力100g包邮"
	goods1["goods_amount"] = 1
	goods1["remark"] = ""
	var goods []interface{} = []interface{}{goods1}
	var goods_ids []int = []int{1}

	var parameter map[string]interface{} = make(map[string]interface{})
	parameter["user_info"] = user_info
	parameter["goods"] = goods
	parameter["goods_ids"] = goods_ids

	jsonStr, err := json.Marshal(parameter)
	if err != nil {
		fmt.Printf("to json error:%v\n", err)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	//defer resp.Body.Close()
	if err != nil {
		fmt.Printf("resp error:%v", err)
	} else {
		statuscode := resp.StatusCode
		hea := resp.Header
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		fmt.Println(statuscode)
		fmt.Println(hea)
	}
}