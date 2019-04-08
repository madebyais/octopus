package util

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

// GenerateTable will print log in table format
func GenerateTable(headers []string, rows [][]string) {
	fmt.Println(``)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetRowLine(true)

	for _, v := range rows {
		table.Append(v)
	}
	table.Render()
}
