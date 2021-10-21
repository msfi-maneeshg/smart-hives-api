package api

import (
	"bytes"
	"context"
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

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

// GetHourlyInsight:
func GetHourlyInsight(w http.ResponseWriter, r *http.Request) {
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	vars := mux.Vars(r)
	var date = vars["date"]
	var period = vars["period"]

	collection := database.Data.Collection(userInfo.Username)
	filterData := bson.D{
		{"date", date},
		{"period", period},
	}

	cur, err := collection.Find(context.Background(), filterData)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cur.Close(context.Background())

	var allData []GetHiveData
	for cur.Next(context.Background()) {
		// To decode into a struct, use cursor.Decode()
		var result GetHiveData
		err := cur.Decode(&result)
		if err != nil {
			common.APIResponse(w, http.StatusInternalServerError, "While processing receivied data:"+err.Error())
			return
		}

		allData = append(allData, result)
	}
	if err := cur.Err(); err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "While processing receivied data:"+err.Error())
		return
	}

	common.APIResponse(w, http.StatusOK, allData)
}

// GetDeviceType:
func GetDeviceType(w http.ResponseWriter, r *http.Request) {
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	url := common.IOT_URL + "device/types/" + userInfo.Username

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}

	common.APIResponse(w, resp.StatusCode, body)
	return
}

// CreateNewDevice:
func CreateNewDevice(w http.ResponseWriter, r *http.Request) {
	var err error
	var objCreateNewDevice NewDeviceInfo
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}
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

	checkDeviceStatus, err := isDeviceExist(userInfo.Username, objCreateNewDevice.DeviceId)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	if checkDeviceStatus {
		common.APIResponse(w, http.StatusBadRequest, "DeviceID is already exist!")
		return
	}

	//-----------add new device
	url := common.IOT_URL + "device/types/" + userInfo.Username + "/devices"
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

// GetDeviceList:
func GetDeviceList(w http.ResponseWriter, r *http.Request) {
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	url := common.IOT_URL + "device/types/" + userInfo.Username + "/devices"

	resp, err := http.Get(url)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var objDevicesInfo DevicesInfo

	err = json.Unmarshal(body, &objDevicesInfo)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(objDevicesInfo.Results) == 0 {
		common.APIResponse(w, http.StatusNotFound, "No device found")
		return
	}

	common.APIResponse(w, resp.StatusCode, objDevicesInfo)
}

// GetDeviceInfo:
func GetDeviceInfo(w http.ResponseWriter, r *http.Request) {

	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	vars := mux.Vars(r)
	var deviceID = vars["deviceID"]

	url := common.IOT_URL + "device/types/" + userInfo.Username + "/devices/" + deviceID

	resp, err := http.Get(url)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		message := "DeviceID not Found"
		common.APIResponse(w, http.StatusNotFound, message)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.APIResponse(w, resp.StatusCode, body)
}

// DeleteDeviceInfo:
func DeleteDeviceInfo(w http.ResponseWriter, r *http.Request) {
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	vars := mux.Vars(r)
	var deviceID = vars["deviceID"]

	url := common.IOT_URL + "device/types/" + userInfo.Username + "/devices/" + deviceID

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		message := "Somwthing went wrong!"
		common.APIResponse(w, http.StatusInternalServerError, message)
		return
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		message := "Something went wrong!"
		common.APIResponse(w, http.StatusInternalServerError, message)
		return
	}
	defer resp.Body.Close()

	message := "Device has been removed!"
	common.APIResponse(w, http.StatusOK, message)
}

// UpdateDeviceInfo:
func UpdateDeviceInfo(w http.ResponseWriter, r *http.Request) {
	var err error
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	var objCreateNewDevice NewDeviceInfo
	vars := mux.Vars(r)
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

	checkDeviceStatus, err := isDeviceExist(userInfo.Username, deviceID)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	if !checkDeviceStatus {
		common.APIResponse(w, http.StatusBadRequest, "DeviceID not found!")
		return
	}

	//-----------add new device
	url := common.IOT_URL + "device/types/" + userInfo.Username + "/devices/" + deviceID
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

	objProfile := GetFarmerProfile(objFarmerProfileDetails.Email, "")
	if objProfile != (FarmerProfileDetails{}) {
		common.APIResponse(w, http.StatusBadRequest, "Email address already used.")
		return
	}

	objProfile = GetFarmerProfile("", objFarmerProfileDetails.Username)
	if objProfile != (FarmerProfileDetails{}) {
		common.APIResponse(w, http.StatusBadRequest, "Username already used.")
		return
	}

	deviceTypeStatus, err := isDeviceTypeExist(objFarmerProfileDetails.Username)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
	if deviceTypeStatus {
		common.APIResponse(w, http.StatusBadRequest, "Username is already exist")
		return
	}

	//--------create profile
	err = CreateNewProfile(objFarmerProfileDetails)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "There is an error while creating ptofile :"+err.Error())
		return
	}

	createNewDeviceType(w, r, objFarmerProfileDetails.Username)

}

// UpdatePassword:
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userInfo, err := CheckUserToken(r)
	if err != nil {
		common.APIResponse(w, http.StatusForbidden, err.Error())
		return
	}

	var objFarmerProfileDetails FarmerProfileDetails
	//------check body request
	if r.Body == nil {
		common.APIResponse(w, http.StatusBadRequest, "Request body can not be blank")
		return
	}
	err = json.NewDecoder(r.Body).Decode(&objFarmerProfileDetails)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "Error:"+err.Error())
		return
	}

	if objFarmerProfileDetails.Password == "" {
		common.APIResponse(w, http.StatusBadRequest, "Invalid password")
		return
	}

	isStrong := common.IsPasswordStrong(objFarmerProfileDetails.Password)
	if !isStrong {
		common.APIResponse(w, http.StatusBadRequest, "Password is not strong")
		return
	}

	objProfile := GetFarmerProfile(userInfo.Email, "")
	if objProfile == (FarmerProfileDetails{}) {
		common.APIResponse(w, http.StatusBadRequest, "Something went wrong.")
		return
	}

	objProfile = GetFarmerProfile("", userInfo.Username)
	if objProfile == (FarmerProfileDetails{}) {
		common.APIResponse(w, http.StatusBadRequest, "Something went wrong.")
		return
	}
	objFarmerProfileDetails.Email = userInfo.Email
	objFarmerProfileDetails.Username = userInfo.Username

	//--------create profile
	err = UpdateUserPassword(objFarmerProfileDetails)
	if err != nil {
		common.APIResponse(w, http.StatusBadRequest, "There is an error while updating password :"+err.Error())
		return
	}

	common.APIResponse(w, http.StatusOK, "Password has been updated successfully!")
}

// Login:
func Login(w http.ResponseWriter, r *http.Request) {

	var objFarmerProfileDetails FarmerProfileDetails
	var objUserSession UserSession

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

	objProfile := GetFarmerProfile(objFarmerProfileDetails.Username, "")
	if objProfile == (FarmerProfileDetails{}) {
		objProfile = GetFarmerProfile("", objFarmerProfileDetails.Username)
		if objProfile == (FarmerProfileDetails{}) {
			common.APIResponse(w, http.StatusBadRequest, "Invalid Credential.")
			return
		}
	}

	if objProfile.Password != objFarmerProfileDetails.Password {
		common.APIResponse(w, http.StatusBadRequest, "Invalid Credential.")
		return
	}

	objUserSession, err = createUserToken(objProfile)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Something went wrong Error:"+err.Error())
		return
	}

	objUserSession.Email = objProfile.Email
	objUserSession.Username = objProfile.Username

	common.APIResponse(w, http.StatusOK, objUserSession)
}

// RefreshToken:
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("refreshToken")
	if refreshToken != "" {
		token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%v", "There was an error")
			}
			return []byte(common.REFERESH_KEY), nil
		})

		if err != nil {
			common.APIResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		var objUserSession UserSession
		objUserSession.RefereshToken = refreshToken
		if token.Valid && ok {
			newToken := jwt.New(jwt.SigningMethodHS256)
			newClaims := newToken.Claims.(jwt.MapClaims)

			objUserSession.Username, _ = claims["username"].(string)
			objUserSession.Email, _ = claims["email"].(string)

			newClaims["authorized"] = true
			newClaims["username"] = objUserSession.Username
			newClaims["email"] = objUserSession.Email
			newClaims["exp"] = time.Now().Add(time.Minute * common.EXPIRE_TIME).Unix()

			objUserSession.UserToken, err = newToken.SignedString([]byte(common.MY_KEY))
			if err != nil {
				common.APIResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			common.APIResponse(w, http.StatusOK, objUserSession)
			return
		}
	}
	common.APIResponse(w, http.StatusBadRequest, "Invalid refresh token.")
}

func createNewDeviceType(w http.ResponseWriter, r *http.Request, deviceType string) {
	url := common.IOT_URL + "device/types"

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

	checkDestinationStatus, err := isDestinationExist(deviceType)
	if err != nil {
		common.APIResponse(w, http.StatusInternalServerError, "Error:"+err.Error())
		return
	}
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

	createPhysicalInterface(w, r, deviceType)
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

	output_body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	common.APIResponse(w, res.StatusCode, output_body)
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

func createPhysicalInterface(w http.ResponseWriter, r *http.Request, deviceType string) {
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
		connectEventTypeWithPI(w, r, deviceType, objOutputInterfaceInfo.ID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectEventTypeWithPI(w http.ResponseWriter, r *http.Request, deviceType, physicalInterfaceID string) {
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
		connectDeviceTypeWithPI(w, r, deviceType, physicalInterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectDeviceTypeWithPI(w http.ResponseWriter, r *http.Request, deviceType, physicalInterfaceID string) {
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
		createLogicalInterface(w, r, deviceType)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func createLogicalInterface(w http.ResponseWriter, r *http.Request, deviceType string) {
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
		connectDeviceTypeWithLI(w, r, deviceType, objOutputInterfaceInfo.ID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func connectDeviceTypeWithLI(w http.ResponseWriter, r *http.Request, deviceType, logicalinterfaceID string) {
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
		defineMapping(w, r, deviceType, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func defineMapping(w http.ResponseWriter, r *http.Request, deviceType, logicalinterfaceID string) {

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
		addNotificationRules(w, r, deviceType, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func addNotificationRules(w http.ResponseWriter, r *http.Request, deviceType, logicalinterfaceID string) {
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

	activateInterface(w, r, deviceType, logicalinterfaceID)
}

func activateInterface(w http.ResponseWriter, r *http.Request, deviceType, logicalinterfaceID string) {

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
		addActionTrigger(w, r, deviceType, logicalinterfaceID)
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}

func addActionTrigger(w http.ResponseWriter, r *http.Request, deviceType, logicalinterfaceID string) {
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
		common.APIResponse(w, http.StatusCreated, "Registration successfully completed!")
		return
	} else {
		common.APIErrorResponse(w, resp.StatusCode, body)
		return
	}
}
