package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"smart-hives/api/common"
	"smart-hives/api/database"
	"strings"
	"time"

	"github.com/IBM/cloudant-go-sdk/auth"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// ProcessFarmerData:
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

// ProcessedFarmerData:
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

// GetDeviceLastEvent:
func GetDeviceLastEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceId = vars["deviceId"]

	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceId + "/events"

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

// GetDeviceTypes:
func GetDeviceTypes(w http.ResponseWriter, r *http.Request) {
	url := common.IOT_URL + "device/types"

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

// GetDevices:
func GetDevices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := common.IOT_URL + "device/types/" + deviceType + "/devices"

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

// CreateNewDeviceType:
func CreateNewDeviceType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := common.IOT_URL + "device/types"
	deviceTypeStatus := isDeviceTypeExist(deviceType)
	if deviceTypeStatus {
		common.APIResponse(w, http.StatusBadRequest, "FarmerID is already exist")
		return
	}

	var objNewDeviceType NewDeviceType
	objNewDeviceType.ID = deviceType
	objNewDeviceType.ClassId = "Device"
	objNewDeviceType.Description = "Hives for " + deviceType

	//-----------add new device
	objByte, _ := json.Marshal(objNewDeviceType)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error while creating new device")
		return
	}

	checkDestinationStatus := isDestinationExist(deviceType)
	if !checkDestinationStatus {
		//-------- create destination
		var objCreateDestination CreateDestination
		objCreateDestination.Name = deviceType
		objCreateDestination.Type = "cloudant"
		objCreateDestination.Configuration.BucketInterval = "DAY"

		createDestinationURL := common.IOT_URL + "historianconnectors/615a95d64a0b1217f089043c/destinations"
		objByte, _ = json.Marshal(objCreateDestination)
		resp, err = http.Post(createDestinationURL, "application/json", bytes.NewBuffer(objByte))
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error while creating destination")
			return
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error while creating destination")
			return
		}

		//-------- create forwarding rule
		var objCreateForwardingRule CreateForwardingRule
		objCreateForwardingRule.Name = deviceType + " rules"
		objCreateForwardingRule.DestinationName = deviceType
		objCreateForwardingRule.Type = "event"
		objCreateForwardingRule.Selector.DeviceType = deviceType
		objCreateForwardingRule.Selector.EventId = "HiveEvent"

		createForwardingURL := common.IOT_URL + "historianconnectors/615a95d64a0b1217f089043c/forwardingrules"
		objByte, _ = json.Marshal(objCreateForwardingRule)
		resp, err = http.Post(createForwardingURL, "application/json", bytes.NewBuffer(objByte))
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error while creating forwardingrules")
			return
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error while creating forwardingrules")
			return
		}
	}

	createPhysicalInterface(w, r)
}

func createEventSchema(w http.ResponseWriter, r *http.Request) {
	endpoint := common.IOT_URL + "draft/schemas"
	// New multipart writer.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormField("name")
	if err != nil {
		fmt.Println("0:", err.Error())
	}
	_, err = io.Copy(fw, strings.NewReader("HiveEventSchema"))
	if err != nil {
		fmt.Println("1:", err.Error())
	}

	fw, err = writer.CreateFormFile("schemaFile", "eventTypeSchema.json")
	if err != nil {
		fmt.Println("2:", err.Error())
	}
	file, err := os.Open("eventTypeSchema.json")
	if err != nil {
		fmt.Println("3:", err.Error())
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		fmt.Println("4:", err.Error())
	}
	// Close multipart writer.
	writer.Close()

	client := &http.Client{}
	r, err = http.NewRequest("POST", endpoint, bytes.NewReader(body.Bytes())) // URL-encoded payload
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Set("Content-Type", "multipart/form-data")

	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.Status)
	defer res.Body.Close()

	body1, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(res.StatusCode)
	w.Write(body1)
	return
}

func createEventType(w http.ResponseWriter, r *http.Request) {
	url := common.IOT_URL + "draft/event/types/"

	var objEventType CreateInterface
	objEventType.Name = "HiveEvent"
	objEventType.SchemaId = common.SCHEMA_ID

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objEventType)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	common.APIResponse(w, http.StatusOK, body)
	return
}

func createPhysicalInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/physicalinterfaces"

	var objCreateInterface CreateInterface
	objCreateInterface.Name = deviceType + "_PI"

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusCreated {
		var objOutputInterfaceInfo OutputInterfaceInfo
		_ = json.Unmarshal(body, &objOutputInterfaceInfo)

		connectEventTypeWithPI(w, r, objOutputInterfaceInfo.ID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectEventTypeWithPI(w http.ResponseWriter, r *http.Request, physicalInterfaceID string) {
	url := common.IOT_URL + "draft/physicalinterfaces/" + physicalInterfaceID + "/events"

	var objCreateInterface CreateInterface
	objCreateInterface.EventId = "HiveEvent"
	objCreateInterface.EventTypeId = common.EVENT_TYPE_ID

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusCreated {
		connectDeviceTypeWithPI(w, r, physicalInterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectDeviceTypeWithPI(w http.ResponseWriter, r *http.Request, physicalInterfaceID string) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/device/types/" + deviceType + "/physicalinterface"

	var objCreateInterface CreateInterface
	objCreateInterface.ID = physicalInterfaceID

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusCreated {
		createLogicalInterface(w, r)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func createLogicalInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/logicalinterfaces/"

	var objCreateInterface CreateInterface
	objCreateInterface.Name = deviceType + "_LI"
	objCreateInterface.Alias = deviceType + "_LI"
	objCreateInterface.SchemaId = common.SCHEMA_ID

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusCreated {

		var objOutputInterfaceInfo OutputInterfaceInfo
		_ = json.Unmarshal(body, &objOutputInterfaceInfo)

		connectDeviceTypeWithLI(w, r, objOutputInterfaceInfo.ID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectDeviceTypeWithLI(w http.ResponseWriter, r *http.Request, logicalinterfaceID string) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/device/types/" + deviceType + "/logicalinterfaces"

	var objCreateInterface CreateInterface
	objCreateInterface.ID = logicalinterfaceID

	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}
	if resp.StatusCode == http.StatusCreated {
		defineMapping(w, r, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func defineMapping(w http.ResponseWriter, r *http.Request, logicalinterfaceID string) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/device/types/" + deviceType + "/mappings"

	var objCreateInterface CreateInterface
	objCreateInterface.LogicalInterfaceId = logicalinterfaceID
	objCreateInterface.NotificationStrategy = "on-state-change"
	var objPropertyMappings PropertyMappings
	objPropertyMappings.HiveEvent.Humidity = "$event.humidity"
	objPropertyMappings.HiveEvent.Weight = "$event.weight"
	objPropertyMappings.HiveEvent.Temperature = "$event.temperature"
	objCreateInterface.PropertyMappings = &objPropertyMappings
	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objCreateInterface)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusCreated {
		addNotificationRules(w, r, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func addNotificationRules(w http.ResponseWriter, r *http.Request, logicalinterfaceID string) {

	url := common.IOT_URL + "draft/logicalinterfaces/" + logicalinterfaceID + "/rules"

	allNotificationRules := []NotificationRules{
		{
			Name:      "minimumTemperature",
			Condition: "$state.temperature < $instance.metadata.minimumTemperature",
		},
		{
			Name:      "maximumTemperature",
			Condition: "$state.temperature > $instance.metadata.maximumTemperature",
		},
		{
			Name:      "minimumHumidity",
			Condition: "$state.humidity < $instance.metadata.minimumHumidity",
		},
		{
			Name:      "maximumHumidity",
			Condition: "$state.humidity > $instance.metadata.maximumHumidity",
		},
	}

	for _, objNotificationRules := range allNotificationRules {
		objNotificationRules.NotificationStrategy.When = "x-in-y"
		objNotificationRules.NotificationStrategy.Count = 1
		objNotificationRules.NotificationStrategy.TimePeriod = 60
		// Create client
		client := &http.Client{}

		objByte, _ := json.Marshal(objNotificationRules)

		// Create request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
			return
		}
		req.Header.Add("Content-Type", "application/json")

		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
			return
		}
		if resp.StatusCode != http.StatusCreated {
			common.APIErrorResponse(w, resp.StatusCode, body)
			//return
		}
	}

	activateInterface(w, r, logicalinterfaceID)
}

func activateInterface(w http.ResponseWriter, r *http.Request, logicalinterfaceID string) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "draft/device/types/" + deviceType

	var objActivateInterface ActivateInterface
	objActivateInterface.Operation = "activate-configuration"
	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objActivateInterface)

	// Create request
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}

	if resp.StatusCode == http.StatusAccepted {
		addActionTrigger(w, r, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func addActionTrigger(w http.ResponseWriter, r *http.Request, logicalinterfaceID string) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	url := common.IOT_URL + "actions/" + common.ACTION_ID + "/triggers"

	var objActionTrigger ActionTrigger
	objActionTrigger.Name = deviceType + " Trigger"
	objActionTrigger.Description = "Call notification action"
	objActionTrigger.Type = "rule"
	objActionTrigger.Enabled = "true"
	objActionTrigger.Configuration.LogicalInterfaceId = logicalinterfaceID
	objActionTrigger.Configuration.RuleId = "*"
	objActionTrigger.Configuration.InstanceId = "*"
	objActionTrigger.Configuration.Type = "*"
	objActionTrigger.Configuration.TypeId = "*"
	objActionTrigger.VariableMappings.DeviceType = "$event.typeId"
	objActionTrigger.VariableMappings.DeviceId = "$event.instanceId"
	objActionTrigger.VariableMappings.Temperature = "$event.state.temperature"
	objActionTrigger.VariableMappings.Humidity = "$event.state.humidity"
	objActionTrigger.VariableMappings.Weight = "$event.state.weight"
	objActionTrigger.VariableMappings.InterfaceId = "$event.logicalInterfaceId"
	objActionTrigger.VariableMappings.Timestamp = "$event.timestamp"
	// Create client
	client := &http.Client{}

	objByte, _ := json.Marshal(objActionTrigger)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Error:"+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error"+err.Error())
		return
	}
	if resp.StatusCode == http.StatusCreated {
		common.APIResponse(w, http.StatusCreated, "Farmer is added!")
		return
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

// CreateNewDevice:
func CreateNewDevice(w http.ResponseWriter, r *http.Request) {
	var err error
	var objCreateNewDevice NewDeviceInfo
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]

	//------check body request
	if r.Body == nil {
		common.APIResponse(w, http.StatusBadRequest, "Request body can not be blank")
		return
	}
	err = json.NewDecoder(r.Body).Decode(&objCreateNewDevice)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "Error:"+err.Error())
		return
	}

	if objCreateNewDevice.DeviceId == "" {
		common.APIResponse(w, http.StatusBadRequest, "DeviceID can not be empty!")
		return
	}

	checkDeviceStatus := isDeviceExist(deviceType, objCreateNewDevice.DeviceId)
	if checkDeviceStatus {
		common.APIResponse(w, http.StatusBadRequest, "DeviceID is already exist!")
		return
	}

	//-----------add new device
	url := common.IOT_URL + "device/types/" + deviceType + "/devices"
	objByte, _ := json.Marshal(objCreateNewDevice)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(objByte))
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error while creating new device")
		return
	}

	common.APIResponse(w, http.StatusOK, "Farmer's Device is added!")
}

// GetDeviceType:
func GetDeviceType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := common.IOT_URL + "device/types/" + deviceType

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

// GetDeviceList:
func GetDeviceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	url := common.IOT_URL + "device/types/" + deviceType + "/devices"

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

// GetDeviceInfo:
func GetDeviceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceID = vars["deviceID"]

	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		message := []byte(`{"msg":"DeviceID not Found"}`)
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(message)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// DeleteDeviceInfo:
func DeleteDeviceInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceID = vars["deviceID"]

	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		message := []byte(`{"msg":"Somwthing went wrong!"}`)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write(message)
		return
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		message := []byte(`{"msg":"Something went wrong!"}`)
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(message)
		return
	}
	defer resp.Body.Close()

	message := []byte(`{"msg":"Device has been removed!"}`)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(message)
}

// UpdateDeviceInfo:
func UpdateDeviceInfo(w http.ResponseWriter, r *http.Request) {
	var err error
	var objCreateNewDevice NewDeviceInfo
	vars := mux.Vars(r)
	var deviceType = vars["deviceType"]
	var deviceID = vars["deviceID"]

	//------check body request
	if r.Body == nil {
		common.APIResponse(w, http.StatusBadRequest, "Request body can not be blank")
		return
	}
	err = json.NewDecoder(r.Body).Decode(&objCreateNewDevice)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "Error:"+err.Error())
		return
	}

	checkDeviceStatus := isDeviceExist(deviceType, deviceID)
	if !checkDeviceStatus {
		common.APIResponse(w, http.StatusBadRequest, "DeviceID not found!")
		return
	}

	//-----------add new device
	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID
	objByte, _ := json.Marshal(objCreateNewDevice)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(objByte))
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Somwthing went wrong!")
		return
	}

	req.Header.Add("Content-Type", "application/json")

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		common.APIResponse(w, resp.StatusCode, "Somwthing went wrong!")
		return
	}
	defer resp.Body.Close()

	common.APIResponse(w, http.StatusOK, "Device info has been updated!")
}

func isDeviceTypeExist(deviceType string) (status bool) {
	url := common.IOT_URL + "device/types/" + deviceType
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
	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID
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
	url := common.IOT_URL + serviceID + "/destinations/" + deviceType
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

// Register:
func Register(w http.ResponseWriter, r *http.Request) {
	var objFarmerProfileDetails FarmerProfileDetails

	//------check body request
	if r.Body == nil {
		common.APIResponse(w, http.StatusBadRequest, "Request body can not be blank")
		return
	}

	err := json.NewDecoder(r.Body).Decode(&objFarmerProfileDetails)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "Error:"+err.Error())
		return
	}

	if objFarmerProfileDetails.Username == "" {
		common.APIResponse(w, http.StatusBadRequest, "Invalid username")
		return
	}

	if objFarmerProfileDetails.Password == "" {
		common.APIResponse(w, http.StatusBadRequest, "Invalid password")
		return
	}

	if objFarmerProfileDetails.Email == "" {
		common.APIResponse(w, http.StatusBadRequest, "Invalid email")
		return
	}

	isValid := common.IsEmailValid(objFarmerProfileDetails.Email)
	if !isValid {
		common.APIResponse(w, http.StatusBadRequest, "Invalid email address")
		return
	}

	isStrong := common.IsPasswordStrong(objFarmerProfileDetails.Password)
	if !isStrong {
		common.APIResponse(w, http.StatusBadRequest, "Password is not strong")
		return
	}

	objProfile := GetFarmerProfile(objFarmerProfileDetails.Email)
	if objProfile != (FarmerProfileDetails{}) {
		common.APIResponse(w, http.StatusBadRequest, "Email address already used.")
		return
	}

	//--------create profile
	err = CreateNewProfile(objFarmerProfileDetails)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "There is an error while creating ptofile :"+err.Error())
		return
	}

	common.APIResponse(w, http.StatusCreated, "Profile Created.")
}
