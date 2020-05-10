package database

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

type IRedisManager interface {
	Open()
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string)
	TryLock(key string, expiration time.Duration) (result bool)
	ReleaseLock(key string) (result bool)
}

type RedisConfig struct{
	Host        string			`json:"host" validate:"required,tcp_addr"`
	Password    string			`json:"password" validate:"gte=0"`
	MaxIdle     int				`json:"max_idle" validate:"required,min=1"`
	MaxActive   int				`json:"max_active" validate:"required,min=1"`
	IdleTimeout time.Duration	`json:"idle_timeout" validate:"required,gte=1"`
}

func NewRedisManager(redisConfig RedisConfig) IRedisManager {
	return &redisManagerImpl{
		config:redisConfig,
		client: redis.NewClient(&redis.Options{
			Addr:     redisConfig.Host,
			Password: redisConfig.Password,
			DB:       0,  // use default DB
			IdleTimeout:redisConfig.IdleTimeout,
		}),
	}
}

type redisManagerImpl struct {
	config RedisConfig
	client *redis.Client
}

/**
 * 开启redis并初始化客户端连接
 */
func (this *redisManagerImpl) Open() {
	pong, err := this.client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	// 初始化后通讯失败
	if err != nil {
		panic(err)
	}
}

/**
 * 存
 */
func (this *redisManagerImpl) Set(key string, value interface{}, expiration time.Duration) error {
	err := this.client.Set(key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

/**
 * 取
 */
func (this *redisManagerImpl) Get(key string) (string, error) {
	val, err := this.client.Get(key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

/**
 * 删除key
 */
func (this *redisManagerImpl) Delete(key string) {
	this.client.Del(key)
}

/**
 * try lock
 */
func (this *redisManagerImpl) TryLock(key string, expiration time.Duration) (result bool) {
	// lock
	resp := this.client.SetNX(key, 1, expiration)
	lockSuccess, err := resp.Result()
	if err != nil || !lockSuccess {
		return false
	}
	return true
}

func (this *redisManagerImpl) ReleaseLock(key string) (result bool) {
	delResp := this.client.Del(key)
	unlockSuccess, err := delResp.Result()
	if err == nil && unlockSuccess > 0 {
		return true
	} else {
		return false
	}
}