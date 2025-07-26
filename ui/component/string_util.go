package component

import "strings"

//go:generate stringer -type=Direction
type Direction uint8

const (
	Direction_V Direction = iota
	Direction_H
)

func slidingWindowDisplayString(str string, width, frame int) string {
	if len(str) <= width {
		return str + strings.Repeat(" ", width-len(str))
	}
	offset := frame % (len(str) - width + 1)
	return str[offset : width+offset]
}
