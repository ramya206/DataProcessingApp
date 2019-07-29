package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)


var clients = make(map[*websocket.Conn]bool)
var devices = make(map[string]*FireFighter)
var broadcast = make(chan SensorData)
var pool *MonitorPool
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func serveWsClients(Pool *Pool, w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	clients[conn] = true
	pool = Pool




}



func handleStreamData() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-pool.broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func handleDevices() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-pool.broadcast  //**** create device broadcast channel
		var devId = msg.DeviceId
		if _,ok := pool.DeviceMap[devId]; !ok {
			device := &FireFighter{Pool : pool, Data : make(chan SensorData), deviceId :  devId}
			//device.Pool.Devices[devId] = true
			pool.DeviceMap[devId] = device
			go device.alertsGenerator()
			go connectivityMonitor()

		}
		pool.Devices[devId] = true
		pool.DeviceMap[devId].Data <- msg

	}
}



func (device *FireFighter) alertsGenerator(){

	var timerBuffer = 60; // in function of samples received (approx 1 sample/sec)
	var counter = 0;
	var backToNormality = false;

	for {
		msg := <-device.Data

		heartRateValue, err := strconv.ParseFloat(msg.Temperature, 64)
		if err!= nil {
			panic(err)
		}
		if backToNormality {counter++}
		if counter == timerBuffer{counter = 0}

		if heartRateValue >= 100 && counter == 0 {
			backToNormality = true;
			//sendAlertMessage to websocket clients
			counter++;
		}

		if (backToNormality && heartRateValue < 100 && counter == 0){

			//send safety message to websocket clients

			backToNormality = false;
			counter = 0;
		}
	}


}


func connectivityMonitor(){



for {

//	mutex.Lock()
for devId,_ := range pool.Devices {
	if !pool.Devices[devId] {

		fmt.Println("connectivityMonitor:  Device timed out!!!.")
		close(pool.DeviceMap[devId].Data)
		delete(pool.DeviceMap, devId)
		delete(pool.Devices, devId)

	} else {

		pool.Devices[devId] = false

	}
}

//	mutex.Unlock()
	time.Sleep(60 * time.Second)

}


//write logic to kill go routine
}

func (h *Pool) run() {
	fmt.Println("Client Pool started..")
	for {
		select {
		case client := <-h.register:
			h.clients[client.Clientid] = client
			fmt.Println("Client Registered..")
			var mesg Message
			mesg.Type = "OnlineUsers"
			mesg.Users = Users


		case client := <-h.unregister:
			if _, ok := h.clients[client.Clientid]; ok {
				delete(h.clients, client.Clientid)
				fmt.Println("Client Unregistered..",h.clients)
				close(client.send)
			}
			fmt.Println("Client Unregistered..")
		case message := <-h.broadcast:
			fmt.Println("Entered broadcast channel..",message)
			for client := range h.clients {
				h.clients[client].send <- message
				fmt.Println("Sending mesg to client channel..",client)

				//close(client.send)
				//delete(h.clients, client)
			}

		}
	}
}

