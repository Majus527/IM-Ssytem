package main

import (
	"net"
	"fmt"
	"sync"
	"io"
	"time"
)

type Server struct {
	Ip	string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (this *Server) ListenMessager() {
	for {
		msg := <- this.Message

		// 将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 处理链接业务
func (this *Server) Handler(conn net.Conn) {
	// fmt.Println("链接建立成功")
	user := NewUser(conn, this)

	// 用户上线
	user.Online()
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			
			// 提取用户的消息(去除\n)
			msg := string(buf[:n-1])

			isLive <- true

			// 处理消息
			user.DoMessage(msg)
		}
	}()

	// 当前handler阻塞
	for {
		select {
			case <- isLive:
			case <- time.After(300 * time.Second):
				// 超时，强制关闭当前链接
				user.SendMsg("you are time out")
				close(user.C)
				conn.Close()
				return
		}
	}

}


// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessager()
	
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)

	}
}