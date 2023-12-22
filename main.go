package main

import (
	"fmt"
	"log"

	"github.com/muka/go-bluetooth/bluez/profile/adapter"
)

// Look for bl devices, list connected
// Select one to try and read
// Read central and peripheral battery usage
// Output usage

// 1: Read both batteries

const (
	BATTERY_LEVEL_UUID   = "00002A19-0000-1000-8000-00805F9B34FB"
	BATTERY_SERVICE_UUID = "0000180F-0000-1000-8000-00805F9B34FB"
)

func main() {
	defaultId := adapter.GetDefaultAdapterID()
	a, err := adapter.GetAdapter(defaultId)
	if err != nil {
		log.Println("Error getting adapter")
		log.Panic(err)
		return
	}

	devices, err := a.GetDevices()
	if err != nil {
		log.Println("Error getting devices")
		log.Panic(err)
		return
	}

	for _, v := range devices {
		// log.Printf("Device: %s\n", v.Properties.Alias)
		/* uuids := v.Properties.UUIDs

		log.Println("Services:")
		for _, uuid := range uuids {


			log.Printf("Service: %s\n", uuid)
		} */

		battChars, err := v.GetCharsByUUID(BATTERY_LEVEL_UUID)
		if err != nil {
			log.Println("Error getting battery level characteristic.")
			log.Print(err)
			fmt.Print("\n")
			continue
		}

		ops := make(map[string]interface{})
		bVals := []int{}

		for _, c := range battChars {

			bVal, err := c.ReadValue(ops)
			if err != nil {
				log.Println("Error reading battery level characteristic")
				log.Print(err)
				fmt.Print("\n")
				continue
			}

			var lvl int

			if bVal[0] == 255 {
				lvl = -1
			} else {
				lvl = int(bVal[0])
			}

			bVals = append(bVals, lvl)
		}

		for _, level := range bVals {
			log.Printf("Battery Level is: %d%%\n", level)
		}

		fmt.Printf("\n")

	}
}
