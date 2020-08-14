package gohttp

import (
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/requisition"
	"github.com/storyicon/grbac"
	"net/http"
	"time"
)

func Authorization(rulesLoader  func()(grbac.Rules, error)) gin.HandlerFunc {
	// 在这里，我们通过“grbac.WithLoader”接口使用自定义Loader功能
	// 并指定应每分钟调用一次LoadAuthorizationRules函数以获取最新的身份验证规则。
	// Grbac还提供一些现成的Loader：
	// grbac.WithYAML
	// grbac.WithRules
	// grbac.WithJSON
	// ...
	rbac, err := grbac.New(grbac.WithLoader(rulesLoader, time.Minute))
	if err != nil {
		panic(err)
	}
	//if err != nil {
	//	c.AbortWithError(http.StatusInternalServerError, err)
	//	return
	//}
	return func(c *gin.Context) {
		roles := []string{requisition.GetRequisition(c).Role}
		state, _ := rbac.IsRequestGranted(c.Request, roles)
		if !state.IsGranted() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}