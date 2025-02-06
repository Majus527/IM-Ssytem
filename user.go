package main

import "net"
import "strings"

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
		server: server,
	}

	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

// 用户上线业务
func (this *User) Online() {
	// 用户上线，将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "is online")
}

// 用户下线业务
func (this *User) Offline() {
	// 用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	// 广播当前用户下线消息
	// this.BroadCast(this, "已下线")
	this.server.BroadCast(this, "is outline")
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	// 查询当前在线用户
	if msg == ":who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + " is online..."
			this.C <- onlineMsg
		}
		this.server.mapLock.Unlock()
		return
	}

	// 修改用户名
	if len(msg) > 7 && msg[:8] == ":rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.C <- "this name has been used"
			return
		}

		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()

		this.Name = newName
		this.C <- "change success:" + this.Name
		return
	}

	// 私聊
	if len(msg) > 4 && msg[:4] == ":to|" {
		// 获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.C <- "please input remote name"
			return
		}

		// 获取消息内容
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.C <- "please input content"
			return
		}

		// 查找对方用户
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.C <- "remote user is not online"
			return
		}

		// 将消息发送给对方用户
		remoteUser.C <- "[" + this.Name + "]" + ":" + content
		return
	}
	
	// 单纯的发消息
	this.server.BroadCast(this, msg)
}

// 监听当前user channel的方法，一旦有消息，直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 实时处理消息
func (this *User) SendMsg(msg string) (err error) {
	// 将消息发送给客户端
	this.conn.Write([]byte(msg + "\n"))
	return
}