package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTableItem struct {
	Name string
	Age  int
}

func TestShowTableWithSlice_StructSlice(t *testing.T) {
	data := []testTableItem{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}
	// ShowTableWithSlice writes to os.Stdout via tablewriter; we mainly verify
	// it doesn't panic with valid input and covers the function.
	ctx := context.Background()
	assert.NotPanics(t, func() {
		ShowTableWithSlice(ctx, data)
	})
}

func TestShowTableWithSlice_PointerSlice(t *testing.T) {
	data := []*testTableItem{
		{Name: "Charlie", Age: 35},
		{Name: "Diana", Age: 28},
	}
	ctx := context.Background()
	assert.NotPanics(t, func() {
		ShowTableWithSlice(ctx, data)
	})
}

func TestShowTableWithSlice_InvalidType_Panics(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		ShowTableWithSlice(ctx, "not a slice")
	})
}

func TestShowTableWithSlice_NonStructSlice_Panics(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		ShowTableWithSlice(ctx, []int{1, 2, 3})
	})
}

func TestShowTableWithSlice_EmptySlice(t *testing.T) {
	ctx := context.Background()
	assert.NotPanics(t, func() {
		ShowTableWithSlice(ctx, []testTableItem{})
	})
}

func TestShowTable_ValidData(t *testing.T) {
	ctx := context.Background()
	data := [][]string{
		{"Name", "Age"},
		{"Alice", "30"},
		{"Bob", "25"},
	}
	assert.NotPanics(t, func() {
		ShowTable(ctx, data)
	})
}

func TestShowTable_EmptyData(t *testing.T) {
	ctx := context.Background()
	// ShowTable sorts data[1:], so empty input would panic on index out of range
	// with nil or empty slice. We test with just a header row.
	assert.NotPanics(t, func() {
		ShowTable(ctx, [][]string{{"Header"}})
	})
}
