package main

import (
	"fmt"
	"os"
)
var pool = newPool()
func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8005"
	}
	fmt.Println("entered...Listen to Kinesis");
	listenToKinesis();
	fmt.Println("entered...Running pool");

	go pool.run()
	fmt.Println("entered...starting handleSender");

	go pool.handleSender()
	fmt.Println("entered...starting connectivityMonitor");

	go pool.connectivityMonitor()
	fmt.Println("entered...starting server");
	server := NewServer()
	server.Run(":" + port)
}
