package auth

import (
	"encoding/json"
)

const (
	key_prefix_refresh_token 	= "/sean-tech/webkit/keys/auth/refreshtoken/"
	key_prefix_access_token 	= "/sean-tech/webkit/keys/auth/accesstoken/"
)

type daoImpl struct {
}

func (this *daoImpl) SaveRefreshTokenItem(userName string, tokenItem *TokenItem) error {

	var jsonVal, err = json.Marshal(tokenItem)
	if err != nil {
		return err
	}
	if err := this.DeleteAccessTokenItem(userName); err != nil {
		return err
	}
	if err := _storage.HashSet(key_prefix_refresh_token, userName, jsonVal); err!=nil {
		return err
	}
	return nil
}

func (this *daoImpl) GetRefreshTokenItem(userName string) (tokenItem *TokenItem, err error) {

	tokenItem = new(TokenItem)
	if jsonVal, err := _storage.HashGet(key_prefix_refresh_token, userName); err!=nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(jsonVal), tokenItem); err != nil {
		return nil, err
	}
	return tokenItem, nil
}

func (this *daoImpl) GetKey(userName string) (key string, err error) {
	if tokenItem, err := this.GetRefreshTokenItem(userName); err != nil {
		return "", err
	} else {
		return tokenItem.Key, nil
	}
}

func (this *daoImpl) DeleteRefreshTokenItem(userName string) error {
	if err := _storage.HashDelete(key_prefix_refresh_token, userName); err != nil {
		return err
	}
	// delete access token
	return this.DeleteAccessTokenItem(userName)
}



func (this *daoImpl) SaveAccessTokenItem(userName string, tokenItem *TokenItem) error {

	var jsonVal, err = json.Marshal(tokenItem)
	if err != nil {
		return err
	}
	if err := _storage.HashSet(key_prefix_access_token, userName, jsonVal); err!=nil {
		return err
	}
	return nil
}

func (this *daoImpl) GetAccessTokenItem(userName string) (tokenItem *TokenItem, err error) {
	tokenItem = new(TokenItem)
	if jsonVal, err := _storage.HashGet(key_prefix_access_token, userName); err!=nil {
		return nil, err
	} else if err := json.Unmarshal([]byte(jsonVal), tokenItem); err != nil {
		return nil, err
	}
	return tokenItem, nil
}

func (this *daoImpl) DeleteAccessTokenItem(userName string) error {
	if err := _storage.HashDelete(key_prefix_access_token, userName); err != nil {
		return err
	}
	return nil
}



func (this *daoImpl) DeleteAllTokenItem() error {
	// delete all refresh token
	if ret, err := _storage.HashKeys(key_prefix_refresh_token); err != nil {
		return err
	} else if err := _storage.HashDelete(key_prefix_refresh_token, ret...); err != nil {
		return err
	}
	// delete all access token
	if ret, err := _storage.HashKeys(key_prefix_access_token); err != nil {
		return err
	} else if err := _storage.HashDelete(key_prefix_access_token, ret...); err != nil {
		return err
	}
	return nil
}