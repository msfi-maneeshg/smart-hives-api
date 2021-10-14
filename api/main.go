package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"smart-hives/api/api"
	"smart-hives/api/database"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const IOTURL = "https://a-8l173e-otjztnyacu:ChLq7u0pO+*hl7JER_@8l173e.internetofthings.ibmcloud.com/api/v0002/"

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	database.Data = client.Database("smart-hives")

	fmt.Println("MongoDB is connected!")
}

func main() {
	//-------setting up route
	router := mux.NewRouter()
	router.HandleFunc("/last-event/{deviceType}/{deviceId}", api.GetDeviceLastEvent).Methods("GET")
	router.HandleFunc("/device-types", api.GetDeviceTypes).Methods("GET")
	router.HandleFunc("/devices/{deviceType}", api.GetDevices).Methods("GET")
	router.HandleFunc("/process/{farmer}", api.ProcessFarmerData).Methods("GET")
	router.HandleFunc("/hive-data/{farmer}/{date}/{period}", api.ProcessedFarmerData).Methods("GET")

	router.HandleFunc("/iot/device/types/{deviceType}", api.GetDeviceType).Methods("GET")
	router.HandleFunc("/iot/device/types/{deviceType}", api.CreateNewDeviceType).Methods("POST")

	router.HandleFunc("/iot/device/types/{deviceType}/devices", api.GetDeviceList).Methods("GET")
	router.HandleFunc("/iot/device/types/{deviceType}/devices", api.CreateNewDevice).Methods("POST")
	router.HandleFunc("/iot/device/types/{deviceType}/devices/{deviceID}", api.GetDeviceInfo).Methods("GET")
	router.HandleFunc("/iot/device/types/{deviceType}/devices/{deviceID}", api.DeleteDeviceInfo).Methods("DELETE")
	router.HandleFunc("/iot/device/types/{deviceType}/devices/{deviceID}", api.UpdateDeviceInfo).Methods("PUT")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "UPDATE", "DELETE", "PUT"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	fmt.Println("Server is started...")
	log.Fatal(http.ListenAndServe(":8000", c.Handler(router)))
}
