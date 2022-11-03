package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	taillog "logAgentEtcd/tailLog"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var client *clientv3.Client

func Init(endpoints []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		fmt.Println("create etcd client err, error: ", err)
	}

	return
}

func GetConfig(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	resp, err := client.Get(ctx, key)
	// 派一个哨兵检测etcd内容改动
	// 放置监听哨兵，监听日志收集配置改动
	watcher := clientv3.NewWatcher(client)
	ctx2, _ := context.WithCancel(context.TODO())
	watchRespChan := watcher.Watch(ctx2, key, clientv3.WithRev(resp.Header.Revision+1))
	// 增删改日志收集进程
	// resp.Header
	fmt.Printf("获取到etcd的配置信息是: %+v\n", resp)
	var (
		LogConfigList []*taillog.LogAgent
	)
	for _, v := range resp.Kvs {
		if string(v.Key) == key {
			fmt.Println("当前条目是：", string(v.Key))
			err = json.Unmarshal(v.Value, &LogConfigList)
			if err != nil {
				fmt.Println("json.Unmarshal err, error:", err)
				return
			}

			for _, cnf := range LogConfigList {
				fmt.Printf("条目信息：%s: %s\n", cnf.Topic, cnf.FilePath)
				// 每个条目都生成日志管理
				map_key := fmt.Sprintf("%s_%s", cnf.Topic, cnf.FilePath)
				taillog.LogMgrMap[map_key] = cnf.NewLogMgr()
			}
		}
	}
	// 处理监听到的etcd改动
	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				fmt.Println("修改为：", string(event.Kv.Value))
				var newConfigList []*taillog.LogAgent
				if string(event.Kv.Value) != "" {
					err = json.Unmarshal(event.Kv.Value, &newConfigList)
					if err != nil {
						fmt.Println("json.Unmarshal err, error:", err)
					}
				}
				taillog.ReloadLogMgr(newConfigList)
			case clientv3.EventTypeDelete:
				fmt.Println("删除了配置", event.Kv.Value)
			}
		}
	}
	return
}
