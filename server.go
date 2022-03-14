package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Mt       int    `json:"mt"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
	Address  string `json:"address"`
	Time     string `json:"time"`
	IsFile   bool   `json:"is_file"`
}

var MessageChan = make(chan Message)

var UserList = make(map[string]*websocket.Conn)

var MessageList = make([]Message, 3)

// HandleMessage 处理收到的消息
func HandleMessage(msg *Message) {
	// 将消息写入消息管道，等待后台任务读取并推送给订阅者
	MessageChan <- *msg

	// 将消息写入总列表
	MessageList = append(MessageList, *msg)
}

// InitMessageTask 负责分发消息的任务
func InitMessageTask() {
	go func() {
		for {
			file := <-MessageChan
			for _, user := range UserList {
				f, _ := json.Marshal(file)
				_ = user.WriteMessage(file.Mt, f)
			}
		}
	}()
}
