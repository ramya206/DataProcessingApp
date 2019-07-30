package main

import (
	"github.com/gorilla/websocket"
	"time"
)

type Profile struct {
	MemberId string `bson:"MemberID,omitempty" json:"MemberID,omitempty"`
	DeviceId string		`bson:"DeviceId,omitempty" json:"DeviceId,omitempty"`
	Name string		`bson:"Name,omitempty" json:"Name,omitempty"`
	Age string		`bson:"Age,omitempty" json:"Age,omitempty"`
	Squad string	`bson:"Squad,omitempty" json:"Squad,omitempty"`
	Status string	`bson:"Status,omitempty" json:"Status,omitempty"`
	Image string	`bson:"Image,omitempty" json:"Image,omitempty"`
}

type SensorData struct {

	DeviceId string
	Location Location
	Temperature string
	Humidity string
	Pressure string
	Proximity string
	Acc_x string
	Acc_y string
	Acc_z string
	Gyr_x string
	Gyr_y string
	Gyr_z string
	Mag_x string
	Mag_y string
	Mag_z string
	Time time.Time

}

type StreamToSocket struct {

	Type string
	Data SensorData
}

type Location struct {
	Latitude string
	Longitude string
}

type Alert struct {
	Type string
	Data interface{}
	SensorData SensorData
	Location Location
}

type Client struct {
	Pool *MonitorPool
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan StreamToSocket
}


type FireFighter struct {

	Pool *MonitorPool

	Data chan SensorData
	//Buffered channel for incoming messages.

	//Client id
	deviceId string
}

type MonitorPool struct {

	Clients map[*websocket.Conn]bool   //clients connected using ws

	broadcast chan StreamToSocket       //channel to send sensor data to clients
	alert chan Alert
	Devices map[string]bool         //Devices sending real time data to server..
	DeviceMap map[string]*FireFighter
	DeviceRegister chan StreamToSocket
	send chan interface{}
}