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
	handleRouter(router)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "UPDATE", "DELETE", "PUT"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	fmt.Println("Server is started...")
	log.Fatal(http.ListenAndServe(":8000", c.Handler(router)))
}

func handleRouter(router *mux.Router) {
	router.HandleFunc("/last-event/{deviceType}/{deviceId}", api.GetDeviceLastEvent).Methods("GET")
	router.HandleFunc("/device-types", api.GetDeviceTypes).Methods("GET")
	router.HandleFunc("/devices/{deviceType}", api.GetDevices).Methods("GET")
	router.HandleFunc("/process/{farmer}", api.ProcessFarmerData).Methods("GET")
	router.HandleFunc("/hive-data/{date}/{period}", api.ProcessedFarmerData).Methods("GET")

	router.HandleFunc("/register", api.Register).Methods("POST")
	router.HandleFunc("/login", api.Login).Methods("POST")
	router.HandleFunc("/refresh-token", api.RefreshToken).Methods("POST")

	router.HandleFunc("/iot/device/types/{deviceType}", api.GetDeviceType).Methods("GET")
	router.HandleFunc("/iot/device/types/{deviceType}", api.CreateNewDeviceType).Methods("POST")

	router.HandleFunc("/devices", api.GetDeviceList).Methods("GET")
	router.HandleFunc("/devices", api.CreateNewDevice).Methods("POST")
	router.HandleFunc("/devices/{deviceID}", api.GetDeviceInfo).Methods("GET")
	router.HandleFunc("/devices/{deviceID}", api.DeleteDeviceInfo).Methods("DELETE")
	router.HandleFunc("/devices/{deviceID}", api.UpdateDeviceInfo).Methods("PUT")
}
