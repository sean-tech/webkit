package gohttp

import (
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/requisition"
	"github.com/storyicon/grbac"
	"net/http"
	"time"
)

func Authorization(rulesLoader  func()(grbac.Rules, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		rbac, err := grbac.New(grbac.WithLoader(rulesLoader, time.Minute))
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			if _logger != nil {
				_logger.Error(err.Error())
			}
			return
		}
		roles := []string{requisition.GetRequisition(c).Role}
		state, _ := rbac.IsRequestGranted(c.Request, roles)
		if !state.IsGranted() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}