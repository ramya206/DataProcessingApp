package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/satori/go.uuid"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"

	//"github.com/fatih/structs"


//"log"
	"net/http"
	"github.com/golang/glog"
	"reflect"
)

var mongodbServer = "localhost"
var mongodbDatabase = "iCare"
var ProfileCollection = "ProfileData"
var Unread_Messages = "UnreadCollection"

func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	n.UseHandler(handlers.CORS(allowedHeaders, allowedMethods, allowedOrigins)(mx))
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/profile/{id}", getRequestedProfile(formatter)).Methods("GET")
	mx.HandleFunc("/profiles", getAllProfile(formatter)).Methods("GET")
	mx.HandleFunc("/profile",addMember(formatter)).Methods("POST")
	mx.HandleFunc("/profile/{id}",removeMember(formatter)).Methods("DELETE")
	mx.HandleFunc("/ws", socketHandler(formatter)).Methods("GET")
}


func socketHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		fmt.Println("entered...",req.URL);
		serveWsClients(w, req)



	}
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		message := "Pong"
		formatter.JSON(w, http.StatusOK, struct{ Test string }{message})
	}
}

func getRequestedProfile(formatter *render.Render) http.HandlerFunc{
	return func(w http.ResponseWriter, req *http.Request){
		params := mux.Vars(req)
		var profileId string = params["id"];
		glog.Info("getRequestedProfile: profileID", profileId)

		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		glog.Info("getRequestedProfile: MongoDB connection success")
		c := session.DB(mongodbDatabase).C(ProfileCollection)

		var jsonProfile Profile
		profileData := make(map[string]Profile)
		profileDataBySquad := make(map[string]interface{})

		//query := bson.M{"deviceId": profileId}
		err = c.Find(bson.M{"deviceId": profileId}).All(&jsonProfile)
		if err!= nil{
			panic(err)
		}

		profileData[jsonProfile.DeviceId] = jsonProfile
		profileDataBySquad[jsonProfile.Squad] = profileData

		glog.Info("getRequestedProfile: Data returned from DB ",jsonProfile)

		glog.Info("getRequestedProfile: Parsing by squad value", profileDataBySquad)
		type ProfileDataFormat struct{
			Type string
			Data map[string]interface{}
		}

		jsonProfileData := ProfileDataFormat{
			Type:"Profile",
			Data : profileDataBySquad,

		}
		glog.Info("getRequestedProfile: Struct data before JSon parsing", jsonProfileData)
		_ = json.NewDecoder(req.Body).Decode(&jsonProfileData)
		glog.Info("getRequestedProfile: Response sending ", jsonProfileData)
		formatter.JSON(w, http.StatusOK, jsonProfileData)



	}
}

func getAllProfile(formatter *render.Render) http.HandlerFunc{
	return func(w http.ResponseWriter, req *http.Request){

		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		glog.Info("getAllProfile: MongoDB connection success")
		c := session.DB(mongodbDatabase).C(ProfileCollection)

		type FilterBySqaud struct{
			Type string
			Data map[string]interface{}
		}
		var queryResult []map[string]interface{}

		pipeline := []bson.M{bson.M{ "$group": bson.M{ "_id": "$Squad", "data": bson.M{ "$push": "$$ROOT"}}}}

		err = c.Pipe(pipeline).All(&queryResult)
		if err!= nil{
			panic(err)
		}

		fmt.Println("getAllProfile: Data returned from DB ",queryResult)

		var  i int = 0

		var arrayOfSquads = make([]map[string]interface{}, len(queryResult))

		for _, val := range queryResult {

			squad := val["_id"]
			fmt.Println("getAllProfile: The Squad value is  *********************",squad);
			var squadMap = make(map[string]interface{})
			profileArray := reflect.ValueOf(val["data"])

			var arrayOfIds = make([]map[string]Profile, profileArray.Len())
 				 for i:=0; i<profileArray.Len() ; i++{

 				 	arrayVal := profileArray.Index(i).Elem()

					 var jsonData Profile
					 if arrayVal.Kind() == reflect.Map {
							resultMap := make(map[string]interface{})
						 for _, e := range arrayVal.MapKeys() {
							 v := arrayVal.MapIndex(e)

							 s:=e.Interface()
							l:= s.(string)
							 switch t := v.Interface().(type)  {
							 case string:
								 fmt.Println(e, t)

								 resultMap[l] = t

							 default:
								 fmt.Println("not found")

							 }
						 }
						 fmt.Println("getAllProfile:The DEvice id from Map *********************",resultMap["DeviceId"]);
						 var devId string = resultMap["DeviceId"].(string)
						 jsonData.DeviceId  = resultMap["DeviceId"].(string)
						 jsonData.Squad = resultMap["Squad"].(string)
						 jsonData.Status = resultMap["Status"].(string)
						 jsonData.Age = resultMap["Age"].(string)
						 jsonData.Name = resultMap["Name"].(string)
						 jsonData.Image = resultMap["Image"].(string)
						  idMap := make(map[string]Profile)

						 idMap[devId] = jsonData
						 if idMap != nil && len(idMap) != 0 {
							 arrayOfIds[i] = idMap
						 }
					 }

					 if err != nil {
						 fmt.Println(err)
					 }

					 fmt.Println(" getAllProfile: Json Profile ...........",jsonData)
				 }
			fmt.Println("getAllProfile: The Array of Ids is  *********************",arrayOfIds);

			squadMap[squad.(string)] = arrayOfIds
			if squadMap != nil && len(squadMap) != 0{
				arrayOfSquads[i] = squadMap
			}
			i++;
		}
		type ProfileDataFormat struct{
			Type string
			Data []map[string]interface{}
		}

		jsonProfileData := ProfileDataFormat{
			Type:"Profile",
			Data : arrayOfSquads,

		}
		glog.Info("getAllProfile: Struct data before JSon parsing", jsonProfileData)
		_ = json.NewDecoder(req.Body).Decode(&jsonProfileData)
		glog.Info("getAllProfile: Response sending ", jsonProfileData)
		formatter.JSON(w, http.StatusOK, jsonProfileData)
	}
}

func addMember(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var addProfile Profile
		_ = json.NewDecoder(req.Body).Decode(&addProfile)
		 message := "New member Added"
		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
			message = "Server Error"
			formatter.JSON(w, http.StatusInternalServerError, struct{ Test string }{message})
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(ProfileCollection)
		glog.Info("addMember: Incoming Data", addProfile)

		addProfile.Status = "Good"

		errin := c.Insert(addProfile)
		if errin != nil {
			panic(err)
			glog.Info("addMember: Incoming Data", addProfile)
			message = "Server Error"
			formatter.JSON(w, http.StatusInternalServerError, struct{ Test string }{message})

		}



		formatter.JSON(w, http.StatusOK, struct{ Test string }{message})
	}
}


func removeMember(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		params := mux.Vars(req)
		var profileId string = params["id"];
		message := "Member successfully deleted!"
		session, err := mgo.Dial(mongodbServer)
		if err != nil {
			panic(err)
			message = "Server Error"
			formatter.JSON(w, http.StatusInternalServerError, struct{ Test string }{message})
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(mongodbDatabase).C(ProfileCollection)
		glog.Info("addMember: Member device Id to delete", profileId)

		errin := c.Remove(bson.M{"DeviceId":profileId})
		if errin != nil {
			panic(err)
			glog.Info("addMember: Error while removing member", err)
			message = "Server Error"
			formatter.JSON(w, http.StatusInternalServerError, struct{ Test string }{message})
		}

		formatter.JSON(w, http.StatusOK, struct{ Test string }{message})
	}
}