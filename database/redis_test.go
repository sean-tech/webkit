package database

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var redisManager IRedisManager
var key_saved string = "github.com/sean-tech/webkit/database/redis_test/key_saved"
func redisStart()  {
	config := RedisConfig{
		Host:        "127.0.0.1:6379",
		Password:    "",
		MaxIdle:     30,
		MaxActive:   30,
		IdleTimeout: 200 * time.Second,
	}
	redisManager = NewRedisManager(config)
	redisManager.Open()
}

func TestRedisSet(t *testing.T) {
	redisStart()
	var savedValue string = "test redis saved value"
	if err := redisManager.Set(key_saved, savedValue, time.Minute *5); err != nil {
		t.Error(err)
	}
	if value, err := redisManager.Get(key_saved); err != nil {
		t.Error(err)
	} else {
		if value == savedValue {
			fmt.Println("saved success:", value)
		} else {
			fmt.Println("saved failed:", value)
		}
	}
}

func TestDelete(t *testing.T) {
	redisStart()
	redisManager.Delete(key_saved)
	if value, err := redisManager.Get(key_saved); err != nil {
		t.Error(err)
	} else {
		fmt.Println("value after delete:", value)
	}
}


const (
	OrderStatusWaitPay = "未支付"
	OrderStatusHasPay = "已支付"
)
type Order struct {
	orderId int64
	status string
}

func TestTryLock(t *testing.T) {
	redisStart()

	const max_by_num = 50
	var order = Order{
		orderId: 123665405101,
		status: OrderStatusWaitPay,
	}

	var key_trylocak string = "github.com/sean-tech/webkit/database/redis_test/key_trylocak/order.pay/order_id/"
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			if lockResult := redisManager.TryLock(key_trylocak, time.Minute * 5); lockResult == false {
				fmt.Println("支付处理中，请勿重复提交...")
				return
			}
			defer func() {
				for i:= 0; i < 3; i++ {
					if releaseSuccess := redisManager.ReleaseLock(key_trylocak); releaseSuccess {
						return
					}
					time.Sleep(time.Millisecond * 10)
				}
			}()
			if order.status == OrderStatusHasPay {
				fmt.Println("订单已支付...")
				return
			}
			order.status = OrderStatusHasPay
			fmt.Println("订单支付成功...")
		}()
	}
	wg.Wait()
	fmt.Println("pay done over")
}