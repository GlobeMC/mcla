
package mcla

import (
	"bufio"
	"io"
)

type lineScanner struct {
	count int
	*bufio.Scanner
}

func newLineScanner(r io.Reader)(*lineScanner){
	bs := bufio.NewScanner(r)
	bs.Buffer(make([]byte, 256 * 1024), 1024 * 1024) // 256KB per line, large enough?
	return &lineScanner{
		count: 0,
		Scanner: bs,
	}
}

func (s *lineScanner)Scan()(bool){
	if !s.Scanner.Scan() {
		return false
	}
	s.count++
	return true
}

func (s *lineScanner)Count()(int){
	return s.count
}
