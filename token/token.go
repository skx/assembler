// Package token contains identifiers for the various things
// we find in our source-scripts.
//
// Our lexer will convert an input-script into a series of tokens,
// which will then be further-processed.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {

	// Type contains the type of the token.
	Type Type

	// Literal contains the literal text of the token.
	Literal string
}

// Our known token-types
const (
	// Basic things
	COMMA       = ","
	EOF         = "EOF"
	LABEL       = "LABEL"
	DATA        = "DATA"
	REGISTER    = "REGISTER"
	INSTRUCTION = "INSTRUCTION"
	IDENTIFIER  = "IDENTIFIER"

	// Data statement
	DB = "DB"

	// Number as operand
	NUMBER = "NUMBER"

	// String for DB
	STRING = "STRING"

	// Something we couldn't handle
	ILLEGAL = "ILLEGAL"
)

// known things we can handle
var known = map[string]Type{
	"DB": DB,

	// instructions we handle
	"mov": INSTRUCTION,
	"xor": INSTRUCTION,
	"inc": INSTRUCTION,
	"add": INSTRUCTION,
	"int": INSTRUCTION,
	"nop": INSTRUCTION,

	// registers
	// TODO: more
	"rax": REGISTER,
	"rbx": REGISTER,
	"rcx": REGISTER,
	"rdx": REGISTER,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {

	if tok, ok := known[identifier]; ok {
		return tok
	}
	return IDENTIFIER
}
