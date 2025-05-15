package tool

import (
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
)

func ShowTable(data [][]string) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	sort.Slice(data[1:], func(i, j int) bool {
		return data[i+1][0] < data[j+1][0]
	})
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()
}
