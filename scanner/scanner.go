// Package scanner implements a scanner for ARC source code. It takes a
// bufio.Reader as source which can then be tokenized through repeated calls to
// the Scan method.
package scanner

import (
	"bufio"
	"bytes"
	"io"

	"github.com/LukasMa/arc/token"
)

var eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// New returns a new instance of Scanner.
func New(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the read token and literal value.
func (s *Scanner) Scan() (token.Token, string) {
	// Read the read rune.
	ch := s.read()

	// If we see a whitespace then consume all contiguous whitespace.
	// If we see a newline then consume all contiguous newline.
	// If we see an exclamation mark then consume as a comment.
	// If we see a dot then consume as a directive.
	// If we see a digit or - consume as an integer.
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
	} else if isDigit(ch) {
		s.unread()
		return s.scanInteger()
	} else if isLetter(ch) || ch == '%' {
		s.unread()
		return s.scanIdent()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return token.EOF, ""
	case '+':
		return token.PLUS, string(ch)
	case '-':
		return token.MINUS, string(ch)
	case '[':
		return token.LBRACKET, string(ch)
	case ']':
		return token.RBRACKET, string(ch)
	case ',':
		return token.COMMA, string(ch)
	case ':':
		return token.COLON, string(ch)
	}

	// No match results in an illegal token.
	return token.ILLEGAL, string(ch)
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Check if the identifier is a keyword.
	if tok := token.Lookup(buf.String()); tok.IsKeyword() {
		return tok, buf.String()
	}

	// Otherwise return as a regular identifier.
	return token.IDENT, buf.String()
}

// scanInteger consumes the current rune and all contiguous integer runes.
func (s *Scanner) scanInteger() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent integer character into the buffer.
	// Non-integer characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isDigit(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Return as an integer.
	return token.INT, buf.String()
}

// scanDirective consumes the current rune and all contiguous directive runes.
func (s *Scanner) scanDirective() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent directive character into the buffer.
	// Non-directive characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
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
		return tok, buf.String()
	}

	// Otherwise return an ILLEGAL token (because it can't be an identifier
	// starting with a '.').
	return token.ILLEGAL, buf.String()
}

// scanComment consumes the current rune and all contiguous comment runes.
func (s *Scanner) scanComment() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character into the buffer.
	// EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	// Return comment with text as literal value.
	return token.COMMENT, buf.String()
}

// scanNewline consumes the current rune and all contiguous newline.
func (s *Scanner) scanNewline() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isNewline(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return token.NL, buf.String()
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (token.Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return token.WS, buf.String()
}

// read reads the read rune from the bufferred reader. Returns the rune(0) if an
// error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { s.r.UnreadRune() }

// isWhitespace returns true if the rune is a space or tab.
func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' }

// isNewline returns true if the rune is a newline.
func isNewline(ch rune) bool { return ch == '\n' || ch == '\r' }

// isLetter returns true if the rune is a letter.
func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// isDigit returns true if the rune is a digit.
func isDigit(ch rune) bool { return (ch >= '0' && ch <= '9') }
