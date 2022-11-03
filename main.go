package main

import (
	"fmt"
	config "logAgentEtcd/config"
	myEtcd "logAgentEtcd/etcd"
	taillog "logAgentEtcd/tailLog"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	// 先从配置文件读取etcd配置
	config.Init()
	logConfig := config.GetConfig()
	//  从ETCD中取出配置列表
	err := myEtcd.Init(strings.FieldsFunc(logConfig.Etcd.Endpoints, func(r rune) bool { return r == rune(',') }))
	if err != nil {
		fmt.Println("etcd init err, error: ", err)
	}
	localKey := fmt.Sprintf(logConfig.Etcd.Key, "ceshi")
	wg.Add(1)
	taillog.Init()
	err = myEtcd.GetConfig(localKey)
	if err != nil {
		fmt.Println("etcd get config err, error:", err)
	}
	wg.Wait()
}
