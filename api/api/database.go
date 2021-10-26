package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"smart-hives/api/common"
	"smart-hives/api/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// =============================== Internal DB Related Functions ===============================

//GetFarmerProfile :
func GetFarmerProfile(emailID, username string) (objProfile FarmerProfileDetails) {
	collection := database.Data.Collection(common.PROFILES)
	filterData := bson.D{
		{"email", emailID},
	}
	if username != "" {
		filterData = bson.D{
			{"username", username},
		}
	}

	err := collection.FindOne(context.TODO(), filterData).Decode(&objProfile)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Fatal(err)
	}

	return objProfile
}

//CreateNewProfile :
func CreateNewProfile(objProfile FarmerProfileDetails) (err error) {
	collection := database.Data.Collection(common.PROFILES)

	_, err = collection.InsertOne(context.TODO(), objProfile)
	if err != nil {
		return err
	}
	return nil
}

//UpdateUserPassword :
func UpdateUserPassword(objProfile FarmerProfileDetails) (err error) {
	collection := database.Data.Collection(common.PROFILES)

	opts := options.Update().SetUpsert(false)
	filter := bson.D{{"username", objProfile.Username}, {"email", objProfile.Email}}
	update := bson.D{{"$set", bson.D{{"password", objProfile.Password}}}}

	result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("%v", "username/email not found")
	}
	return nil
}

//DeleteDeviceData :
func DeleteDeviceData(tableName, deviceID string) (err error) {
	collection := database.Data.Collection(tableName)

	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})

	_, err = collection.DeleteOne(context.TODO(), bson.D{{"deviceID", deviceID}}, opts)
	if err != nil {
		return err
	}

	return nil
}

// =============================== IoT Related Functions ===============================

//GetDevicesForIoT: Getting list of devices from IoT for a specific device type.
func GetDevicesForIoT(deviceType string) (objDevicesInfo DevicesInfo, err error) {
	url := common.IOT_URL + "device/types/" + deviceType + "/devices"

	resp, err := http.Get(url)
	if err != nil {
		return objDevicesInfo, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return objDevicesInfo, err
	}

	err = json.Unmarshal(body, &objDevicesInfo)
	if err != nil {
		return objDevicesInfo, err
	}

	return objDevicesInfo, nil
}

//DeleteDeviceFromIoT: Deleting device from IoT by Device type and ID.
func DeleteDeviceFromIoT(deviceType, deviceID string) (err error) {
	url := common.IOT_URL + "device/types/" + deviceType + "/devices/" + deviceID

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}
	defer resp.Body.Close()

	return err
}

// =============================== Cloudant Related Functions ===============================

//GetDeviceNotificationFromCloudant: Deleting device notification from Cloudant by Device type and ID.
func GetDeviceNotificationFromCloudant(deviceType string) (err error) {
	url := common.IBM_URL + "iotp-notification/_design/iotp/_view/by-deviceType?key=" + deviceType

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", common.IBM_AUTH)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}
	defer resp.Body.Close()

	return nil
}

//DeleteDeviceNotificationFromCloudant: Deleting device notification from Cloudant by Device type and ID.
func DeleteDeviceNotificationFromCloudant(docID, revID string) (err error) {
	url := common.IBM_URL + "iotp-notification/" + docID + "?rev=" + revID

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", common.IBM_AUTH)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%v", "Somwthing went wrong!")
	}
	defer resp.Body.Close()

	return nil
}
