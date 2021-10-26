package main

import (
	"fmt"
	"os"
	"smart-hives/process/aggregated"
	"smart-hives/process/database"
	"time"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Warning! Please send some command arguments.")
		//return
	}
	//args = args[1:]
	// Validating commond arguments
	// for _, objArgument := range args {
	// 	if strings.Contains(objArgument, "-farmer=") {
	// 		farmer = strings.Replace(objArgument, "-farmer=", "", 1)
	// 	}
	// }
	database.ConnectDatabase("smart-hives")
	current := time.Now().Minute()
	if current < 30 {
		current = 30 - current
	} else {
		current = 90 - current
	}

	for {
		aggregated.ProcessFarmerData("")
		fmt.Println("Waiting for next ", current, "Minutes to process...")
		time.Sleep(time.Duration(current) * time.Minute)
		if current != 60 {
			current = 60
		}
	}
}
