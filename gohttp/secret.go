package gohttp

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/requisition"
	"github.com/sean-tech/gokit/validate"
	"log"
)

type SecretParams struct {
	Secret string	`json:"secret" validate:"required,base64"`
}

type RsaConfig struct {
	ServerPubKey 		string 			`json:"server_pub_key" validate:"required"`
	ServerPriKey		string 			`json:"server_pri_key" validate:"required"`
	ClientPubKey 		string 			`json:"client_pub_key" validate:"required"`
}

type TokenParseFunc func(ctx context.Context, token string) (userId uint64, userName, role, key string, err error)

/**
 * rsa拦截校验
 */
func InterceptRsa() gin.HandlerFunc {
	var rsa = _config.Rsa
	if err := validate.ValidateParameter(rsa); err != nil {
		log.Fatal(err)
	}
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		g.getGinRequisition().Rsa = rsa


		var code = STATUS_CODE_SUCCESS
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
		} else if jsonBytes, err = encrypt.GetRsa().Decrypt(rsa.ServerPriKey, encrypted); err != nil { // decrypt
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err = encrypt.GetRsa().Verify(rsa.ClientPubKey, jsonBytes, signDatas); err != nil { // sign verify
			code = STATUS_CODE_SECRET_CHECK_FAILED
		}
		// code check
		if code != STATUS_CODE_SUCCESS {
			g.ResponseError(requisition.NewError(nil, code))
			ctx.Abort()
			return
		}
		g.getGinRequisition().SecretMethod = secret_method_rsa
		g.getGinRequisition().Params = jsonBytes
		// next
		ctx.Next()
	}
}

/**
 * token拦截校验
 */
func InterceptToken(tokenParse TokenParseFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		if userId, userName, role, key, err := tokenParse(ctx, ctx.GetHeader("Authorization")); err != nil {
			g.ResponseError(err)
			ctx.Abort()
			return
		} else {
			requisition.GetRequisition(ctx).UserId = userId
			requisition.GetRequisition(ctx).UserName = userName
			requisition.GetRequisition(ctx).Role = role
			g.getGinRequisition().Key, _ = hex.DecodeString(key)
			// next
			ctx.Next()
		}
	}
}

/**
 * aes拦截校验
 */
func InterceptAes() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g := Gin{ctx}
		var code = STATUS_CODE_SUCCESS
		var params SecretParams
		var encrypted []byte
		var jsonBytes []byte

		// params handle
		if err := g.Ctx.Bind(&params); err != nil { // bind
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if err := validate.ValidateParameter(params); err != nil { // validate
			code = STATUS_CODE_INVALID_PARAMS
		} else if encrypted, err = base64.StdEncoding.DecodeString(params.Secret); err != nil { // decode
			code = STATUS_CODE_SECRET_CHECK_FAILED
		} else if jsonBytes, err = encrypt.GetAes().DecryptCBC(encrypted, g.getGinRequisition().Key); err != nil { // decrypt
			code = STATUS_CODE_SECRET_CHECK_FAILED
		}
		// code check
		if code != STATUS_CODE_SUCCESS {
			g.ResponseError(requisition.NewError(nil, code))
			ctx.Abort()
			return
		}

		g.getGinRequisition().SecretMethod = secret_method_aes
		g.getGinRequisition().Params = jsonBytes
		// next
		ctx.Next()
	}
}
