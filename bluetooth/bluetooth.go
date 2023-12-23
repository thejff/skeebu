package bluetooth

import (
	"fmt"
	"strconv"

	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

const (
	BATTERY_LEVEL_UUID = "00002A19-0000-1000-8000-00805F9B34FB"
	MANU_NAME_UUID     = "00002A29-0000-1000-8000-00805F9B34FB"
	MODEL_NUM_STR_UUID = "00002A24-0000-1000-8000-00805F9B34FB"
	DB_HASH_UUID       = "00002B2A-0000-1000-8000-00805F9B34FB"
)

type IKeeb interface {
	Init()
	GetBatteryLevels() Levels
	GetStats() Stats

	SetCapacity(capacity int)
	SetChargeRate(rate int)
	GetCapacity() int
	GetChargeRate() int
}

type Levels struct {
	Main       Level
	Peripheral []Level
}

type Level struct {
	Name  string
	Level int
}

type Stats struct {
	Name         string
	Alias        string
	Manufacturer string
	BatteryCount int
}

type keeb struct {
	dev      *device.Device1
	batts    batteries
	name     string
	manuName string
	alias    string
	dbHash   []byte
}

type batteries struct {
	main       battery
	peripheral []battery
	count      int
	stats      batteryStats
}

type batteryStats struct {
	capacity   int
	chargeRate int
}

type battery struct {
	name    string
	level   int
	offset  int
	service string
}

func ListDevices() (map[string]*device.Device1, error) {
	defaultAdap, err := adapter.GetDefaultAdapter()

	if err != nil {
		return nil, err
	}

	devices, err := defaultAdap.GetDevices()
	if err != nil {
		return nil, err
	}

	devMap := make(map[string]*device.Device1)
	for _, d := range devices {
		name := d.Properties.Alias
		devMap[name] = d
	}

	return devMap, nil
}

func SelectDevice(dev *device.Device1) IKeeb {
	var k IKeeb = &keeb{
		dev,
		batteries{},
		"",
		"",
		"",
		[]byte{},
	}

	return k
}

func (k *keeb) Init() {
	if err := k.getInfo(); err != nil {
		fmt.Printf("Error getting info\n%s\n", err.Error())
	}

	if err := k.getBatteries(); err != nil {
		fmt.Printf("Error getting battery data\n%s\n", err.Error())
	}

	// fmt.Printf("Data:\n%v\n", k)
}

func (k *keeb) HasUpdated() (bool, error) {

	return false, nil
}

func (k *keeb) SetAlias(alias string) error {

	return nil
}

func (k *keeb) SetBatteryAlias(service string, alias string) error {

	return nil
}

func (k *keeb) SetCapacity(c int) {
	k.batts.stats.capacity = c
}

func (k *keeb) SetChargeRate(r int) {
	k.batts.stats.chargeRate = r
}

func (k *keeb) GetCapacity() int {
	return k.batts.stats.capacity
}

func (k *keeb) GetChargeRate() int {
	return k.batts.stats.chargeRate
}

func (k *keeb) GetBatteryLevels() Levels {
	levels := Levels{}

	levels.Main = Level{
		Name:  k.batts.main.name,
		Level: k.batts.main.level,
	}

	for _, p := range k.batts.peripheral {
		level := Level{
			Name:  p.name,
			Level: p.level,
		}

		levels.Peripheral = append(levels.Peripheral, level)
	}

	return levels
}

func (k *keeb) GetStats() Stats {
	return Stats{
		Name:         k.name,
		Alias:        k.alias,
		Manufacturer: k.manuName,
		BatteryCount: k.batts.count,
	}
}

func (k *keeb) getInfo() error {

	if err := k.getManuName(); err != nil {
		return err
	}

	if err := k.getModelName(); err != nil {
		return err
	}

	k.getAlias()

	return nil
}

func (k *keeb) getManuName() error {
	manuChar, err := k.dev.GetCharByUUID(MANU_NAME_UUID)
	if err != nil {
		return err
	}

	bManuName, err := manuChar.GetValue()
	if err != nil {
		return err
	}

	k.manuName = string(bManuName)
	return nil
}

func (k *keeb) getModelName() error {
	modelChar, err := k.dev.GetCharByUUID(MODEL_NUM_STR_UUID)
	if err != nil {
		return err
	}

	bModelName, err := modelChar.GetValue()
	if err != nil {
		return err
	}

	k.name = string(bModelName)

	return nil
}

func (k *keeb) getAlias() {
	k.alias = k.dev.Properties.Alias
}

func (k *keeb) getBatteries() error {
	ops := make(map[string]interface{})

	battChars, err := k.dev.GetCharsByUUID(BATTERY_LEVEL_UUID)
	if err != nil {
		return err
	}

	batts := []battery{}

	for _, c := range battChars {

		batt := battery{}

		s := c.Properties.Service
		sStr := fmt.Sprintf("%v", s)
		batt.service = sStr

		sLen := len(sStr)
		offsetStr := sStr[sLen-2:]
		offset, err := strconv.ParseInt(offsetStr, 10, 0)
		if err != nil {
			return err
		}

		batt.offset = int(offset)

		bVal, err := c.ReadValue(ops)
		if err != nil {
			return err
		}

		var lvl int
		if bVal[0] == 255 {
			lvl = -1
		} else {
			lvl = int(bVal[0])
		}

		batt.level = lvl

		batts = append(batts, batt)
	}

	k.batts.count = len(batts)

	lowestOffset := 999
	lowestIndex := 999
	for i, b := range batts {
		if b.offset < lowestOffset {
			lowestOffset = b.offset
			lowestIndex = i
		}
	}

	if lowestIndex == 999 {
		lowestIndex = 0
	}

	main := batts[lowestIndex]
	main.name = "Main"

	k.batts.main = main

	// Remove main
	batts[lowestIndex] = batts[len(batts)-1]
	batts = batts[:len(batts)-1]

	// Add the peripherals
	for i, b := range batts {
		b.name = fmt.Sprintf("Peripheral %d", i+1)
		k.batts.peripheral = append(k.batts.peripheral, b)
	}

	return nil
}

func (k *keeb) ReadBatteryLevels() error {

	return nil
}

func (k *keeb) ReadBatteryLevel(b battery) error {

	return nil
}
