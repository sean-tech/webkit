package auth

import (
	"encoding/json"
	"github.com/sean-tech/webkit/database"
)

const (
	key_prefix_refresh_token 	= "/sean-tech/webkit/keys/auth/refreshtoken/"
	key_prefix_access_token 	= "/sean-tech/webkit/keys/auth/accesstoken/"
)

type authDaoImpl struct {
}

func (this *authDaoImpl) SaveRefreshTokenItem(userName string, tokenItem *TokenItem) error {

	var jsonVal, err = json.Marshal(tokenItem)
	if err != nil {
		return err
	}
	if err := this.DeleteAccessTokenItem(userName); err != nil {
		return err
	}
	if _, err := database.Redis().Client().HSet(key_prefix_refresh_token, userName, string(jsonVal)).Result(); err!=nil {
		return err
	}
	return nil
}

func (this *authDaoImpl) GetRefreshTokenItem(userName string) (tokenItem *TokenItem, err error) {

	tokenItem = new(TokenItem)
	if jsonVal, err := database.Redis().Client().HGet(key_prefix_refresh_token, userName).Result(); err!=nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(jsonVal), tokenItem); err != nil {
		return nil, err
	}
	return tokenItem, nil
}

func (this *authDaoImpl) GetKey(userName string) (key string, err error) {
	if tokenItem, err := this.GetRefreshTokenItem(userName); err != nil {
		return "", err
	} else {
		return tokenItem.Key, nil
	}
}

func (this *authDaoImpl) DeleteRefreshTokenItem(userName string) error {
	if _, err := database.Redis().Client().HDel(key_prefix_refresh_token, userName).Result(); err != nil {
		return err
	}
	// delete access token
	return this.DeleteAccessTokenItem(userName)
}



func (this *authDaoImpl) SaveAccessTokenItem(userName string, tokenItem *TokenItem) error {

	var jsonVal, err = json.Marshal(tokenItem)
	if err != nil {
		return err
	}
	if _, err := database.Redis().Client().HSet(key_prefix_access_token, userName, string(jsonVal)).Result(); err!=nil {
		return err
	}
	return nil
}

func (this *authDaoImpl) GetAccessTokenItem(userName string) (tokenItem *TokenItem, err error) {
	tokenItem = new(TokenItem)
	if jsonVal, err := database.Redis().Client().HGet(key_prefix_access_token, userName).Result(); err!=nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(jsonVal), tokenItem); err != nil {
		return nil, err
	}
	return tokenItem, nil
}

func (this *authDaoImpl) DeleteAccessTokenItem(userName string) error {
	if _, err := database.Redis().Client().HDel(key_prefix_access_token, userName).Result(); err != nil {
		return err
	}
	return nil
}



func (this *authDaoImpl) DeleteAllTokenItem() error {
	// delete all refresh token
	if ret, err := database.Redis().Client().HKeys(key_prefix_refresh_token).Result(); err != nil {
		return err
	} else if _, err := database.Redis().Client().HDel(key_prefix_refresh_token, ret...).Result(); err != nil {
		return err
	}
	// delete all access token
	if ret, err := database.Redis().Client().HKeys(key_prefix_access_token).Result(); err != nil {
		return err
	} else if _, err := database.Redis().Client().HDel(key_prefix_access_token, ret...).Result(); err != nil {
		return err
	}
	return nil
}