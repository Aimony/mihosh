package utils

import "strings"

// DisplayWidth 计算字符串的显示宽度（中文占2个单位，英文占1个单位）
func DisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r > 127 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

// PadString 将字符串填充到指定显示宽度
func PadString(s string, targetWidth int) string {
	currentWidth := DisplayWidth(s)
	if currentWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-currentWidth)
}

// TruncateString 根据显示宽度截断字符串（支持中文）
func TruncateString(s string, maxWidth int) string {
	if DisplayWidth(s) <= maxWidth {
		return s
	}
	// 逐字符截断直到符合宽度
	result := ""
	currentWidth := 0
	for _, r := range s {
		var rw int
		if r > 127 {
			rw = 2
		} else {
			rw = 1
		}
		if currentWidth+rw > maxWidth-2 {
			break
		}
		result += string(r)
		currentWidth += rw
	}
	return result + ".."
}
