package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/webkit/gohttp"
)

type apiImpl struct {
}

func (this *apiImpl) NewAuth(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter NewAuthParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(AuthResult)
	if err := Service().NewAuth(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}

func (this *apiImpl) AuthRefresh(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter AuthRefreshParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(AuthResult)
	if err := Service().AuthRefresh(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}

func (this *apiImpl) AccessTokenAuth(ctx *gin.Context) {
	g := gohttp.Gin{
		Ctx: ctx,
	}
	var parameter AccessTokenAuthParameter
	if err := g.BindParameter(&parameter); err != nil {
		g.ResponseError(err)
		return
	}
	var result = new(TokenItem)
	if err := Service().AccessTokenAuth(ctx, &parameter, result); err != nil {
		g.ResponseError(err)
		return
	}
	g.ResponseData(result)
}