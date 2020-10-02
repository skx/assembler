package parser

import (
	"testing"

	"github.com/skx/assembler/token"
)

func TestComment(t *testing.T) {

	p := New(";; This is a test")

	out := p.NextToken()
	if out.Instruction.Type != token.EOF {
		t.Fatalf("Failed to skip comment")
	}
}

func TestMove(t *testing.T) {

	p := New("mov rax, rbx")

	out := p.NextToken()
	if out.Instruction.Literal != "mov" {
		t.Fatalf("Failed to find mov")
	}

	if len(out.Operands) != 2 {
		t.Fatalf("mov - wrong arg count")
	}
	if out.Operands[0].Literal != "rax" {
		t.Fatalf("mov - wrong first arg")
	}
	if out.Operands[1].Literal != "rbx" {
		t.Fatalf("mov - wrong second arg")
	}
}
