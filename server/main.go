package main

import "fmt"

func main() {
	server := NewServer("127.0.0.1:8080")
	server.start()
	fmt.Println("Server runnng")
}
