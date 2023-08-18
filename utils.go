
package mcla

import (
	"strings"
)

func rsplit(line string, b byte)(left, right string){
	i := strings.LastIndexByte(line, b)
	if i < 0 {
		return "", line
	}
	return line[:i], line[i + 1:]
}

func split(line string, b byte)(left, right string){
	i := strings.IndexByte(line, b)
	if i < 0 {
		return line, ""
	}
	return line[:i], line[i + 1:]
}

func lcsSplit[T comparable](a, b []T)(n int, a1, a2, b1, b2 []T){
	if len(a) == 0 || len(b) == 0 {
		return 0, a, nil, b, nil
	}
	type ele struct {
		a, b int
		len int
	}
	ch := make([]ele, len(b) + 1)
	for i, p := range a {
		var last ele
		for j, q := range b {
			cur := ch[j+1]
			if p == q {
				if last.len == 0 {
					last = ele{i, j, 1}
				}else if last.a + last.len == i && last.b + last.len == j {
					last.len++
				}
				ch[j + 1] = last
			}else if prev := ch[j]; prev.len > cur.len {
				ch[j + 1] = prev
			}
			last = cur
		}
	}
	res := ch[len(b)]
	if res.len == 0 {
		return
	}
	return res.len, a[:res.a], a[res.a + res.len:], b[:res.b], b[res.b + res.len:]
}

func lcsLength[T comparable](a, b []T)(n int){
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n, a1, a2, b1, b2 := lcsSplit(a, b)
	n += lcsLength(a1, b1) + lcsLength(a2, b2)
	return
}

func lcsPercent[T comparable](a, b []T)(v float32){
	if len(b) > len(a) {
		a, b = b, a
	}
	if len(b) == 0 {
		return 0.0
	}
	n := (float32)(lcsLength(a, b))
	return n / (float32)(len(a))
}

func lineMatchPercent(a, b string)(float32){
	return lcsPercent(([]rune)(a), ([]rune)(b))
}
