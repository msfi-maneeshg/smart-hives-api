package aggregated

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"smart-hives/process/database"
	"strings"
	"sync"
	"time"

	"github.com/IBM/cloudant-go-sdk/auth"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const IOTURL = "https://a-62m15c-ubghzixbav:r+@6D*-wMzAw6U&4tA@62m15c.internetofthings.ibmcloud.com/api/v0002/"

func ProcessFarmerData(farmer string) {
	var currentDateTime = time.Now().UTC()

	if farmer == "" {
		fmt.Println("Error: Farmer value can not be empty!")
		return
	}

	//---------get the all iot device type
	resp, err := http.Get(IOTURL + "device/types")
	if err != nil {
		fmt.Println("Getting error while trying to get all device type from IOTP, Error:" + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Records not found!")
		return
	}

	bodyResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Getting error while trying to get read body response, Error:" + err.Error())
		return
	}

	var objDeviceTypeResultSet DeviceTypeResultSet

	err = json.Unmarshal(bodyResponse, &objDeviceTypeResultSet)
	if err != nil {
		fmt.Println("Getting error while trying to unmarshling body response, Error:" + err.Error())
		return
	}
	fmt.Println(objDeviceTypeResultSet)
	authenticator, err := auth.NewCouchDbSessionAuthenticator(
		CLOUDANT_USERNAME,
		CLOUDANT_PASSWORD,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("INFO: Connecting with cloudant service...")
	service, err := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName:   "CLOUDANT",
			URL:           CLOUDANT_URL,
			Authenticator: authenticator,
		},
	)
	if err != nil {
		fmt.Print(err)
	}

	var wg sync.WaitGroup
	for _, deviceTypeInfo := range objDeviceTypeResultSet.Result {
		if deviceTypeInfo.ID == "Farmer-1-Hives" {
			deviceTypeInfo.ID = "farmer-1"
		}
		wg.Add(1)
		go ProcessData(deviceTypeInfo.ID, currentDateTime, service, &wg)
	}
	wg.Wait()

	fmt.Println("Process is completed")
	duration := time.Since(currentDateTime)
	log.Print("Process took ", duration)
}

func RecoverProcessData() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
	}
}

func ProcessData(farmer string, currentDateTime time.Time, service *cloudantv1.CloudantV1, wg *sync.WaitGroup) {
	defer wg.Done()
	defer RecoverProcessData()
	currentDate := currentDateTime.Format("2006-01-02")
	postViewOptions := service.NewPostViewOptions(
		IOT_DB_PREFIX+farmer+"_"+currentDateTime.Add(-1*time.Hour).Format("2006-01-02"),
		"iotp",
		"by-date",
	)

	searchKeyDate := currentDateTime
	postViewOptions.SetIncludeDocs(true)
	postViewOptions.SetStartkey(searchKeyDate.Add(-1*time.Hour).Format("2006-01-02T15:00:00") + ".000Z")
	postViewOptions.SetEndkey(searchKeyDate.Format("2006-01-02T15:00:00") + ".000Z")
	fmt.Println("INFO: Fetching records from Cloudant for " + farmer + " ...")
	allDocsResult, _, err := service.PostView(postViewOptions)
	if err != nil {
		panic(err)
	}

	var objGetDBData GetDBData
	b, _ := json.MarshalIndent(allDocsResult, "", "  ")
	json.Unmarshal(b, &objGetDBData)

	fmt.Println("INFO: Binding data for " + farmer + "...")
	var objHiveDataSets = make(map[string]HiveDataSet)

	timestamp := currentDateTime.Format("2006-01-02T15:04:05.000Z")
	periodOfDay := fmt.Sprintf("%02d-%02d", currentDateTime.Add(-1*time.Hour).Hour(), currentDateTime.Hour())
	for _, objData := range objGetDBData.Rows {
		if objData.Doc.DeviceID == "" {
			continue
		}
		objHiveDataSet := objHiveDataSets[objData.Doc.DeviceID]

		objHiveDataSet.TotalRecords++
		temperature := objData.Doc.Data.Temperature
		humidity := objData.Doc.Data.Humidity
		weight := objData.Doc.Data.Weight

		objHiveDataSet.TotalTemperature += temperature
		objHiveDataSet.TotalHumidity += humidity
		objHiveDataSet.TotalWeight += weight

		if objHiveDataSet.MinTemperature == nil || *objHiveDataSet.MinTemperature > temperature {
			objHiveDataSet.MinTemperature = &temperature
		}
		if objHiveDataSet.MaxTemperature == nil || *objHiveDataSet.MaxTemperature < temperature {
			objHiveDataSet.MaxTemperature = &temperature
		}

		if objHiveDataSet.MinHumidity == nil || *objHiveDataSet.MinHumidity > humidity {
			objHiveDataSet.MinHumidity = &humidity
		}
		if objHiveDataSet.MaxHumidity == nil || *objHiveDataSet.MaxHumidity < humidity {
			objHiveDataSet.MaxHumidity = &humidity
		}

		if objHiveDataSet.MinWeight == nil || *objHiveDataSet.MinWeight > weight {
			objHiveDataSet.MinWeight = &weight
		}
		if objHiveDataSet.MaxWeight == nil || *objHiveDataSet.MaxWeight < weight {
			objHiveDataSet.MaxWeight = &weight
		}
		hiveEventData := objData.Doc.Data
		hiveEventData.Timestamp = objData.Doc.Timestamp

		objHiveDataSet.HiveEventData = append(objHiveDataSet.HiveEventData, hiveEventData)
		objHiveDataSets[objData.Doc.DeviceID] = objHiveDataSet

	}

	for deviceID, objHiveDataSet := range objHiveDataSets {
		objHiveDataSet.AvgTemperature = objHiveDataSet.TotalTemperature / objHiveDataSet.TotalRecords
		objHiveDataSet.AvgHumidity = objHiveDataSet.TotalHumidity / objHiveDataSet.TotalRecords
		objHiveDataSet.AvgWeight = objHiveDataSet.TotalWeight / objHiveDataSet.TotalRecords
		recordKey := strings.ReplaceAll(currentDate+"_"+deviceID+"_"+periodOfDay, "-", "_")
		finalInputData := bson.M{
			"_id":            recordKey,
			"type":           "event",
			"deviceID":       deviceID,
			"period":         periodOfDay,
			"date":           currentDate,
			"timestamp":      timestamp,
			"data":           objHiveDataSet.HiveEventData,
			"avgTemperature": objHiveDataSet.AvgTemperature,
			"maxTemperature": objHiveDataSet.MaxTemperature,
			"minTemperature": objHiveDataSet.MinTemperature,
			"avgHumidity":    objHiveDataSet.AvgHumidity,
			"maxHumidity":    objHiveDataSet.MaxHumidity,
			"minHumidity":    objHiveDataSet.MinHumidity,
			"avgWeight":      objHiveDataSet.AvgWeight,
			"maxWeight":      objHiveDataSet.MaxWeight,
			"minWeight":      objHiveDataSet.MinWeight,
		}
		collection := database.Data.Collection(farmer)
		DeleteOldHiveRecord(recordKey, collection)
		res, err := collection.InsertOne(context.Background(), finalInputData)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		id := res.InsertedID
		fmt.Println("Data is inserted for "+farmer+" at ID :", id)

	}
}

func DeleteOldHiveRecord(key string, collection *mongo.Collection) {
	filterKey := bson.D{{"_id", key}}
	_, err := collection.DeleteOne(context.Background(), filterKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
