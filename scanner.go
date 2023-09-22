
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
	return &lineScanner{
		count: 0,
		Scanner: bufio.NewScanner(r),
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
