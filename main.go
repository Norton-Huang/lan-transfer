package main

import (
	"fmt"
	"github.com/zserge/lorca"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	PORT = 9999
)

func main() {
	// 开启服务
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", GetLanIp(), PORT),
		Handler: InitRouter(),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln("web启动失败: ", err)
		}
	}()

	// 开启后台任务
	InitMessageTask()

	// 窗口服务
	web := fmt.Sprintf("http://%s:%d/web", GetLanIp(), PORT)
	ui, err := lorca.New(web, "", 1000, 700, "--disable-sync", "--disable-translate")
	if err != nil {
		log.Fatalln("窗口开启失败：", err)
	}
	defer ui.Close()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM) // 监听中断信号及终止信号
	select {
	case <-signChan:
	case <-ui.Done():
	}
}
