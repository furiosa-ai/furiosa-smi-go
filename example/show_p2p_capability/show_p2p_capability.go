package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	devices, err := smi.ListDevices()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	header := table.Row{"#"}
	for _, device := range devices {
		info, err := device.DeviceInfo()
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
		header = append(header, filepath.Base(info.Name()))
	}
	t.AppendHeader(header)

	for _, device1 := range devices {
		info1, err := device1.DeviceInfo()
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}

		row := table.Row{filepath.Base(info1.Name())}
		for _, device2 := range devices {
			p2pAccessible, err := device1.GetP2PAccessible(device2)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				os.Exit(1)
			}

			row = append(row, p2pAccessibleToString(p2pAccessible))
		}
		t.AppendRow(row)
	}

	t.Render()
}

func p2pAccessibleToString(p2pAccessible bool) string {
	if p2pAccessible {
		return "Accessible"
	}
	return "Not Accessible"
}
