package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/thejff/skeebu/bluetooth"
)

type ChargeStats struct {
	Main        ChargeStat
	Peripherals map[string]ChargeStat
}

type ChargeStat struct {
	Hours    int
	Minutes  int
	Capacity int
	Used     int
}

const (
	C_SELF_DISCHARGE  = 0.216
	C_BOARD_QUIESCENT = 0.068
	C_ZMK_USAGE       = 2.3
	C_ZMK_TYPING      = 0.029
	C_BUFFER          = 0.7

	P_SELF_DISCHARGE  = 0.216
	P_BOARD_QUIESCENT = 0.068
	P_ZMK_USAGE       = 0.076
	P_ZMK_TYPING      = 0.028
	P_BUFFER          = 0.1

	// Average voltage, 3.7v battery full at 4.2v
	VOLTAGE = 4
)

func prettifyDischarge(t time.Duration) string {
	if t.Hours() < 24 {
		return fmt.Sprintf("%.f hours", t.Hours())
	}

	days := t.Hours() / 24

	if days < 30 {
		return fmt.Sprintf("%.f days", days)
	}

	months := days / 30
	daysRemain := int(days) % 30

	return fmt.Sprintf("%.f months and %d days", months, daysRemain)

}

func estimateDischargeTime(capacity int, isPeripheral bool) time.Duration {

	usage :=
		C_SELF_DISCHARGE +
			C_BOARD_QUIESCENT +
			C_ZMK_USAGE +
			C_ZMK_TYPING +
			C_BUFFER

	if isPeripheral {
		usage =
			P_SELF_DISCHARGE +
				P_BOARD_QUIESCENT +
				P_ZMK_USAGE +
				P_ZMK_TYPING +
				P_BUFFER
	}

	watts := usage / 1000.0

	amps := watts / VOLTAGE

	milliamps := amps * 1000.0

	remaining := float64(capacity) / milliamps

	hours := time.Duration(remaining) * time.Hour

	return hours
}

func calculateStats(k bluetooth.IKeeb) (ChargeStats, error) {

	usedFn := func(c float32, l float32) float32 {
		return (c - (l/100.0)*c)
	}

	stats := ChargeStats{}
	stats.Peripherals = make(map[string]ChargeStat)

	batLevels := k.GetBatteryLevels()

	capacity := float32(k.GetCapacity())
	chargeRate := float32(k.GetChargeRate())

	mainLevel := float32(batLevels.Main.Level)
	mainUsed := usedFn(capacity, mainLevel)

	stats.Main.Capacity = int(capacity)
	stats.Main.Used = int(mainUsed)

	hours, mins, err := hoursAndMins(mainUsed, chargeRate)
	if err != nil {
		return stats, err
	}

	stats.Main.Hours = hours
	stats.Main.Minutes = mins

	for _, p := range batLevels.Peripheral {
		used := usedFn(capacity, float32(p.Level))
		stat := ChargeStat{
			Capacity: int(capacity),
			Used:     int(used),
		}

		hours, mins, err := hoursAndMins(used, chargeRate)
		if err != nil {
			return stats, err
		}

		stat.Hours = hours
		stat.Minutes = mins

		stats.Peripherals[p.Name] = stat
	}

	return stats, nil
}

func hoursAndMins(used float32, chargeRate float32) (int, int, error) {

	fChargeTime := used / chargeRate
	sChargeTime := fmt.Sprintf("%.2f", fChargeTime)
	timeSplit := strings.Split(sChargeTime, ".")

	hours := timeSplit[0]
	iHours, err := strconv.Atoi(hours)
	if err != nil {
		return 0, 0, err
	}

	fMin, err := strconv.ParseFloat(
		fmt.Sprintf("0.%s", timeSplit[1]),
		32,
	)
	if err != nil {
		return 0, 0, err
	}

	fMin = 60 * fMin

	iMins := 1
	if fMin > 1.0 {
		iMins = int(fMin)
	}

	if fMin <= 0.0 {
		iMins = 0
	}

	return iHours, iMins, nil
}

func (c *ChargeStat) Prettify() string {

	return ""
}
