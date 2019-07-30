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
var broadcast = make(chan StreamToSocket)
//var pool *MonitorPool
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}


func newPool() *MonitorPool {
	return &MonitorPool{
		Clients: make(map[*websocket.Conn]bool  ), //clients connected using ws
		broadcast: make( chan StreamToSocket  ) ,    //channel to send sensor data to clients
		alert: make(chan Alert),
		DeviceRegister : make(chan StreamToSocket),
		Devices : make(map[string]bool   ),      //Devices sending real time data to server..
		DeviceMap : make(map[string]*FireFighter),
		send : make(chan interface{}),
	}
}


func serveWsClients(Pool *MonitorPool, w http.ResponseWriter, r *http.Request) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	//pool = Pool
	Pool.Clients[conn] = true



	// client.Pool.broadcast <- mesg
//getAllProfile



}


func (client *Client) handlesendtOthisclient(){

}



func (pool *MonitorPool) run() {
	fmt.Println("Client Pool started..")
	for {
		select {
		case msg := <-pool.broadcast:
			pool.send<-msg



		case msg := <-pool.DeviceRegister:
			var sensorData SensorData
			sensorData = msg.Data
			var devId = sensorData.DeviceId
			if _,ok := pool.DeviceMap[devId]; !ok {
				device := &FireFighter{Pool : pool, Data : make(chan SensorData), deviceId :  devId}
				//device.Pool.Devices[devId] = true
				pool.DeviceMap[devId] = device
				go device.alertsGenerator()
				//go connectivityMonitor()

			}
			pool.Devices[devId] = true
			pool.DeviceMap[devId].Data <- sensorData
			//fmt.Println("Client Unregistered..")
		case msg := <-pool.alert:
			pool.send<-msg

		}
	}
}



func handleStreamData() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-pool.broadcast
		// Send it out to every client that is currently connected
		for client := range pool.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(pool.Clients, client)
			}
		}
	}
}

func handleDevices() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-pool.broadcast  //**** create device broadcast channel
		var sensorData SensorData
		sensorData = msg.Data
		var devId = sensorData.DeviceId
		if _,ok := pool.DeviceMap[devId]; !ok {
			device := &FireFighter{Pool : pool, Data : make(chan SensorData), deviceId :  devId}
			//device.Pool.Devices[devId] = true
			pool.DeviceMap[devId] = device
			go device.alertsGenerator()


		}
		pool.Devices[devId] = true
		pool.DeviceMap[devId].Data <- sensorData

	}
}


func (pool *MonitorPool) handleSender() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-pool.send  //**** create device broadcast channel
		for client := range pool.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(pool.Clients, client)
			}
		}

	}
}


func (device *FireFighter) alertsGenerator(){

	var timerBuffer = 60; // in function of samples received (approx 1 sample/sec)
	var counter = 0;
	var backToNormality = false;
	type heartbeat struct{
		DeviceId string

	}
	profileDataBySquad := getProfileById(device.deviceId)
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


			alert := Alert{
				Type:"HeartBeatAlert",
				SensorData : msg,
				Data: profileDataBySquad,
				Location:msg.Location,

			}
			//sendAlertMessage to websocket clients
			pool.alert<-alert
			counter++;
		}

		if (backToNormality && heartRateValue < 100 && counter == 0){

			//send safety message to websocket clients
			alert := Alert{
				Type:"HeartBeatNormal",
				SensorData : msg,
				Data: profileDataBySquad,
				Location:msg.Location,

			}
			pool.alert<-alert
			backToNormality = false;
			counter = 0;
		}
	}


}


func (pool *MonitorPool) connectivityMonitor(){



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
	time.Sleep(30 * time.Second)

}


//write logic to kill go routine
}



