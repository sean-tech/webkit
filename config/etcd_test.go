package config

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
)

const (
	PATH = "sean-tech/webkit/etcdtest"
)

var (
	_dialTimeout    = 5 * time.Second
	_requestTimeout = 3 * time.Second
	_endpoints []string = []string{"127.0.0.1:2379"}
)

func TestPut(t *testing.T) {
	var value string = "value 3"
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   _endpoints,
		DialTimeout: _dialTimeout,
		Username: "",
		Password: "",
	})
	if err != nil {
		t.Error(err)
	}
	defer cli.Close()

	if resp, err := cli.Put(context.Background(), PATH, value, clientv3.WithPrevKV()); err != nil {
		t.Error(err)
	} else {
		_ = resp
		//Log.Println(resp)
	}
	fmt.Println("put success")
}

func TestDelete(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   _endpoints,
		DialTimeout: _dialTimeout,
	})
	if err != nil {
		t.Error(err)
	}
	defer cli.Close()

	if resp, err := cli.Delete(context.Background(), PATH); err != nil {
		t.Error(err)
	} else {
		//fmt.Println(resp)
		_ = resp
	}
	fmt.Println("delete success")
}

func TestGet(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   _endpoints,
		DialTimeout: _dialTimeout,
	})
	if err != nil {
		t.Error(err)
	}
	defer cli.Close()
	// global get
	if resp, err := cli.Get(context.Background(), PATH); err != nil {
		t.Error(err)
	} else {
		//if len(resp.Kvs) != 1 {
		//	err := errors.New("gloabl config get error:kvs count not only 1")
		//	t.Error(err)
		//}
		//kvs := resp.Kvs[0]
		//var global *GlobalConfig
		//if global, err = globalConfigWithJson(kvs.Value); err != nil {
		//	return nil, err
		//}
		//return global, nil
		fmt.Printf("%+v\n", resp.Kvs)
	}
}
