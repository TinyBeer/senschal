package util

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sort"

	"seneschal/internal/ilog"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
)

var Err_InvalidDataType = errors.New("invalid data type")

func ShowTableWithSlice(ctx context.Context, data any) {
	dt := reflect.TypeOf(data)
	dv := reflect.ValueOf(data)
	if dt.Kind() != reflect.Slice {
		panic(Err_InvalidDataType)
	}
	de := dt.Elem()

	if de.Kind() == reflect.Ptr {
		de = de.Elem()
	}
	if de.Kind() != reflect.Struct {
		panic(Err_InvalidDataType)
	}
	var tdata [][]string
	head := make([]string, 0, de.NumField())
	for i := range de.NumField() {
		head = append(head, de.Field(i).Name)
	}
	tdata = append(tdata, head)

	for i := range dv.Len() {
		d := make([]string, 0, de.NumField())

		v := dv.Index(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for j := range de.NumField() {
			d = append(d, fmt.Sprint(v.Field(j)))
		}
		tdata = append(tdata, d)
	}
	ShowTable(ctx, tdata)
}

func ShowTable(ctx context.Context, data [][]string) {
	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithRenderer(renderer.NewMarkdown()),
	)
	sort.Slice(data[1:], func(i, j int) bool {
		return data[i+1][0] < data[j+1][0]
	})
	table.Header(data[0])

	var err error
	err = table.Bulk(data[1:])
	if err != nil {
		slog.ErrorContext(
			ctx, "bulk table failed",
			ilog.ErrorAttr(err),
		)
	}
	err = table.Render()
	if err != nil {
		slog.ErrorContext(
			ctx, "render table failed",
			ilog.ErrorAttr(err),
		)
	}
}
