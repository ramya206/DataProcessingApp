package main

import (
	"os"
)
var pool = newPool()
func main() {

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8005"
	}

	listenToKinesis();
	go pool.run()
	go pool.handleSender()
	go pool.connectivityMonitor()

	server := NewServer()
	server.Run(":" + port)
}
