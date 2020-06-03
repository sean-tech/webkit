package config

import (
	"context"
	"errors"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"strconv"
	"strings"
	"time"
)
const (
	dial_timeout    = 5 * time.Second
	request_timeout = 3 * time.Second
)

var _cli *clientv3.Client

func clientInit() {
	if _params == nil {
		panic(errors.New("etcd client init with nil cmd params"))
	}

	var err error
	if _cli, err = clientv3.New(clientv3.Config{
		Endpoints:   _params.EtcdEndPoints,
		DialTimeout: dial_timeout,
		Username:    _params.EtcdConfigUserName,
		Password:    _params.EtcdConfigPassword,
	}); err != nil {
		panic(err)
	}
	// defer _cli.Close()
}

func PutConfig(cfg *AppConfig, module, salt string) error {

	if err := cfg.Validate(); err != nil {
		return err
	}
	var path = fmt.Sprintf("%s/%s", _params.EtcdConfigPath, module)
	if value, err := configEncrypt(cfg, module, salt); err != nil {
		return err
	} else if resp, err := _cli.Put(context.Background(), path, string(value), clientv3.WithPrevKV()); err != nil {
		return err
	} else {
		_ = resp
		//Log.Println(resp)
		return nil
	}
}

func GetConfig(module, salt string) (cfg *AppConfig, err error) {
	var path = fmt.Sprintf("%s/%s", _params.EtcdConfigPath, module)
	if resp, err := _cli.Get(context.Background(), path); err != nil {
		return nil, err
	} else if len(resp.Kvs) != 1 {
		return nil, errors.New("config get error:kvs count not only one")
	} else {
		fmt.Printf("%+v\n", resp.Kvs)
		kv := resp.Kvs[0]
		if cfg, err := configDecrypt(string(kv.Value), module, salt); err != nil {
			return nil, err
		} else {
			return cfg, nil
		}
	}
}

func DeleteConfig(module string) error {
	var path = fmt.Sprintf("%s/%s", _params.EtcdConfigPath, module)
	if resp, err := _cli.Delete(context.Background(), path); err != nil {
		return err
	} else {
		//fmt.Println(resp)
		_ = resp
		return nil
	}
}


func PutWorkerId( workerId int64, module, ip string) error {
	var path = fmt.Sprintf("%s/%s/%s", _params.EtcdConfigPath, module, ip)
	if resp, err := _cli.Put(context.Background(), path, strconv.FormatInt(workerId, 10), clientv3.WithPrevKV()); err != nil {
		return err
	} else {
		_ = resp
		//Log.Println(resp)
		return nil
	}
}

func GetWorkerId(module, ip string) (workerId int64, err error) {
	var path = fmt.Sprintf("%s/%s/%s", _params.EtcdConfigPath, module, ip)
	if resp, err := _cli.Get(context.Background(), path); err != nil {
		return 0, err
	} else if len(resp.Kvs) != 1 {
		return 0, errors.New("config get error:kvs count not only one")
	} else {
		fmt.Printf("%+v\n", resp.Kvs)
		kv := resp.Kvs[0]
		if workerId, err = strconv.ParseInt(string(kv.Value), 10, 64); err != nil {
			return 0, err
		}
		return workerId, nil
	}
}

func DeleteWorkerId(module, ip string) error {
	var path = fmt.Sprintf("%s/%s/%s", _params.EtcdConfigPath, module, ip)
	if resp, err := _cli.Delete(context.Background(), path); err != nil {
		return err
	} else {
		//fmt.Println(resp)
		_ = resp
		return nil
	}
}

type Worker struct {
	Ip string
	WorkerId int64
}
func GetAllWorkers(module string) (workers []Worker, err error) {
	var path = fmt.Sprintf("%s/%s/", _params.EtcdConfigPath, module)
	var resp *clientv3.GetResponse
	if resp, err = _cli.Get(context.Background(), path, clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", resp.Kvs)
	for _, kv := range resp.Kvs {
		var workerId int64
		if workerId, err = strconv.ParseInt(string(kv.Value), 10, 64); err != nil {
			continue
		}
		ip := strings.Replace(string(kv.Key), path, "", 1)
		workers = append(workers, Worker{
			Ip:       ip,
			WorkerId: workerId,
		})
	}
	return workers, nil
}