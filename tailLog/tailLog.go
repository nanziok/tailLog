package taillog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hpcloud/tail"
)

type LogAgent struct {
	Topic    string `json:"topic"`
	FilePath string `json:"filePath"`
}
type LogMgr struct {
	Instance   *LogAgent
	Begin_time time.Time
	Ctx        context.Context
	Cancel     context.CancelFunc
}

var LogMgrMap map[string]*LogMgr

type msg struct {
	topic string
	value string
}

var msgChan chan *msg

func Init() {
	LogMgrMap = make(map[string]*LogMgr, 16)
	msgChan = make(chan *msg, 100000)
	go sendMsg()
}

func (t *LogAgent) NewLogMgr() (logMgr *LogMgr) {
	ctx, cancel := context.WithCancel(context.Background())
	logMgr = &LogMgr{
		Instance:   t,
		Begin_time: time.Now(),
		Ctx:        ctx,
		Cancel:     cancel,
	}
	// 开启一个进程去收集日志
	go logMgr.doTailLog()
	return
}

// doTailLog 去收集日志
func (t *LogMgr) doTailLog() {
	for {
		fileConfig := tail.Config{
			Location:  &tail.SeekInfo{Offset: 2, Whence: os.SEEK_END},
			Follow:    true,
			ReOpen:    true,
			MustExist: false,
			Poll:      true,
			Pipe:      true,
		}
		tails, err := tail.TailFile(t.Instance.FilePath, fileConfig)
		if err != nil {
			fmt.Println("filetail init err, error:", err)
		}
		var (
			line *tail.Line
			ok   bool
		)
		for {
			select {
			case <-t.Ctx.Done():
				return
			case line, ok = <-tails.Lines:
				if !ok {
					fmt.Printf("通道：%s,获取新的日志行错误\n", t.Instance.Topic)
					continue
				}
				fmt.Printf("检测到通道：%s, 有新的日志：%s\n", t.Instance.Topic, line.Text)
			default:
				time.Sleep(time.Millisecond * 50)
			}
		}
	}
}

// 关闭收集日志进程
func (t *LogMgr) cancelTailLog() {
	t.Cancel()
}

func ReloadLogMgr(reNew []*LogAgent) {
	origin := LogMgrMap
	var map_key string
	var matchAll bool
	for k, v := range origin {
		matchAll = false
		for _, v_n := range reNew {
			map_key = fmt.Sprintf("%s_%s", v_n.Topic, v_n.FilePath)
			if k == map_key {
				matchAll = true
			}
		}
		if !matchAll {
			v.Cancel()
			delete(origin, k)
		}
	}
	for _, v_n := range reNew {
		matchAll = false
		map_key = fmt.Sprintf("%s_%s", v_n.Topic, v_n.FilePath)
		for k, _ := range origin {
			if k == map_key {
				matchAll = true
			}
		}
		if !matchAll {
			origin[map_key] = v_n.NewLogMgr()
		}
	}
}

// sendMsg 公用的发送日志方法
func sendMsg() {
	for {
		select {
		case msg := <-msgChan:
			fmt.Printf("通道有新的内容，%+v\n", msg)
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
}
