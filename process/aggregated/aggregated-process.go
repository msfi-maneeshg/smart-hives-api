package aggregated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/cloudant-go-sdk/auth"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

func ProcessFarmerData(farmer string) {
	var currentDateTime = time.Now().UTC()

	if farmer == "" {
		fmt.Println("Error: Farmer value can not be empty!")
		return
	}

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
	currentDate := currentDateTime.Format("2006-01-02")
	if currentDateTime.Hour() > 0 && currentDateTime.Hour() < 6 {
		//process data for back date if process running in between 00:00AM to 06:00AM for last quatre of back date
		currentDate = currentDateTime.Add(-24 * time.Hour).Format("2006-01-02")
	}

	postAllDocsOptions := service.NewPostAllDocsOptions(
		IOT_DB_PREFIX + farmer + "_" + currentDate,
	)
	postAllDocsOptions.SetIncludeDocs(true)
	// postAllDocsOptions.SetLimit(1000)

	fmt.Println("INFO: Fetching records from Cloudant...")
	allDocsResult, _, err := service.PostAllDocs(postAllDocsOptions)
	if err != nil {
		panic(err)
	}

	var objGetDBData GetDBData
	b, _ := json.MarshalIndent(allDocsResult, "", "  ")
	json.Unmarshal(b, &objGetDBData)

	fmt.Println("INFO: Binding data...")
	var objHiveDataSets = make(map[string]HiveDataSet)

	timestamp := currentDateTime.Format("2006-01-02T15:04:05.000Z")
	quarterNoOfDayPart := GetQuarterNoOfDay(currentDateTime)

	for _, objData := range objGetDBData.Rows {
		eventTimestamp, _ := time.Parse("2006-01-02T15:04:05.000Z", objData.Doc.Timestamp)
		quarterNoOfDayPartForEvent := GetQuarterNoOfDay(eventTimestamp.Add(6 * time.Hour))
		if objData.Doc.DeviceID == "" || quarterNoOfDayPartForEvent != quarterNoOfDayPart {
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

		//------------create a new document
		eventDoc := cloudantv1.Document{}
		eventDoc.SetProperty("type", "event")
		eventDoc.SetProperty("deviceID", deviceID)
		eventDoc.SetProperty("date", timestamp)
		eventDoc.SetProperty("data", objHiveDataSet.HiveEventData)

		eventDoc.SetProperty("avgTemperature", objHiveDataSet.AvgTemperature)
		eventDoc.SetProperty("maxTemperature", objHiveDataSet.MaxTemperature)
		eventDoc.SetProperty("minTemperature", objHiveDataSet.MinTemperature)

		eventDoc.SetProperty("avgHumidity", objHiveDataSet.AvgHumidity)
		eventDoc.SetProperty("maxHumidity", objHiveDataSet.MaxHumidity)
		eventDoc.SetProperty("minHumidity", objHiveDataSet.MinHumidity)

		eventDoc.SetProperty("avgWeight", objHiveDataSet.AvgWeight)
		eventDoc.SetProperty("maxWeight", objHiveDataSet.MaxWeight)
		eventDoc.SetProperty("minWeight", objHiveDataSet.MinWeight)

		putDocumentOptions := service.NewPutDocumentOptions(
			IOT_DB_PREFIX+farmer,
			strings.ReplaceAll(currentDate+"_"+deviceID+"_"+quarterNoOfDayPart, "-", "_"),
		)

		putDocumentOptions.SetDocument(&eventDoc)

		fmt.Println("INFO: Storing aggregated data into cloudant for " + deviceID + "...")
		_, _, err := service.PutDocument(putDocumentOptions)
		if err != nil {
			fmt.Println(err.Error())
		}

	}
}

func GetQuarterNoOfDay(currentDateTime time.Time) string {
	var quarterNumber string
	currentDateTime = currentDateTime.Add(-6 * time.Hour) // for last 6 hour
	if currentDateTime.Hour() < 6 {
		quarterNumber = "Q1"
	} else if currentDateTime.Hour() < 12 {
		quarterNumber = "Q2"
	} else if currentDateTime.Hour() < 18 {
		quarterNumber = "Q3"
	} else if currentDateTime.Hour() < 24 {
		quarterNumber = "Q4"
	}

	return quarterNumber
}
