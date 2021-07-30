package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/IBM/cloudant-go-sdk/auth"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
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
	router.HandleFunc("/process/{farmer}", ProcessFarmerData).Methods("GET")

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

type GetDBData struct {
	TotalRows int64        `json:"total_rows"`
	Rows      []DBDataRows `json:"rows"`
}

type DBDataRows struct {
	Doc DBDataRowsDoc `json:"doc"`
}
type DBDataRowsDoc struct {
	Data       DBDataRowsDocData `json:"data"`
	DeviceID   string            `json:"deviceId"`
	DeviceType string            `json:"deviceType"`
}

type DBDataRowsDocData struct {
	Humidity    int64 `json:"humidity"`
	Temperature int64 `json:"temperature"`
	Weight      int64 `json:"weight"`
}

func ProcessFarmerData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var farmer = vars["farmer"]
	fmt.Println(farmer)

	authenticator, err := auth.NewCouchDbSessionAuthenticator(
		"apikey-v2-29mnuuarysnz6zwv1np8fzp808a5e4052m4783hjkflh",
		"993856ca873efb33cc67fc6c82d6c7e8",
	)
	if err != nil {
		panic(err)
	}

	service, err := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName:   "CLOUDANT",
			URL:           "https://apikey-v2-29mnuuarysnz6zwv1np8fzp808a5e4052m4783hjkflh:993856ca873efb33cc67fc6c82d6c7e8@433c346a-cb7c-4736-8e95-0bc99303fe1a-bluemix.cloudantnosqldb.appdomain.cloud",
			Authenticator: authenticator,
		},
	)
	if err != nil {
		fmt.Print(err)
	}
	currentDate := time.Now().Format("2006-01-02")
	postAllDocsOptions := service.NewPostAllDocsOptions(
		"iotp_62m15c_" + farmer + "_" + currentDate,
	)
	postAllDocsOptions.SetIncludeDocs(true)
	// postAllDocsOptions.SetStartkey("abc")
	postAllDocsOptions.SetLimit(1000)

	allDocsResult, _, err := service.PostAllDocs(postAllDocsOptions)
	if err != nil {
		panic(err)
	}

	var objGetDBData GetDBData
	b, _ := json.MarshalIndent(allDocsResult, "", "  ")
	json.Unmarshal(b, &objGetDBData)

	var totalTemperature, totalRecords, avgTemperature int64
	var minTemperature, maxTemperature *int64

	for _, objData := range objGetDBData.Rows {
		totalRecords++
		temperature := objData.Doc.Data.Humidity
		totalTemperature += temperature

		if minTemperature == nil || *minTemperature > temperature {
			minTemperature = &temperature
		}

		if maxTemperature == nil || *maxTemperature < temperature {
			maxTemperature = &temperature
		}

	}
	avgTemperature = totalTemperature / totalRecords

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")

	//------------create a new document
	eventDoc := cloudantv1.Document{}
	eventDoc.SetProperty("type", "event")
	eventDoc.SetProperty("userid", "abc123")
	eventDoc.SetProperty("eventType", "addedToBasket")
	eventDoc.SetProperty("avgTemperature", avgTemperature)
	eventDoc.SetProperty("maxTemperature", maxTemperature)
	eventDoc.SetProperty("minTemperature", minTemperature)
	eventDoc.SetProperty("productId", "1000042")
	eventDoc.SetProperty("date", timestamp)

	putDocumentOptions := service.NewPutDocumentOptions(
		farmer+"-aggregated",
		currentDate+"_Q1",
	)
	putDocumentOptions.SetDocument(&eventDoc)

	documentResult, _, err := service.PutDocument(putDocumentOptions)
	if err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(documentResult, "", "  ")
	fmt.Println(string(b))

}
