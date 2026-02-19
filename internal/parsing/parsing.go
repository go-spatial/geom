package parsing

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

const (
	EOF byte = 255
)

// SplitFunc is used to tokenized the input. The arguments are the initial substring of the remaining unprocessed data and
// a flag, atEOF, that reports whether the Reader has no more data to give. The return values
// are the number of bytes to advance the input, the token, the count of the token, text body for the token (should be a new copy),
// and an error if there is one.
// Internally this makes use of [bufio#SplitFunc](https://pkg.go.dev/bufio#SplitFunc) function.
type SplitFunc func(data []byte, atEOF bool) (advance int, token byte, count uint64, body []byte, err error)

func Err(err error) (int, byte, uint64, []byte, error) { return 0, 0, 0, nil, err }
func Eof() (int, byte, uint64, []byte, error)          { return 0, EOF, 0, nil, io.EOF }
func MoreData() (int, byte, uint64, []byte, error)     { return 0, 0, 0, nil, nil }
func SplitWrap(fn SplitFunc) func([]byte, bool) (advance int, body []byte, err error) {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		advance, symb, count, bdy, err := fn(data, atEOF)
		if advance == 0 && symb == 0 && count == 0 && bdy == nil && err == nil {
			return 0, nil, nil
		}
		return advance, EncodeSymbol(symb, count, bdy), err
	}
}

func IsNewLine(r rune) bool {
	return r == '\n' || r == '\r'
}

func IsSpace(r rune) bool {
	if IsNewLine(r) {
		return false
	}
	return unicode.IsSpace(r)
}

func EncodeSymbolForFunc(fn func() (int, byte, uint64, []byte, error)) (int, []byte, error) {
	a, s, c, b, err := fn()
	if err != nil {
		return 0, nil, err
	}
	return a, EncodeSymbol(s, c, b), err
}

func EncodeSymbol(sym byte, count uint64, content []byte) (symbol []byte) {
	if sym == 0 && count == 0 && content == nil {
		// More data code
		return nil
	}
	symbol = make([]byte, 1+8, 1+8+len(content))
	symbol[0] = sym
	binary.PutUvarint(symbol[1:], count)
	return append(symbol, content...)
}

func GetRune(data []byte, atEOF bool) (r rune, n int, err error) {
	fchar, n := utf8.DecodeRune(data)
	// Not enough characters to tokenize.
	if fchar == utf8.RuneError {
		if !atEOF {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("Unknown byte sequence.")
	}
	return fchar, n, nil
}

func GetPossibleDouble(r rune, n1 int, sym byte, data []byte, atEOF bool) (advance int, symbol []byte, err error) {
	if len(data) == 0 {
		if !atEOF {
			// Need more data.
			return 0, nil, nil
		}
		b := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(b, r)
		return n1, EncodeSymbol(sym, 1, b), nil

	}
	fchar, n, err := GetRune(data, atEOF)
	if err != nil {
		return 0, nil, err
	}
	if n == 0 {
		// Need more data.
		if !atEOF {
			return 0, nil, nil
		}
		b := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(b, r)
		return n1, EncodeSymbol(sym, 1, b), nil
	}

	if fchar != r {
		b := make([]byte, utf8.RuneLen(r))
		utf8.EncodeRune(b, r)
		return n1, EncodeSymbol(sym, 1, b), nil
	}
	b := make([]byte, utf8.RuneLen(r)*2)
	utf8.EncodeRune(b, r)
	utf8.EncodeRune(b[n:], r)
	return n + n1, EncodeSymbol(sym+1, 1, b), nil
}

func GetSeq(sym byte, fn func(r rune) bool, data []byte, isEOF bool) (advance int, symbol []byte, err error) {
	advance, sym, count, bdy, err := GetSeq1(sym, fn, data, isEOF)
	if advance == 0 && sym == 0 && count == 0 && bdy == nil && err == nil {
		return 0, nil, nil
	}
	return advance, EncodeSymbol(sym, count, bdy), err
}
func GetSeq1(sym byte, fn func(r rune) bool, data []byte, isEOF bool) (advance int, symbol byte, count uint64, body []byte, err error) {
	var contents []byte
	var fchar rune
	var n int
	var num uint64
	d := data

	for len(d) > 0 {
		fchar, n, err = GetRune(d, isEOF)
		if err != nil {
			return Err(err)
		}
		if !fn(fchar) {
			// Assume that the function calling this has made sure there is at least one of the
			// Items.
			return advance, sym, num, contents, nil
		}
		num++
		contents = data[:advance+1]
		advance += n
		d = data[advance:]
	}
	if !isEOF {
		// We need more data. Because it is possible that there are more items.
		return MoreData()
	}
	return advance, sym, num, contents, nil
}

type Position struct {
	Row uint
	Col uint
}

func (pos Position) String() string { return fmt.Sprintf("(%v,%v)", pos.Row, pos.Col) }
func (pos Position) Add(row, col uint) Position {
	pos.Row, pos.Col = pos.Row+row, pos.Col+col
	return pos
}

type Scanner struct {
	scanner *bufio.Scanner
	// symbol is the encoded symbol and contents for that symbol as well as a count
	// symbol[0]	—— symbol token value
	// symbol[1:9]	—— count uint64
	// symbol[9:]	—— text of symbol
	symbol      []byte
	nextBytes   []byte
	hasNext     bool
	previousPos Position
	currentPos  Position
	atEnd       bool
}

func NewScanner(r io.Reader, splitFn func([]byte, bool) (int, []byte, error)) *Scanner {
	s := new(Scanner)
	fn := func(b []byte, atEnd bool) (int, []byte, error) {
		adv, body, err := splitFn(b, atEnd)
		pos := calculatePosition(b, adv, s.currentPos)
		if err != nil || adv <= 0 {
			if err != nil {
				err = ErrAt{Err: err, Pos: pos}
			}
			return adv, body, err
		}
		s.currentPos = pos
		return adv, body, nil
	}
	s.scanner = bufio.NewScanner(r)
	s.scanner.Split(fn)
	return s
}

func calculatePosition(b []byte, adv int, current Position) Position {
	if adv == 0 {
		return current
	}
	rows := uint(bytes.Count(b[:adv], []byte("\n")))
	if rows == 0 {
		return current.Add(0, uint(adv))
	}
	if idx := bytes.LastIndexByte(b[:adv], '\n'); idx != -1 {
		return current.Add(rows, uint(len(b[:adv])-idx))
	}
	return current.Add(rows, 0)
}

func (s *Scanner) RawPeek() (sym []byte, more bool) {
	if s.atEnd {
		return []byte{EOF}, false
	}
	if s.hasNext {
		return s.nextBytes, true
	}
	if !s.scanner.Scan() {
		s.atEnd = true
		return []byte{EOF}, false
	}
	s.hasNext = true
	s.previousPos = s.currentPos
	s.nextBytes = s.scanner.Bytes()
	return s.nextBytes, true
}

func (s *Scanner) Position() Position {
	if s == nil {
		return Position{}
	}
	if s.hasNext {
		return s.previousPos
	}
	return s.currentPos
}

func (s *Scanner) NextBytes() []byte {
	b, m := s.RawPeek()
	if !m {
		return nil
	}
	return b[9:]
}
func (s *Scanner) NextText() string {
	b := s.NextBytes()
	if b == nil {
		return ""
	}
	return string(b)
}

func (s *Scanner) NextMore() bool {
	_, more := s.RawPeek()
	return more
}
func (s *Scanner) NextSymbol() byte {
	b, m := s.RawPeek()
	if !m {
		return EOF
	}
	return b[0]
}

func (s *Scanner) RawBytes() []byte { return s.symbol }
func (s *Scanner) Symbol() byte {
	return s.symbol[0]

}
func (s *Scanner) Bytes() []byte {
	if len(s.symbol) <= 9 {
		return nil
	}
	return s.symbol[9:]
}
func (s *Scanner) Text() string {
	b := s.Bytes()
	if b == nil {
		return ""
	}
	return string(b)
}
func (s *Scanner) Err() error  { return s.scanner.Err() }
func (s *Scanner) AtEnd() bool { return s.atEnd }

func (s *Scanner) Scan() bool {
	if s.atEnd {
		return false
	}
	if s.hasNext {
		s.symbol = s.nextBytes
		s.hasNext = false
		return true
	}
	if !s.scanner.Scan() {
		s.atEnd = true
		return false
	}
	s.symbol = s.scanner.Bytes()
	return true
}
