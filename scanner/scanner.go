/*
Package scanner implements a buffered scanner which provides lexical analysis
(tokenizing) of ARC source code. A scanner takes a bufio.Reader as source which
can then be tokenized through repeated calls to the Scan method.
*/
package scanner

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lukasmalkmus/arc/token"
)

var eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r              *bufio.Reader
	pos            token.Pos
	resetCharCount bool
}

// New returns a new instance of Scanner.
func New(r io.Reader) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(r),
		pos: token.Pos{Filename: "", Line: 1, Char: 0},
	}
}

// NewFileScanner returns a new instance of Scanner, but will exclusively take
// an *os.File as argument instead of the more general io.Reader interface.
// Therefore it will enhance token positions with the filename.
func NewFileScanner(f *os.File) *Scanner {
	return &Scanner{
		r:   bufio.NewReader(f),
		pos: token.Pos{Filename: f.Name(), Line: 1, Char: 0},
	}
}

// Scan returns the read token and literal value.
func (s *Scanner) Scan() (token.Token, string, token.Pos) {
	// Read the read rune.
	ch, pos := s.read()

	// If we see a whitespace then consume all contiguous whitespace.
	// If we see a newline then consume all contiguous newline.
	// If we see an exclamation mark then consume as a comment.
	// If we see a dot then consume as a directive.
	// If we see a digit consume as an integer.
	// If we see a letter or % then consume as an ident or reserved word.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isNewline(ch) {
		s.unread()
		return s.scanNewline()
	} else if ch == '!' {
		s.unread()
		return s.scanComment()
	} else if ch == '.' {
		s.unread()
		return s.scanDirective()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if isNumber(ch) {
		s.unread()
		return s.scanInteger()
	} else if ch == '%' {
		s.unread()
		return s.scanRegister()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		pos.Char--
		return token.EOF, "", pos
	case '+':
		return token.PLUS, string(ch), pos
	case '-':
		return token.MINUS, string(ch), pos
	case '[':
		return token.LBRACKET, string(ch), pos
	case ']':
		return token.RBRACKET, string(ch), pos
	case ',':
		return token.COMMA, string(ch), pos
	case ':':
		return token.COLON, string(ch), pos
	}

	// No match results in an illegal token.
	return token.ILLEGAL, string(ch), pos
}

// scanComment consumes the current rune and all contiguous comment runes.
func (s *Scanner) scanComment() (token.Token, string, token.Pos) {
	// Create a buffer and drop first character.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent character into the buffer.
	// Newline or EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof || isNewline(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Return comment with text as literal value.
	return token.COMMENT, buf.String(), pos
}

// scanDirective consumes the current rune and all contiguous directive runes.
func (s *Scanner) scanDirective() (token.Token, string, token.Pos) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent directive character into the buffer.
	// Non-directive characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if !isLetter(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Check if the identifier is a directive.
	if tok := token.Lookup(buf.String()); tok.IsDirective() {
		return tok, buf.String(), pos
	}

	// Otherwise return an ILLEGAL token (because it can't be an identifier
	// starting with a '.').
	return token.ILLEGAL, buf.String(), pos
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (token.Token, string, token.Pos) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isNumber(ch) && ch != '_' {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Make sure the last character is not an underscore, which is illegal.
	if ch := buf.Bytes()[buf.Len()-1]; ch == '_' {
		return token.ILLEGAL, buf.String(), pos
	}

	// Check if the identifier is a keyword.
	if tok := token.Lookup(buf.String()); tok.IsKeyword() {
		return tok, buf.String(), pos
	}

	// Otherwise return as a regular identifier.
	return token.IDENT, buf.String(), pos
}

// scanInteger consumes the current rune and all contiguous integer runes.
func (s *Scanner) scanInteger() (token.Token, string, token.Pos) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)
	sawX := false

	// Read every subsequent integer character into the buffer.
	// Non-integer characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if (ch == 'x' || ch == 'X') && sawX {
			s.unread()
			break
		} else if !isNumber(ch) && (ch != 'x' && ch != 'X') {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Check if literal can be parsed to valid integer.
	if _, err := strconv.ParseInt(buf.String(), 0, 64); err != nil {
		return token.ILLEGAL, buf.String(), pos
	}
	val := strings.Replace(buf.String(), "X", "x", -1)

	// Return as an integer.
	return token.INT, val, pos
}

// scanNewline consumes the current rune and all contiguous newline.
func (s *Scanner) scanNewline() (token.Token, string, token.Pos) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if !isNewline(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Strip Carriage-Return.
	clean := stripCR(buf.Bytes())
	buf.Reset()
	buf.Write(clean)

	// Increase position. The position of the token is decremented because the
	// newline is found in the previous line. Furthermore, resetCharCount flag
	// gets set.
	s.pos.Line += buf.Len()
	pos.Line += buf.Len() - 1
	s.resetCharCount = true

	return token.NL, buf.String(), pos
}

// scanRegister consumes the current rune and all contiguous register ident
// runes.
func (s *Scanner) scanRegister() (token.Token, string, token.Pos) {
	// Create a buffer and drop first character.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isNumber(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// No identifier after % char is not a valid register.
	if buf.Len() < 2 {
		return token.ILLEGAL, buf.String(), pos
	}

	// First identifier char must be a 'r'.
	if ch := buf.Bytes()[1]; ch != 'r' {
		return token.ILLEGAL, buf.String(), pos
	}

	return token.REG, buf.String(), pos
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (token.Token, string, token.Pos) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.read()
	buf.WriteRune(ch)

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch, _ := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return token.WS, buf.String(), pos
}

// read reads the next rune from the bufferred reader. Returns the rune(0) if an
// error occurs (or io.EOF is returned).
func (s *Scanner) read() (rune, token.Pos) {
	// Reset character count.
	if s.resetCharCount {
		s.pos.Char = 0
		s.resetCharCount = false
	}
	s.pos.Char++

	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof, s.pos
	}
	return ch, s.pos
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	s.r.UnreadRune()
	s.pos.Char--
}

// isWhitespace returns true if the rune is a space or tab.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' }

// isNewline returns true if the rune is a newline.
func isNewline(ch rune) bool { return ch == '\n' || ch == '\r' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isNumber returns true if the rune is a digit.
func isNumber(ch rune) bool { return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') }

// stripCR removes every carriage-return from a slice of bytes, effectively
// turning a CRLF into a LF.
func stripCR(b []byte) []byte {
	c := make([]byte, 0)
	for _, ch := range b {
		if ch == '\n' {
			c = append(c, ch)
		}
	}
	return c
}
