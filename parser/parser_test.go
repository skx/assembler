package parser

import (
	"testing"
)

func TestComment(t *testing.T) {

	p := New(";; This is a test")

	out := p.Next()
	if out != nil {
		t.Fatalf("Failed to skip comment")
	}
}

func TestMove(t *testing.T) {

	p := New("mov rax, rbx")

	out := p.Next()

	outI, ok := out.(Instruction)
	if !ok {
		t.Fatalf("didn't get an instruction structure")
	}

	if len(outI.Operands) != 2 {
		t.Fatalf("mov - wrong arg count")
	}
	if outI.Operands[0].Literal != "rax" {
		t.Fatalf("mov - wrong first arg")
	}
	if outI.Operands[1].Literal != "rbx" {
		t.Fatalf("mov - wrong second arg")
	}
}
