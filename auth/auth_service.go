package auth

import (
	"context"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/sean-tech/gokit/encrypt"
	"github.com/sean-tech/gokit/requisition"
	"github.com/sean-tech/gokit/validate"
	"strconv"
	"time"
)



type serviceImpl struct {
	authCode string
}

func (this *serviceImpl) NewAuth(ctx context.Context, parameter *NewAuthParameter, result *AuthResult) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	if parameter.AuthCode != this.authCode {
		return requisition.NewError(nil, status_code_auth_code_wrong)
	}

	// refresh token
	var refreshTokenItem *TokenItem; var err error
	if refreshTokenItem, err = generateToken(parameter.UUID, parameter.Client, parameter.UserId, parameter.UserName, parameter.Role, ""); err != nil {
		return err
	}
	var key = hex.EncodeToString(encrypt.GetAes().GenerateKey())
	refreshTokenItem.Key = key
	if err := dao().SaveRefreshTokenItem(parameter.UserName, refreshTokenItem); err != nil {
		return requisition.NewError(err, status_code_auth_token_savefailed)
	}

	// access token
	var accessTokenItem *TokenItem
	if accessTokenItem, err = generateToken(parameter.UUID, parameter.Client, parameter.UserId, parameter.UserName, parameter.Role, refreshTokenItem.Id); err != nil {
		return requisition.NewError(err, status_code_auth_token_generatefailed)
	}
	accessTokenItem.Key = key
	if err := dao().SaveAccessTokenItem(parameter.UserName, accessTokenItem); err != nil {
		return requisition.NewError(err, status_code_auth_token_savefailed)
	}

	// response
	*result = AuthResult{
		RefreshToken: refreshTokenItem.Token,
		Key:          refreshTokenItem.Key,
		AccessToken:  accessTokenItem.Token,
	}
	return nil
}

func (this *serviceImpl) DelAuth(ctx context.Context, parameter *DelAuthParameter, result *bool) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	if parameter.AuthCode != this.authCode {
		return requisition.NewError(nil, status_code_auth_code_wrong)
	}
	if err := dao().DeleteAccessTokenItem(parameter.UserName); err != nil {
		return err
	}
	if err := dao().DeleteRefreshTokenItem(parameter.UserName); err != nil {
		return err
	}
	*result = true
	return nil
}

func (this *serviceImpl) AuthRefresh(ctx context.Context, parameter *AuthRefreshParameter, result *AuthResult) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	// access token validate
	var accessTokenAuthParameter = &AccessTokenAuthParameter{
		AccessToken: parameter.AccessToken,
	}
	var accessTokenItem = TokenItem{}; var err error
	if err = this.AccessTokenAuth(ctx, accessTokenAuthParameter, &accessTokenItem); err == nil {
		return requisition.NewError(err, status_code_auth_shouldnot_refresh)
	}
	if e, ok := err.(interface{Code() int}); ok && e.Code() != status_code_auth_token_shouldrefresh {
		return err
	}

	// refresh token validate
	var refreshTokenItem *TokenItem
	if refreshTokenItem, err = dao().GetRefreshTokenItem(accessTokenItem.UserName); err != nil {
		return err
	}
	if time.Now().Unix() > refreshTokenItem.ExpiresAt {
		_ = dao().DeleteRefreshTokenItem(refreshTokenItem.UserName)
		return requisition.NewError(nil, status_code_auth_token_timeout)
	}
	var refreshTokenClaims *TokenClaims
	if refreshTokenClaims, err = parseToken(refreshTokenItem.Token); err != nil {
		return err
	}

	// signed id fit
	if refreshTokenItem.Id != accessTokenItem.SignedId {
		if accessTokenItem.UUID != refreshTokenItem.UUID { // sso : other device login
			return requisition.NewError(nil, status_code_auth_token_otherlogin)
		}
		return requisition.NewError(nil, status_code_auth_token_checkfaild)
	}

	// new auth
	var newAuthParameter = &NewAuthParameter{
		AuthCode: this.authCode,
		UUID:     refreshTokenItem.UUID,
		UserId:   refreshTokenClaims.UserId,
		UserName: refreshTokenItem.UserName,
		Role: 	  refreshTokenItem.Role,
		Client:   refreshTokenItem.Client,
	}
	return this.NewAuth(ctx, newAuthParameter, result)
}

func (this *serviceImpl) AccessTokenAuth(ctx context.Context, parameter *AccessTokenAuthParameter, accessTokenItem *TokenItem) error {
	if err := validate.ValidateParameter(parameter); err != nil {
		return err
	}
	// parse token
	var accessTokenClaims *TokenClaims; var err error
	if accessTokenClaims, err = parseToken(parameter.AccessToken); err != nil {
		return err
	}
	// saved token validate
	var savedAccessTokenItem *TokenItem
	if savedAccessTokenItem, err = dao().GetAccessTokenItem(accessTokenClaims.UserName); err != nil {
		return err
	}
	if savedAccessTokenItem.Token != parameter.AccessToken {
		if accessTokenClaims.UUID != savedAccessTokenItem.UUID {
			return requisition.NewError(err, status_code_auth_token_otherlogin)
		}
		return requisition.NewError(err, status_code_auth_token_checkfaild)
	}
	*accessTokenItem = *savedAccessTokenItem
	return nil
}


type TokenClaims struct {
	UUID 		string 		`json:"uuid"`
	UserId 		uint64 		`json:"userId"`
	UserName 	string 		`json:"userName"`
	Role 		string		`json:"role"`
	SignedId	string		`json:"signed_id"`
	jwt.StandardClaims
}

func generateToken(uuid, client string, userId uint64, userName, role, signedId string) (tokenItem *TokenItem, err error) {
	expireTime := time.Now().Add(_config.RefreshTokenExpiresTime)
	iat := time.Now().Unix()
	jti := strconv.FormatInt(_idWorker.GetId(), 10)
	c := TokenClaims{
		UUID: 			uuid,
		UserId:			userId,
		UserName:       userName,
		Role: 			role,
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
			Role: 	   role,
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
		return nil, requisition.NewError(nil, status_code_auth_token_empyt)
	}

	// parse token
	tokenClaims, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(_config.TokenSecret), nil
	})
	if err != nil {
		e, ok := err.(jwt.ValidationError)
		if ok {
			return nil, err
		}
		switch e.Errors {
		case jwt.ValidationErrorExpired:       // EXP validation failed
			return nil, requisition.NewError(nil, status_code_auth_token_timeout)
		case jwt.ValidationErrorSignatureInvalid: fallthrough // Signature validation failed
		case jwt.ValidationErrorIssuedAt: fallthrough      // IAT validation failed
		case jwt.ValidationErrorIssuer: fallthrough        // ISS validation failed
		case jwt.ValidationErrorId:            // JTI validation failed
			return nil, requisition.NewError(nil, status_code_auth_token_checkfaild)
		default:
			return nil, err
		}
	} else if tokenClaims == nil {
		return nil, requisition.NewError(nil, status_code_auth_token_checkfaild)
	} else if !tokenClaims.Valid {
		return nil, requisition.NewError(nil, status_code_auth_token_checkfaild)
	}

	// token info validate
	tokenInfo, ok := tokenClaims.Claims.(*TokenClaims)
	if !ok {
		return nil, requisition.NewError(nil, status_code_auth_token_typewrong)
	} else if tokenInfo.Issuer != _config.TokenIssuer {
		return nil, requisition.NewError(nil, status_code_auth_token_checkfaild)
	} else if time.Now().Unix() > tokenInfo.ExpiresAt {
		return nil, requisition.NewError(nil, status_code_auth_token_shouldrefresh)
	}
	return tokenInfo, nil
}
