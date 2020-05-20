package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/webkit/gohttp"
)

type authApiImpl struct {
}

func (this *authApiImpl) NewAuth(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter NewAuthParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(AuthResult)
	if err := GetService().NewAuth(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}

func (this *authApiImpl) AuthRefresh(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter AuthRefreshParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(AuthResult)
	if err := GetService().AuthRefresh(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}

func (this *authApiImpl) AccessTokenAuth(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter AccessTokenAuthParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(TokenItem)
	if err := GetService().AccessTokenAuth(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}