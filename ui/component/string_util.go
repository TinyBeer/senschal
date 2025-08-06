package component

import (
	"strings"
	"unicode"
)

//go:generate stringer -type=Direction
type Direction uint8

const (
	Direction_V Direction = iota
	Direction_H
)

func slidingWindowDisplayString(str string, width, frame int) string {
	if !containWideCharacter(str) {
		if l := len(str); l <= width {
			return str + strings.Repeat(" ", width-l)
		}
		offset := frame % (len(str) - width + 1)
		return str[offset : width+offset]
	}

	candidate := getCandidate(str, width)

	idx := frame % len(candidate)
	return candidate[idx]
}

func getDisplayLength(str string) int {
	length := 0
	for _, r := range []rune(str) {
		if isWideCharacter(r) {
			length += 2
		} else {
			length++
		}
	}
	return length
}

func getCandidate(str string, width int) []string {
	rs := []rune(str)
	var candidate []string
	for i := range rs {
		ld := 0
		j := i
		for ; j < len(rs); j++ {
			add := 1
			if isWideCharacter(rs[j]) {
				add = 2
			}
			if ld+add > width {
				break
			}

			ld += add
		}
		str := string(rs[i:j])
		candidate = append(candidate, str+strings.Repeat(" ", width-ld))
		if j == len(rs) {
			break
		}
	}
	return candidate
}

// containWideCharacter 判断字符串是否包含占用2个字符显示的字符（全角字符）
func containWideCharacter(str string) bool {
	for _, r := range []rune(str) {
		if isWideCharacter(r) {
			return true
		}
	}
	return false
}

// isWideCharacter 判断单个字符是否为全角字符
func isWideCharacter(r rune) bool {
	// 全角ASCII、全角标点符号的范围
	if r >= 0xFF01 && r <= 0xFFEF {
		return true
	}
	// 中文、日文、韩文的Unicode范围
	if (r >= 0x4E00 && r <= 0x9FFF) || // 中文
		(r >= 0x3040 && r <= 0x309F) || // 日文平假名
		(r >= 0x30A0 && r <= 0x30FF) || // 日文片假名
		(r >= 0xAC00 && r <= 0xD7AF) { // 韩文
		return true
	}
	// 其他可能的宽字符（如Emoji）
	if unicode.Is(unicode.Han, r) || // 汉字
		unicode.Is(unicode.Hangul, r) || // 韩文
		unicode.Is(unicode.Hiragana, r) || // 平假名
		unicode.Is(unicode.Katakana, r) { // 片假名
		return true
	}
	return false
}
