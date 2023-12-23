package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/thejff/skeebu/bluetooth"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	BATTERY_LEVEL_UUID = "00002A19-0000-1000-8000-00805F9B34FB"
	BATTERY_TEST_UUID  = "0000180F-0000-1000-8000-00805F9B34FB"
)

// Look for bl devices, list connected
// Select one to try and read
// Read central and peripheral battery usage
// Output usage

// 1: Read both batteries

/*
TUI:
Show BL devices
Select Device
Show %
Allow user to input battery capacity in mAh
Allow user to input charge rate (e.g nice!nano is 100mA)
Estimate mAh > 65% = Used: 280mAh, Remaining: 520mAh
Estimate time to charge = Used/Charge Rate, e.g: 280/100 = 2.8 hours

Show for one or more batteries

Split into 3 views
1. Top Left: Select device, saved or new
2. Bottom left: Settings: Set capacity, Set charge rate, save
3. Right: Stats, progress bar showing capacity, names etc.
*/

func test(v *device.Device1) {

	battChars, err := v.GetCharsByUUID(BATTERY_LEVEL_UUID)
	if err != nil {
		log.Println("Error getting battery level characteristic.")
		log.Print(err)
		fmt.Print("\n")
		return
	}

	// ops := make(map[string]interface{})
	// bVals := []int{}

	for _, c := range battChars {

		s := c.Properties.Service
		// fmt.Printf("s: %v\n", s)

		ops := make(map[string]interface{})

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

		// bVals = append(bVals, lvl)

		sLen := len(s)

		text := ""
		if s[sLen-2:] == "10" {
			// fmt.Sprintf("Main")
			text += "Main"
		} else {
			text += "Peripheral"
			// fmt.Sprintf("Secondary")
		}

		text = fmt.Sprintf("%s level is %d%%", text, lvl)
		fmt.Printf("------\nService: %s\n%s\n------\n", s, text)

	}

	/* for i, level := range bVals {
		log.Print("Device: ")
		if i == 0 {
			log.Println("Main")
		} else {
			log.Printf("Peripheral %d\n", i)
		}
		log.Printf("Battery Level is: %d%%\n", level)
	} */

	fmt.Printf("\n")
}

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

		k := bluetooth.SelectDevice(v)
		k.Init()

		k.SetCapacity(820)
		k.SetChargeRate(100)

		stats := k.GetStats()
		fmt.Printf(
			"Device: %s by %s\nAlias: %s\nNumber of batteries: %d\n\n",
			stats.Name,
			stats.Manufacturer,
			stats.Alias,
			stats.BatteryCount,
		)

		batLevels := k.GetBatteryLevels()

		chargeStats, err := calculateStats(k)
		if err != nil {
			fmt.Print(err)
			return
		}

		mainDischarge := estimateDischargeTime(chargeStats.Main.Capacity-chargeStats.Main.Used, false)

		fmt.Printf(
			"%s battery: %d%%\nUsed: %dmAh of %dmAh\nIt will take approximately %dh and %dm to charge.\n",
			batLevels.Main.Name,
			batLevels.Main.Level,
			chargeStats.Main.Capacity-chargeStats.Main.Used,
			chargeStats.Main.Capacity,
			chargeStats.Main.Hours,
			chargeStats.Main.Minutes,
		)

		mainPrettyDis := prettifyDischarge(mainDischarge)
		fmt.Printf("You will probably need to charge it in about %s.\n\n", mainPrettyDis)

		for _, p := range batLevels.Peripheral {

			cStats := chargeStats.Peripherals[p.Name]
			pDischarge := estimateDischargeTime(cStats.Capacity-cStats.Used, true)

			fmt.Printf(
				"%s battery: %d%%\nUsed: %dmAh of %dmAh\nIt will take approximately %dh and %dm to charge.\n",
				p.Name,
				p.Level,
				cStats.Capacity-cStats.Used,
				cStats.Capacity,
				cStats.Hours,
				cStats.Minutes,
			)

			prettyDis := prettifyDischarge(pDischarge)
			fmt.Printf("You will probably need to charge it in about %s.\n\n", prettyDis)
		}

		return

		for {
			test(v)

			fmt.Print("\n\nSleeping for 5...\n\n")
			time.Sleep(5 * time.Second)
		}

		// log.Printf("Device: %s\n", v.Properties.Alias)
		/* uuids := v.Properties.UUIDs

		log.Println("Services:")
		for _, uuid := range uuids {


			log.Printf("Service: %s\n", uuid)
		} */

		/* chars, err := v.GetCharacteristics()
		if err != nil {
			log.Println("Error getting characteristics.")
			log.Print(err)
			fmt.Print("\n")
		} */

		/*
			for _, c := range chars {
				fmt.Printf("UUID: %s\n", c.Properties.UUID)
			} */

		/* services, err := v.GetAllServicesAndUUID()
		if err != nil {
			log.Println("Error getting service data.")
			log.Print(err)
			fmt.Print("\n")
		}

		for _, service := range services {
			uuidSplit := strings.Split(service, ":")
			if uuidSplit[0] == BATTERY_LEVEL_UUID {
				offset := uuidSplit[1][len(uuidSplit[1])-2:]
				/* fmt.Printf("UUID: %s\n", uuidSplit[0])
				fmt.Printf("Offset: %s\n\n", uuidSplit[1][len(uuidSplit[1])-2:]) */

		/*		bLvlChar, err := v.GetCharByUUID(uuidSplit[0])
					if err != nil {
						log.Println("Error getting service characteristic.")
						log.Print(err)
						fmt.Print("\n")
					}

					fmt.Printf("bLvlChar.Properties.Service: %v\n", bLvlChar.Properties.Service)

					bLvl, err := bLvlChar.ReadValue(ops)
					if err != nil {
						log.Println("Error getting battery level characteristic.")
						log.Print(err)
						fmt.Print("\n")
					}

					fmt.Printf("Device ")
					offsetInt, err := strconv.ParseInt(offset, 0, 0)
					if err != nil {
						log.Println("Error getting battery level characteristic.")
						log.Print(err)
						fmt.Print("\n")
					}

					if offsetInt > 10 {
						fmt.Printf("Peripheral")
					} else {
						fmt.Printf("Main")
					}

					fmt.Printf(": %d\n", bLvl[0])

				}
			} */

	}

	return

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Wuh oh, there's been an error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	spinner  spinner.Model
}

func initialModel() model {
	spin := spinner.New()
	spin.Spinner = spinner.Jump
	spin.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("#7D56F4"))

	return model{
		choices:  []string{"Dev 1", "Dev 2"},
		selected: make(map[int]struct{}),
		spinner:  spin,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	}

	return m, nil
}

func (m model) View() string {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(2).
		PaddingLeft(4).
		PaddingBottom(1).
		Width(40)

	s := "Which device are we checking the batteries of?\n\n"

	for i, choice := range m.choices {
		cursor := " "

		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit."
	s += fmt.Sprintf(" %s\n", m.spinner.View())
	// str := fmt.Sprintf("\n%s Press q to quit.\n", m.spinner.View())

	// return s
	return style.Render(s)
}
