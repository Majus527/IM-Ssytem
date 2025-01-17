package main

func main() {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
	// 测试一下插件
}
