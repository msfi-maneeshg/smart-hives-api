package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const IOTURL = "https://a-8l173e-otjztnyacu:ChLq7u0pO+*hl7JER_@8l173e.internetofthings.ibmcloud.com/api/v0002/"
const DEVICE_AUTH_TOKEN = "smarthives@12345"
const DEVICE_AUTH_USER = "use-token-auth"

type BulkDeviceOutput struct {
	Results []BulkDeviceResult `json:"results"`
}

type BulkDeviceResult struct {
	TypeId   string `json:"typeId"`
	DeviceId string `json:"deviceId"`
}

type Payload struct {
	Temperature int `json:"temperature"`
	Humidity    int `json:"humidity"`
	Weight      int `json:"weight"`
}

var ActiveDevices = map[string]bool{}

func main() {
	var wg sync.WaitGroup
	runningThread := map[string]bool{}

	for {
		//-------fatching list of devicetypes
		objBulkDeviceOutput := getDeviceTypes()
		if len(objBulkDeviceOutput.Results) == 0 {
			fmt.Println("Alert: No device found!")
			return
		}

		allDevices := map[string]bool{}

		for i, deviceInfo := range objBulkDeviceOutput.Results {
			ActiveDevices[deviceInfo.DeviceId+":"+deviceInfo.TypeId] = true
			allDevices[deviceInfo.DeviceId+":"+deviceInfo.TypeId] = true
			if okay := runningThread[deviceInfo.DeviceId+":"+deviceInfo.TypeId]; okay {
				continue
			}
			runningThread[deviceInfo.DeviceId+":"+deviceInfo.TypeId] = true
			wg.Add(1)
			go publishEvent(i, deviceInfo, &wg)
			time.Sleep(1 * time.Second)

		}
		for ID := range ActiveDevices {
			if _, okay := allDevices[ID]; !okay {
				delete(ActiveDevices, ID)
				delete(runningThread, ID)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func getDeviceTypes() (objBulkDeviceOutput BulkDeviceOutput) {
	url := IOTURL + "bulk/devices"
	// url := IOTURL + "bulk/devices?typeId=farmer-1"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error Body:", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("malformed input")
	}

	if resp.StatusCode == http.StatusOK {
		_ = json.Unmarshal(body, &objBulkDeviceOutput)
		return objBulkDeviceOutput

	} else {
		fmt.Println("Response error:", string(body))
	}
	return objBulkDeviceOutput
}

func publishEvent(i int, deviceInfo BulkDeviceResult, wg *sync.WaitGroup) {
	defer wg.Done()
	topicName := "iot-2/evt/HiveEvent/fmt/json"
	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://8l173e.messaging.internetofthings.ibmcloud.com:1883")
	opts.SetClientID("d:8l173e:" + deviceInfo.TypeId + ":" + deviceInfo.DeviceId)
	opts.SetUsername(DEVICE_AUTH_USER)
	opts.SetPassword(DEVICE_AUTH_TOKEN)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	var objPayload Payload
	initialWeight := 50.0
	for {
		objPayload.Humidity = getRandomNumber(30, 70)
		objPayload.Temperature = getRandomNumber(30, 50)
		objPayload.Weight = int(initialWeight)
		objPayloadByte, _ := json.Marshal(objPayload)
		fmt.Println("EventData send for ("+deviceInfo.TypeId+","+deviceInfo.DeviceId+")", string(objPayloadByte))
		token := c.Publish(topicName, 0, false, string(objPayloadByte))
		token.Wait()
		initialWeight = initialWeight + (float64(getRandomNumber(0, 50)) / float64(100))
		if initialWeight > 200 {
			initialWeight = 50.0
		}

		if _, okay := ActiveDevices[deviceInfo.DeviceId+":"+deviceInfo.TypeId]; !okay {
			break
		}
		time.Sleep(20 * time.Second)
	}
}

func getRandomNumber(min, max int) int {
	return rand.Intn(max-min) + min
}
