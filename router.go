package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"log"
	"net/http"
	"time"
)

func InitRouter() *gin.Engine {
	/**
	总的服务包括
	1.显示文件列表的web界面（可进行消息的订阅）
	2.web界面支持多文件的上传
	3.负责分发文件的server服务
	*/
	router := gin.New()
	router.LoadHTMLGlob("template/*")
	router.GET("/web", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"server": fmt.Sprintf("ws://%s:%d/server", GetLanIp(), PORT),
			"web":    fmt.Sprintf("http://%s:%d/web", GetLanIp(), PORT),
			"upload": fmt.Sprintf("http://%s:%d/upload", GetLanIp(), PORT),
			"qrcode": fmt.Sprintf("http://%s:%d/qrcode", GetLanIp(), PORT),
		})
	})
	router.POST("/upload", upload)
	router.GET("/server", server)
	router.GET("/qrcode", webQrcode)

	return router
}

/**
文件上传
*/
func upload(c *gin.Context) {
	ip, _ := c.RemoteIP()
	form, _ := c.MultipartForm()
	files := form.File["files"]
	for _, file := range files {
		f, _ := file.Open()
		dat := make([]byte, file.Size)
		n, _ := f.Read(dat)

		HandleMessage(&Message{
			Mt:       1,
			Filename: file.Filename,
			Data:     dat[:n],
			Address:  ip.String(),
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			IsFile:   true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
	})
}

/**
订阅消息（创建ws连接就说明已订阅）
*/
func server(c *gin.Context) {
	// 创建长连接
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalln("server开启错误：", err)
	}
	// 缓存连接
	UserList[ws.RemoteAddr().String()] = ws

	ip, _ := c.RemoteIP()

	// 新订阅的用户，将错过的消息也一并推送过去
	go func() {
		for _, msg := range MessageList {
			f, _ := json.Marshal(msg)
			_ = ws.WriteMessage(msg.Mt, f)
		}
	}()

	// 死循环监听用户消息
	go func() {
		defer ws.Close()

		for {
			mt, msg, err := ws.ReadMessage()
			if err != nil {
				break
			}

			HandleMessage(&Message{
				Mt:       mt,
				Filename: "",
				Data:     msg,
				Address:  ip.String(),
				Time:     time.Now().Format("2006-01-02 15:04:05"),
				IsFile:   false,
			})
		}
	}()
}

/**
生成二维码
*/
func webQrcode(c *gin.Context) {
	// 生成一个访问web页面的二维码，供移动端客户端访问
	webCode, _ := qrcode.Encode(fmt.Sprintf("http://%s:%d/web", GetLanIp(), PORT), qrcode.Medium, 150)
	c.Writer.Write(webCode)
}
