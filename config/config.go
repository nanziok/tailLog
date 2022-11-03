package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

type Kafka struct {
	Address string `ini:"address"`
}
type Etcd struct {
	Endpoints string `ini:"endpoints"`
	Key       string `ini:"key"`
}

type Cfg struct {
	Kafka `ini:"kafka"`
	Etcd  `ini:"etcd"`
}

var cnf = new(Cfg)

func Init() {
	err := ini.MapTo(cnf, "./tmp/config.ini")
	if err != nil {
		fmt.Println("ini mapTo err, error:", err)
		return
	}
	fmt.Printf("get ini config success, %v\n", cnf)

}

func GetConfig() *Cfg {
	return cnf
}
