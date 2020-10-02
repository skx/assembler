// Package lexer contains our lexer.
package lexer

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/skx/assembler/token"
)

// Lexer holds our object-state.
type Lexer struct {
	// The current character position
	position int

	// The next character position
	readPosition int

	// The current character
	ch rune

	// A rune slice of our input string
	characters []rune
}

// New creates a Lexer instance from the given string
func New(input string) *Lexer {

	// Line counting starts at one.
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// read forward one character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

}

// NextToken reads and returns the next token, skipping any intervening
// white space, and swallowing any comments, in the process.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	// skip single-line comments
	if l.ch == rune(';') {
		l.skipComment()
		return (l.NextToken())
	}
	if l.ch == rune('#') {
		l.skipComment()
		return (l.NextToken())
	}

	switch l.ch {

	case rune(':'):
		label, err := l.readLabel()
		if err != nil {
			tok.Literal = err.Error()
			tok.Type = token.ILLEGAL
		} else {
			tok = token.Token{Type: token.LABEL, Literal: label}
		}

	case rune('.'):
		label, err := l.readLabel()
		if err != nil {
			tok.Literal = err.Error()
			tok.Type = token.ILLEGAL
		} else {
			tok = token.Token{Type: token.DATA, Literal: label}
		}

	case rune(','):
		tok = token.Token{Type: token.COMMA, Literal: ","}
	case rune('"'):
		str, err := l.readString('"')
		if err == nil {
			tok.Literal = str
			tok.Type = token.STRING
		} else {
			tok.Literal = err.Error()
			tok.Type = token.ILLEGAL
		}
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		// Number
		if isDigit(l.ch) {
			tok := l.readDecimal()
			return tok
		}

		// Instruction/Register
		tok.Literal = l.readIdentifier()
		if len(tok.Literal) > 0 {
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		}

		// Not an instruction/register (+LABEL)
		tok.Type = token.IDENTIFIER
		return tok

	}

	l.readChar()

	return tok
}

// readIdentifier is designed to read an identifier (name of variable,
// function, etc).
func (l *Lexer) readIdentifier() string {

	id := ""

	for isIdentifier(l.ch) {
		id += string(l.ch)
		l.readChar()
	}
	return id
}

// skip over any white space.
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip a comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read a number.  We only care about numerical digits here, floats will
// be handled elsewhere.
func (l *Lexer) readNumber() string {

	id := ""

	for isDigit(l.ch) || l.ch == rune('x') {
		id += string(l.ch)
		l.readChar()
	}
	return id
}

// read a decimal number, either int or floating-point.
func (l *Lexer) readDecimal() token.Token {

	//
	// Read an integer-number.
	//
	integer := l.readNumber()

	//
	// Just an integer.
	//
	return token.Token{Type: token.NUMBER, Literal: integer}
}

// read a string, deliminated by the given character.
func (l *Lexer) readString(delim rune) (string, error) {
	out := ""

	for {
		l.readChar()

		if l.ch == rune(0) {
			return "", fmt.Errorf("unterminated string")
		}
		if l.ch == delim {
			break
		}
		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {

			// Line ending with "\" + newline
			if l.peekChar() == '\n' {
				// consume the newline.
				l.readChar()
				continue
			}

			l.readChar()

			if l.ch == rune(0) {
				return "", errors.New("unterminated string")
			}
			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)

	}

	return out, nil
}

// read a label
func (l *Lexer) readLabel() (string, error) {
	out := ""

	for {
		l.readChar()

		if l.ch == rune(0) {
			if len(out) > 1 {
				return out, nil
			}
			return "", fmt.Errorf("unterminated label")
		}
		if isWhitespace(l.ch) {
			return out, nil
		}
		out = out + string(l.ch)
	}

	return out, nil
}

// determinate ch is identifier or not.  Identifiers may be alphanumeric,
// but they must start with a letter.  Here that works because we are only
// called if the first character is alphabetical.
func isIdentifier(ch rune) bool {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '$' || ch == '_' || ch == '-' {
		return true
	}
	return false
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}
