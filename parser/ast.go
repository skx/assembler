package parser

import (
	"fmt"
	"github.com/skx/assembler/token"
)

// Node is something we return from our parser.
type Node interface {
	// Output this as a readable string
	String() string
}

// Error contains an error-message
type Error struct {
	Node
	Value string
}

// String outputs this Error structure as a string.
func (e Error) String() string {
	return fmt.Sprintf("<ERROR:%s>", e.Value)
}

// Data holds a data-statement, which might look like either of these:
//
//   .foo DB "Steve"
//   .bar DB 0x030, 0x40, 0x90
//
type Data struct {
	Node

	// Name is the name of the data-section
	Name string

	// Contents holds the string/byte data for the reference
	Contents []byte
}

// String outputs this Data structure as a string.
func (d Data) String() string {
	return fmt.Sprintf("<DATA: name:%s data:%v>", d.Name, d.Contents)
}

// Instruction holds a parsed instruction.
//
// For example "mov rax, rax".
//
type Instruction struct {
	Node

	// Instruction holds the instruction we've found, as a string.
	Instruction string

	// Operands holds the operands for this instruction.
	//
	// This will usually be an integer, a pair of registers,
	// or a register and an integer
	Operands []token.Token
}

// String outputs this Error structure as a string
func (d Instruction) String() string {
	return fmt.Sprintf("<INSTRUCTION: %s args:%v>", d.Instruction, d.Operands)
}

// Label holds a label, as seen when it is defined.
//
// For example ":foo" will define a label with name "foo".
type Label struct {
	Node

	// Name has the name of the instruction
	Name string
}

// String outputs this Label structure as a string.
func (l Label) String() string {
	return fmt.Sprintf("<LABEL: %s>", l.Name)
}
