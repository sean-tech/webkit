package auth

import (
	"context"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/foundation"
	"strconv"
	"time"
)



type authServiceImpl struct {
	authCode string
}

func (this *authServiceImpl) NewAuth(ctx context.Context, parameter *NewAuthParameter, result *AuthResult) error {
	if parameter.AuthCode != this.authCode {
		return foundation.NewError(status_code_auth_code_wrong, status_msg_auth_code_wrong)
	}

	// refresh token
	var refreshTokenItem *TokenItem; var err error
	if refreshTokenItem, err = generateToken(parameter.UUID, parameter.Client, parameter.UserId, parameter.UserName, ""); err != nil {
		return err
	}
	var key = hex.EncodeToString(encrypt.GetAes().GenerateKey())
	refreshTokenItem.Key = key
	if err := getDao().SaveRefreshTokenItem(parameter.UserName, refreshTokenItem); err != nil {
		return foundation.NewError(status_code_auth_token_savefailed, status_msg_auth_token_savefailed)
	}

	// access token
	var accessTokenItem *TokenItem
	if accessTokenItem, err = generateToken(parameter.UUID, parameter.Client, parameter.UserId, parameter.UserName, refreshTokenItem.Id); err != nil {
		return foundation.NewError(status_code_auth_token_generatefailed, status_msg_auth_token_generatefailed)
	}
	accessTokenItem.Key = key
	if err := getDao().SaveAccessTokenItem(parameter.UserName, accessTokenItem); err != nil {
		return foundation.NewError(status_code_auth_token_savefailed, status_msg_auth_token_savefailed)
	}

	// response
	*result = AuthResult{
		RefreshToken: refreshTokenItem.Token,
		Key:          refreshTokenItem.Key,
		AccessToken:  accessTokenItem.Token,
	}
	return nil
}

func (this *authServiceImpl) AuthRefresh(ctx context.Context, parameter *AuthRefreshParameter, result *AuthResult) error {
	// access token validate
	var accessTokenAuthParameter = &AccessTokenAuthParameter{
		AccessToken: parameter.AccessToken,
	}
	var accessTokenItem = TokenItem{}; var err error
	if err = this.AccessTokenAuth(ctx, accessTokenAuthParameter, &accessTokenItem); err == nil {
		return foundation.NewError(status_code_auth_shouldnot_refresh, status_msg_auth_shouldnot_refresh)
	}
	if e, ok := err.(interface{Code() int}); ok && e.Code() != status_code_auth_token_shouldrefresh {
		return err
	}

	// refresh token validate
	var refreshTokenItem *TokenItem
	if refreshTokenItem, err = getDao().GetRefreshTokenItem(accessTokenItem.UserName); err != nil {
		return err
	}
	if time.Now().Unix() > refreshTokenItem.ExpiresAt {
		_ = getDao().DeleteRefreshTokenItem(refreshTokenItem.UserName)
		return foundation.NewError(status_code_auth_token_timeout, status_msg_auth_token_timeout)
	}
	var refreshTokenClaims *TokenClaims
	if refreshTokenClaims, err = parseToken(refreshTokenItem.Token); err != nil {
		return err
	}

	// signed id fit
	if refreshTokenItem.Id != accessTokenItem.SignedId {
		if accessTokenItem.UUID != refreshTokenItem.UUID { // sso : other device login
			return foundation.NewError(status_code_auth_token_otherlogin, status_msg_auth_token_otherlogin)
		}
		return foundation.NewError(status_code_auth_token_checkfaild, status_msg_auth_token_checkfaild)
	}

	// new auth
	var newAuthParameter = &NewAuthParameter{
		AuthCode: this.authCode,
		UUID:     refreshTokenItem.UUID,
		UserId:   refreshTokenClaims.UserId,
		UserName: refreshTokenItem.UserName,
		Client:   refreshTokenItem.Client,
	}
	return this.NewAuth(ctx, newAuthParameter, result)
}

func (this *authServiceImpl) AccessTokenAuth(ctx context.Context, parameter *AccessTokenAuthParameter, accessTokenItem *TokenItem) error {
	// parse token
	var accessTokenClaims *TokenClaims; var err error
	if accessTokenClaims, err = parseToken(parameter.AccessToken); err != nil {
		return err
	}
	// saved token validate
	var savedAccessTokenItem *TokenItem
	if savedAccessTokenItem, err = getDao().GetAccessTokenItem(accessTokenClaims.UserName); err != nil {
		return err
	}
	if savedAccessTokenItem.Token != parameter.AccessToken {
		if accessTokenClaims.UUID != savedAccessTokenItem.UUID {
			return foundation.NewError(status_code_auth_token_otherlogin, status_msg_auth_token_otherlogin)
		}
		return foundation.NewError(status_code_auth_token_checkfaild, status_msg_auth_token_checkfaild)
	}
	*accessTokenItem = *savedAccessTokenItem
	return nil
}


type TokenClaims struct {
	UUID 		string 		`json:"uuid"`
	UserId 		uint64 		`json:"userId"`
	UserName 	string 		`json:"userName"`
	SignedId	string		`json:"signed_id"`
	jwt.StandardClaims
}

func generateToken(uuid string,  client string, userId uint64, userName string, signedId string) (tokenItem *TokenItem, err error) {
	expireTime := time.Now().Add(_config.RefreshTokenExpiresTime)
	iat := time.Now().Unix()
	jti := strconv.FormatInt(_idWorker.GetId(), 10)
	c := TokenClaims{
		UUID: 			uuid,
		UserId:			userId,
		UserName:       userName,
		SignedId: 		signedId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    _config.TokenIssuer,
			Id:			jti,
			IssuedAt:iat,
			NotBefore: iat,
			Subject:"client",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	if token, err := tokenClaims.SignedString([]byte(_config.TokenSecret)); err != nil {
		return nil, err
	} else {
		return &TokenItem{
			Id:        jti,
			ExpiresAt: expireTime.Unix(),
			UUID:      uuid,
			SignedId:  signedId,
			UserId:    userId,
			UserName:  userName,
			Token:     token,
			Key:       "",
			Client:    client,
		}, nil
	}
}

/**
 * 解析token
 */
func parseToken(token string) (*TokenClaims, error) {
	if token == "" || len(token) == 0 {
		return nil, foundation.NewError(status_code_auth_token_empyt, status_msg_auth_token_empyt)
	}

	// parse token
	tokenClaims, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(_config.TokenSecret), nil
	})
	if err != nil {
		return nil, err
	} else if tokenClaims == nil {
		return nil, foundation.NewError(status_code_auth_token_checkfaild, status_msg_auth_token_checkfaild)
	} else if !tokenClaims.Valid {
		return nil, foundation.NewError(status_code_auth_token_checkfaild, status_msg_auth_token_checkfaild)
	}

	// token info validate
	tokenInfo, ok := tokenClaims.Claims.(*TokenClaims)
	if !ok {
		return nil, foundation.NewError(status_code_auth_token_typewrong, status_msg_auth_token_typewrong)
	} else if tokenInfo.Issuer != _config.TokenIssuer {
		return nil, foundation.NewError(status_code_auth_token_checkfaild, status_msg_auth_token_checkfaild)
	} else if time.Now().Unix() > tokenInfo.ExpiresAt {
		return nil, foundation.NewError(status_code_auth_token_shouldrefresh, status_msg_auth_token_shouldrefresh)
	}
	return tokenInfo, nil
}
