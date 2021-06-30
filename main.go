package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const IOTURL = "https://a-62m15c-ubghzixbav:r+@6D*-wMzAw6U&4tA@62m15c.internetofthings.ibmcloud.com/api/v0002/"

func main() {

	//-------setting up route
	router := mux.NewRouter()
	router.HandleFunc("/last-event/{deviceType}/{deviceId}", GetDeviceLastEvent).Methods("GET")
	router.HandleFunc("/device-types", GetDeviceTypes).Methods("GET")
	router.HandleFunc("/devices/{deviceType}", GetDevices).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "UPDATE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	fmt.Println("Server is started...")
	log.Fatal(http.ListenAndServe(":8000", c.Handler(router)))
}

type DeviceLastEventInfo struct {
	Timestamp string     `json:"timestamp"`
	Payload   string     `json:"payload"`
	Data      DeviceData `json:"data"`
}

type DeviceData struct {
	Temperature int64 `json:"temperature"`
	Humidity    int64 `json:"humidity"`
	Weight      int64 `json:"weight"`
}

func GetDeviceLastEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceId = vars["deviceId"]

	url := IOTURL + "device/types/" + deviceType + "/devices/" + deviceId + "/events"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}

	lastEvent := []DeviceLastEventInfo{}

	err = json.Unmarshal(body, &lastEvent)
	if err != nil {
		fmt.Println(err.Error())
	}
	dataByte, err := base64.StdEncoding.DecodeString(lastEvent[0].Payload)
	if err != nil {
		panic("malformed input")
	}
	var deviceData DeviceData
	err = json.Unmarshal(dataByte, &deviceData)
	if err != nil {
		fmt.Println(err.Error())
	}
	lastEvent[0].Data = deviceData

	finalOutput, err := json.Marshal(lastEvent)
	if err != nil {
		fmt.Println(err.Error())
	}

	w.WriteHeader(200)
	w.Write(finalOutput)
	return
}

func GetDeviceTypes(w http.ResponseWriter, r *http.Request) {
	url := IOTURL + "device/types"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}
	w.WriteHeader(200)
	w.Write(body)
	return
}

func GetDevices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := IOTURL + "device/types/" + deviceType + "/devices"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}
	w.WriteHeader(200)
	w.Write(body)
	return
}
