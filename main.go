package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	key := "iotp_8l173e_farmer-7_"
	dt, _ := time.Parse("2006-01-02", "2021-10-04")
	currentDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	for {
		dt = dt.Add(24 * time.Hour)
		dtstr2 := dt.Format("2006-01-02")

		url := "https://apikey-v2-29mnuuarysnz6zwv1np8fzp808a5e4052m4783hjkflh:993856ca873efb33cc67fc6c82d6c7e8@433c346a-cb7c-4736-8e95-0bc99303fe1a-bluemix.cloudant.com/" + key + dtstr2

		// Create client
		client := &http.Client{}

		// Create request
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(resp.Status, " for ", key+dtstr2)

		if dtstr2 == currentDate {
			break
		}
	}

}
