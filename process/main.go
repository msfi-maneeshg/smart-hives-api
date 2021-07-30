package main

import (
	"fmt"
	"os"
	"smart-hives/process/aggregated"
	"strings"
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

	aggregated.ProcessFarmerData(farmer)
}
