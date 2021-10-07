package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"smart-hives/api/database"
	"time"

	"github.com/IBM/cloudant-go-sdk/auth"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

const IOTURL = "https://a-8l173e-otjztnyacu:ChLq7u0pO+*hl7JER_@8l173e.internetofthings.ibmcloud.com/api/v0002/"

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

func ProcessedFarmerData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var farmer = vars["farmer"]
	var date = vars["date"]
	var period = vars["period"]

	collection := database.Data.Collection(farmer)
	filterData := bson.D{{"date", date}, {"period", period}}
	cur, err := collection.Find(context.Background(), filterData)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())
	var allData []GetHiveData
	for cur.Next(context.Background()) {
		// To decode into a struct, use cursor.Decode()
		var result GetHiveData
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		allData = append(allData, result)
	}
	if err := cur.Err(); err != nil {
		fmt.Println(err.Error())
		return
	}
	body, _ := json.Marshal(allData)
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write(body)
	return
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

// CreateNewDevieType:
func CreateNewDevieType(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := IOTURL + "device/types"
	deviceTypeStatus := isDeviceTypeExist(deviceType)
	if deviceTypeStatus {
		message := []byte(`{"msg":"FarmerID is already exist"}`)
		w.WriteHeader(400)
		w.Write(message)
		return
	}

	var objNewDeviceType NewDeviceType
	objNewDeviceType.ID = deviceType
	objNewDeviceType.ClassId = "Device"
	objNewDeviceType.Description = "Hives for " + deviceType
	objNewDeviceType.metadata.MaxHumidity = 70
	objNewDeviceType.metadata.MinHumidity = 30
	objNewDeviceType.metadata.MaxTemperature = 45
	objNewDeviceType.metadata.MinTemperature = 35
	objNewDeviceType.metadata.MaxWeight = 200
	objNewDeviceType.metadata.MinWeight = 50

	//-----------add new device
	objByte, _ := json.Marshal(objNewDeviceType)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(objByte))
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		message := []byte(`{"msg":"Error while creating new device"}`)
		w.WriteHeader(500)
		w.Write(message)
		return
	}

	checkDestinationStatus := isDestinationExist(deviceType)
	if !checkDestinationStatus {
		//-------- create destination
		var objCreateDestination CreateDestination
		objCreateDestination.Name = deviceType
		objCreateDestination.Type = "cloudant"
		objCreateDestination.Configuration.BucketInterval = "DAY"

		createDestinationURL := IOTURL + "historianconnectors/615a95d64a0b1217f089043c/destinations"
		objByte, _ = json.Marshal(objCreateDestination)
		resp, err = http.Post(createDestinationURL, "application/json", bytes.NewBuffer(objByte))
		if err != nil {
			fmt.Println("error Body:", err.Error())
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			message := []byte(`{"msg":"Error while creating destination"}`)
			w.WriteHeader(500)
			w.Write(message)
			return
		}

		//-------- create forwarding rule
		var objCreateForwardingRule CreateForwardingRule
		objCreateForwardingRule.Name = deviceType + " rules"
		objCreateForwardingRule.DestinationName = deviceType
		objCreateForwardingRule.Type = "event"
		objCreateForwardingRule.Selector.DeviceType = deviceType
		objCreateForwardingRule.Selector.EventId = "HiveEvent"

		createForwardingURL := IOTURL + "historianconnectors/615a95d64a0b1217f089043c/forwardingrules"
		objByte, _ = json.Marshal(objCreateForwardingRule)
		resp, err = http.Post(createForwardingURL, "application/json", bytes.NewBuffer(objByte))
		if err != nil {
			fmt.Println("error Body:", err.Error())
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			message := []byte(`{"msg":"Error while creating forwarding rules"}`)
			w.WriteHeader(500)
			w.Write(message)
			return
		}
	}
	message := []byte(`{"msg":"Farmer is added"}`)
	w.WriteHeader(200)
	w.Write(message)
	w.Header().Set("Content-Type", "application/json")
}

// CreateNewDevie:
func CreateNewDevie(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceID = vars["deviceID"]
	url := IOTURL + "device/types/" + deviceType + "/devices"

	checkDeviceStatus := isDeviceExist(deviceType, deviceID)
	if checkDeviceStatus {
		message := []byte(`{"msg":"HiveID is already exist"}`)
		w.WriteHeader(400)
		w.Write(message)
		return
	}

	var objCreateNewDevice CreateNewDevice
	objCreateNewDevice.DeviceId = deviceID
	objCreateNewDevice.Metadata.MaxHumidity = 70
	objCreateNewDevice.Metadata.MinHumidity = 30
	objCreateNewDevice.Metadata.MaxTemperature = 45
	objCreateNewDevice.Metadata.MinTemperature = 35
	objCreateNewDevice.Metadata.MaxWeight = 200
	objCreateNewDevice.Metadata.MinWeight = 50

	//-----------add new device
	objByte, _ := json.Marshal(objCreateNewDevice)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(objByte))
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		message := []byte(`{"msg":"Error while creating new device"}`)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(message)
		return
	}

	message := []byte(`{"msg":"Farmer's Hive is added"}`)
	w.WriteHeader(200)
	w.Write(message)
	w.Header().Set("Content-Type", "application/json")
}

//GetDevieType
func GetDevieType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := IOTURL + "device/types/" + deviceType

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func isDeviceTypeExist(deviceType string) (status bool) {
	url := IOTURL + "device/types/" + deviceType
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status
}

func isDeviceExist(deviceType, deviceID string) (status bool) {
	url := IOTURL + "device/types/" + deviceType + "/devices/" + deviceID
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status
}

func isDestinationExist(deviceType string) (status bool) {
	serviceID := "615a95d64a0b1217f089043c"
	url := IOTURL + serviceID + "/destinations/" + deviceType
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		status = true
	}

	return status
}