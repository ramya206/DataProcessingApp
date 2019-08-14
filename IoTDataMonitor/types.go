package main

import (
	"encoding/json"
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
	Key string		`bson:"Key,omitempty" json:"key,omitempty"`
	Image string	`bson:"Image,omitempty" json:"Image,omitempty"`
}

type RawSensorData struct{
	State State	`json:"state,omitempty"`
	Metadata string	`json:"metadata,omitempty"`
	Version string	`json:"version,omitempty"`
	Timestamp json.Number	`json:"timestamp,omitempty,Number"`
	ClientToken string	`json:"clientToken,omitempty"`
}

type State struct {
	Reported SensorData  `json:"reported,omitempty"`
}


type LocationData struct{

	Latitude	json.Number `json:"latitude,omitempty,Number"`
	Longitude	json.Number `json:"longitude,omitempty,Number"`
	Altitude 	json.Number `json:"altitude (m),omitempty,Number"`
	time		interface{} 	`json:"time utc,omitempty"`

}


type SensorData struct {

	DeviceId string 	`json:"deviceId,omitempty"`
	Location Location	`json:"location,omitempty"`
	Temperature json.Number	`json:"temperature,omitempty,Number"`
	Humidity string 	`json:"humidity,omitempty"`
	Pressure string		`json:"pressure,omitempty"`
	Proximity string	`json:"proximity,omitempty"`
	Acc_x string	`json:"acc_x,omitempty"`
	Acc_y string	`json:"acc_y,omitempty"`
	Acc_z string	`json:"acc_z,omitempty"`
	Gyr_x string	`json:"gyr_x,omitempty"`
	Gyr_y string	`json:"gyr_y,omitempty"`
	Gyr_z string	`json:"gyr_z,omitempty"`
	Mag_x string	`json:"mag_x,omitempty"`
	Mag_y string	`json:"mag_y,omitempty"`
	Mag_z string	`json:"mag_z,omitempty"`
	Time time.Time	`json:"time,omitempty"`
	Squad string `json:"squad,omitempty"`
	Status string `json:"status,omitempty"`

	Altitude json.Number

}

type StreamToSocket struct {

	Type string
	Data map[string]SensorData
}

type StreamToSocket1 struct {

	Type string
	Data SensorData
}

type StreamToSocketLatLong struct {

	Type string
	Data LocationData
}

type Location struct {
	Latitude json.Number	`json:"lat,omitempty"`
	Longitude json.Number	`json:"lng,omitempty"`
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

	broadcast chan interface{}       //channel to send sensor data to clients
	alert chan Alert
	Devices map[string]bool         //Devices sending real time data to server..
	DeviceMap map[string]*FireFighter
	DeviceRegister chan StreamToSocket1
	send chan interface{}
}