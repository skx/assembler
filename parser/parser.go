package parser

import (
	"fmt"

	"github.com/skx/assembler/lexer"
	"github.com/skx/assembler/token"
)

// Instruction holds a parsed instruction
type Instruction struct {

	// instruction holds the instruction we've found.
	Instruction token.Token

	// Operands holds the operands for this instruction.
	//
	// This will usually be an integer, a pair of registers,
	// or a register and an integer
	Operands []token.Token
}

// Parse holds our state.
type Parser struct {
	// program is our program, as a series of tokens
	program []token.Token

	// Position within the string
	position int
}

// New creates a new Parse, which will parse the specified
// input program into a series of tokens.
func New(input string) *Parser {

	// Create our parser
	p := &Parser{}

	// Create the lexer object.
	l := lexer.New(input)

	// Parse our program into a series of tokens
	tok := l.NextToken()
	for tok.Type != token.EOF {

		p.program = append(p.program, tok)
		tok = l.NextToken()
	}

	return p

}

func (p *Parser) NextToken() *Instruction {

	// Loop until we've exhausted our input.
	for p.position < len(p.program) {

		// The token we're operating upon
		tok := p.program[p.position]

		switch tok.Literal {

		case "inc":
			// We want one argument
			args, err := p.TakeOneArgument()
			if err != nil {
				return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: err.Error()}}

			}
			return &Instruction{Instruction: tok, Operands: args}

		case "mov":
			// We want two arguments
			args, err := p.TakeTwoArguments()
			if err != nil {
				return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: err.Error()}}

			}
			return &Instruction{Instruction: tok, Operands: args}

		case "add":
			// We want two arguments
			args, err := p.TakeTwoArguments()
			if err != nil {
				return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: err.Error()}}

			}
			return &Instruction{Instruction: tok, Operands: args}

		case "int":
			// We want one argument
			args, err := p.TakeOneArgument()
			if err != nil {
				return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: err.Error()}}

			}
			return &Instruction{Instruction: tok, Operands: args}

		case "xor":
			// We want two arguments
			args, err := p.TakeTwoArguments()
			if err != nil {
				return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: err.Error()}}

			}
			return &Instruction{Instruction: tok, Operands: args}
		}

		return &Instruction{Instruction: token.Token{Type: token.ILLEGAL, Literal: fmt.Sprintf("unknown instruction %v", tok)}}

	}

	return &Instruction{Instruction: token.Token{Type: token.EOF}}
}

func (p *Parser) TakeTwoArguments() ([]token.Token, error) {

	var toks []token.Token

	// skip the instruction
	p.position++

	// ensure we're not out of the program
	if p.position >= len(p.program) {
		return toks, fmt.Errorf("unexpected EOF")
	}

	// add the argument
	one := p.program[p.position]
	if one.Type != token.REGISTER && one.Type != token.NUMBER {
		return toks, fmt.Errorf("expected REG|NUM, got %v", one)
	}
	toks = append(toks, one)

	// Skip the comma
	p.position++
	if p.position >= len(p.program) {
		return toks, fmt.Errorf("unexpected EOF")
	}
	c := p.program[p.position]
	if c.Type != token.COMMA {
		return toks, fmt.Errorf("expected ',', got %v", c)
	}

	// Get the second arg.
	p.position++
	if p.position >= len(p.program) {
		return toks, fmt.Errorf("unexpected EOF")
	}
	two := p.program[p.position]
	if two.Type != token.NUMBER && two.Type != token.REGISTER {
		return toks, fmt.Errorf("expected REGISTER|NUMBER, got %v", two)
	}
	toks = append(toks, two)

	p.position++

	return toks, nil
}

func (p *Parser) TakeOneArgument() ([]token.Token, error) {

	var toks []token.Token

	// skip the instruction
	p.position++

	// ensure we're not out of the program
	if p.position >= len(p.program) {
		return toks, fmt.Errorf("unexpected EOF")
	}

	// add the argument
	one := p.program[p.position]
	if one.Type != token.REGISTER && one.Type != token.NUMBER {
		return toks, fmt.Errorf("expected REG|NUM, got %v", one)
	}
	toks = append(toks, one)

	p.position++

	return toks, nil
}
