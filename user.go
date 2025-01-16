package main

import "net"

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
	if msg == ":who" {
		// 查询当前在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + " is online..."
			this.C <- onlineMsg
		}
		this.server.mapLock.Unlock()

		return
	}
	// 将消息广播给全部的在线user
	// Message <- "[" + this.Addr + "]" + this.Name + ":" + msg
	this.server.BroadCast(this, msg)
}

// 监听当前user channel的方法，一旦有消息，直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
