package database

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"sync"
	"time"
)

type IRedisManager interface {
	Open()
	Client() *redis.Client

	// base set & get
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string)

	// hash table set & get
	HashExists(key, field string) (bool, error)
	HashLen(key string) (int64, error)
	HashSet(key string, values ...interface{}) error
	HashGet(key, field string) (string, error)
	HashMSet(key string, values ...interface{}) error
	HashMGet(key string, fields ...string) ([]interface{}, error)
	HashDelete(key string, fields ...string) error
	HashKeys(key string) ([]string, error)
	HashVals(key string) ([]string, error)
	HashGetAll(key string) (map[string]string, error)

	// try lock
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

var (
	_redisConfig RedisConfig
	_redisManagerOnce sync.Once
	_redisManager     IRedisManager
)

/**
 * 根据配置接口对象初始化
 */
func SetupRedis(redisConfig RedisConfig) IRedisManager {
	_redisConfig = redisConfig
	return Redis()
}

func Redis() IRedisManager {
	_redisManagerOnce.Do(func() {
		_redisManager = NewRedisManager(_redisConfig)
	})
	return _redisManager
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
 * redis client
 */
func (this *redisManagerImpl) Client() *redis.Client {
	return this.client
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
 * 名称为key的hash中是否存在键为field的域
 */
func (this *redisManagerImpl) HashExists(key, field string) (bool, error) {
	if exists, err := this.client.HExists(key, field).Result(); err != nil {
		return false, err
	} else {
		return exists, nil
	}
}
/**
 * 返回名称为key的hash中元素个数
 */
func (this *redisManagerImpl) HashLen(key string) (int64, error) {
	if len, err := this.client.HLen(key).Result(); err != nil {
		return 0, err
	} else {
		return len, nil
	}
}
/**
 * 向名称为key的hash中添加元素field
 */
func (this *redisManagerImpl) HashSet(key string, values ...interface{}) error {
	if _, err := this.client.HSet(key, values...).Result(); err != nil {
		return err
	}
	return nil
}
/**
 * 返回名称为key的hash中field对应的value
 */
func (this *redisManagerImpl) HashGet(key, field string) (string, error) {
	if ret, err := this.client.HGet(key, field).Result(); err!=nil {
		return "", err
	} else {
		return ret, nil
	}
}
/**
 * 向名称为key的hash中添加元素field
 */
func (this *redisManagerImpl) HashMSet(key string, values ...interface{}) error {
	if _, err := this.client.HMSet(key, values...).Result(); err != nil {
		return err
	}
	return nil
}
/**
 * 返回名称为key的hash中field i对应的value
 */
func (this *redisManagerImpl) HashMGet(key string, fields ...string) ([]interface{}, error) {
	if ret, err := this.client.HMGet(key, fields...).Result(); err!=nil {
		return nil, err
	} else {
		return ret, nil
	}
}
/**
 * 删除名称为key的hash中键为field的域
 */
func (this *redisManagerImpl) HashDelete(key string, fields ...string) error {
	if _, err := this.client.HDel(key, fields...).Result(); err != nil {
		return err
	}
	return nil
}
/**
 * 返回名称为key的hash中所有键
 */
func (this *redisManagerImpl) HashKeys(key string) ([]string, error) {
	if keys, err := this.client.HKeys(key).Result(); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}
/**
 * 返回名称为key的hash中所有键对应的value
 */
func (this *redisManagerImpl) HashVals(key string) ([]string, error) {
	if keys, err := this.client.HVals(key).Result(); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}
/**
 * 返回名称为key的hash中所有的键（field）及其对应的value
 */
func (this *redisManagerImpl) HashGetAll(key string) (map[string]string, error) {
	if m, err := this.client.HGetAll(key).Result(); err != nil {
		return nil, err
	} else {
		return m, nil
	}
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