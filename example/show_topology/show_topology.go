// Command show_topology prints two device-by-device tables: the link type
// between every pair of NPU devices, and whether each pair is P2P
// accessible. It is a port of the Rust crate's examples/show_topology.rs.
// Run it on a host with Furiosa NPUs:
//
//	go run ./example/show_topology
package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

func main() {
	devices, err := api.ListDevices()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	header := []string{"Device"}
	for i := range devices {
		header = append(header, devices[i].DeviceName())
	}

	rows := make([][]string, 0, len(devices))
	for i := range devices {
		row := []string{devices[i].DeviceName()}
		for j := range devices {
			linkType, err := devices[i].DeviceToDeviceLinkType(&devices[j])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			row = append(row, linkType.String())
		}

		rows = append(rows, row)
	}

	renderTable(header, rows)

	fmt.Println()

	header = []string{"P2P"}
	for i := range devices {
		header = append(header, devices[i].DeviceName())
	}

	rows = make([][]string, 0, len(devices))
	for i := range devices {
		row := []string{devices[i].DeviceName()}
		for j := range devices {
			row = append(row, fmt.Sprintf("%t", devices[i].P2PAccessible(&devices[j])))
		}

		rows = append(rows, row)
	}

	renderTable(header, rows)
}

func renderTable(header []string, rows [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.SeparateRows = true
	t.Style().Format.Header = text.FormatDefault
	t.AppendHeader(toRow(header))
	for _, row := range rows {
		t.AppendRow(toRow(row))
	}

	t.Render()
}

func toRow(cells []string) table.Row {
	row := make(table.Row, 0, len(cells))
	for _, cell := range cells {
		row = append(row, cell)
	}

	return row
}
