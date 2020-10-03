// Package parser consumes tokens from the lexer, and generates the AST
// which is then walked to generate binary code.
package parser

import (
	"fmt"
	"strconv"

	"github.com/skx/assembler/instructions"
	"github.com/skx/assembler/lexer"
	"github.com/skx/assembler/token"
)

// Parser holds our state.
type Parser struct {
	// program holds our lexed program, as a series of tokens.
	program []token.Token

	// position holds our current offset within the program
	// above.
	position int
}

// New creates a new Parser, which will parse the specified
// input program into a series of tokens, and then allow it
// to be parsed.
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

	// Now we have a parser complete with a series of tokens
	return p

}

// Next returns the stream of parsed "things" from the input source program.
//
// The things we return include:
//
//  * Instructions.
//  * Label definitions.
//  * Data references.
//
// There might be more things in the future.
func (p *Parser) Next() Node {

	// Loop until we've exhausted our input.
	for p.position < len(p.program) {

		// The token we're operating upon
		tok := p.program[p.position]

		switch tok.Type {

		case token.DATA:
			return p.parseData()

		case token.INSTRUCTION:
			return p.parseInstruction()

		case token.LABEL:
			return p.parseLabel()
		}
	}

	return nil
}

// parseData handles input of the form:
//
//  .NAME DB "String content here"
//
// TODO:
//
//  .NAME DB 0x01, 0x02, 0x03 ...
func (p *Parser) parseData() Node {

	// create the data-structure, with the name.
	d := Data{Name: p.program[p.position].Literal}

	// skip the DATA
	p.position++

	// ensure we're not out of the program
	if p.position >= len(p.program) {
		return Error{Value: "Unexpected EOF parsing data"}
	}

	// Next token should be DB
	db := p.program[p.position]
	if db.Type != token.DB {
		return Error{Value: fmt.Sprintf("expected DB, got %v", db)}
	}

	// move forward
	p.position++
	if p.position >= len(p.program) {
		return Error{Value: "Unexpected EOF parsing data"}
	}

	//
	// We support:
	//   .foo DB "String"
	//
	// Or
	//   .foo DB 0x03, 0x4...
	//
	// If the next token is a string handle that.
	cur := p.program[p.position]
	if cur.Type == token.STRING {
		// bump past the string
		p.position++

		d.Contents = []byte(cur.Literal)
		return d
	}

	// If the type isn't a number that's an error
	if cur.Type != token.NUMBER {
		return Error{Value: fmt.Sprintf("expected string|number-array, got %v", cur)}
	}

	// OK so we've got number
	for cur.Type == token.NUMBER {

		// Parse it
		num, err := strconv.ParseInt(cur.Literal, 0, 64)
		if err != nil {
			return Error{Value: fmt.Sprintf("failed to convert '%s' to number:%s", cur.Literal, err)}
		}

		// Add to the array
		d.Contents = append(d.Contents, byte(num))

		// skip past the number
		p.position++

		// end of program?
		if p.position >= len(p.program) {
			break
		}

		// if the next token is not a comma then we're done
		if p.program[p.position].Type != token.COMMA {
			break
		}

		// Otherwise skip over the comma
		p.position++

		// end of program?
		if p.position >= len(p.program) {
			break
		}

		cur = p.program[p.position]
	}

	return d
}

// parseInstruction is our workhorse
//
// We either return an `Instruction` or an `Error`
//
func (p *Parser) parseInstruction() Node {

	// Get the current instruction
	tok := p.program[p.position]

	// Find out how many arguments it has
	count, ok := instructions.InstructionLengths[tok.Literal]

	// If that failed then it is an unknown instruction, probably
	if !ok {
		return Error{Value: fmt.Sprintf("unknown instructoin %v", tok)}
	}

	// No args?  Just return the instruction and bump the position
	if count == 0 {
		p.position++
		return Instruction{Instruction: tok.Literal}
	}

	if count == 1 {
		args, err := p.TakeOneArgument()
		if err != nil {
			return Error{Value: err.Error()}

		}
		return Instruction{Instruction: tok.Literal, Operands: args}

	}
	if count == 2 {

		args, err := p.TakeTwoArguments()
		if err != nil {
			return Error{Value: err.Error()}

		}
		return Instruction{Instruction: tok.Literal, Operands: args}
	}

	return Error{Value: fmt.Sprintf("unhandled argument-count for token %v", tok)}
}

// parseLabel handles input of the form:
//
//  :foo
func (p *Parser) parseLabel() Node {

	// create the label-structure, with the name.
	l := Label{Name: p.program[p.position].Literal}

	// skip the label itself
	p.position++

	return l
}

// TakeTwoArguments handles fetching two arguments for an instruction.
//
// Arguments may be register-names, numbers, or label-values
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
	if one.Type != token.REGISTER && one.Type != token.NUMBER && one.Type != token.IDENTIFIER {
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
	if two.Type != token.NUMBER && two.Type != token.REGISTER && two.Type != token.IDENTIFIER {
		return toks, fmt.Errorf("expected REGISTER|NUMBER|IDENTIFIER, got %v", two)
	}
	toks = append(toks, two)

	p.position++

	return toks, nil
}

// TakeOneArgument reads the argument for a single-arg instruction.
//
// Arguments may be a register-name, number, or a label-value.
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
	if one.Type != token.REGISTER && one.Type != token.NUMBER && one.Type != token.IDENTIFIER {
		return toks, fmt.Errorf("expected REGISTER|NUMBER|IDENTIFIER, got %v", one)
	}
	toks = append(toks, one)

	p.position++

	return toks, nil
}
