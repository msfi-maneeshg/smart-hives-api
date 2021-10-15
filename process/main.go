package main

import (
	"fmt"
	"os"
	"smart-hives/process/aggregated"
	"smart-hives/process/database"
	"strings"
	"time"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Warning! Please send some command arguments.")
		return
	}
	args = args[1:]
	var farmer string
	// Validating commond arguments
	for _, objArgument := range args {
		if strings.Contains(objArgument, "-farmer=") {
			farmer = strings.Replace(objArgument, "-farmer=", "", 1)
		}
	}

	database.ConnectDatabase("smart-hives")
	for {
		aggregated.ProcessFarmerData(farmer)
		time.Sleep(30 * time.Minute)
	}

}
