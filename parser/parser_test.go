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

func TestData(t *testing.T) {

	type TestCase struct {
		Input string
		Data  []byte
	}

	tests := []TestCase{
		TestCase{Input: ".data DB \"Steve\"",
			Data: []byte{83, 116, 101, 118, 101},
		},
		TestCase{Input: ".foo DB 1\n.bar DB 3,3",
			Data: []byte{1},
		},
		TestCase{Input: ".foo DB 32, 44",
			Data: []byte{32, 44},
		},
		TestCase{Input: ".foo DB 32, ",
			Data: []byte{32},
		},
	}

	// For each test
	for _, test := range tests {

		// Parse
		p := New(test.Input)

		// We expect a single data-statement
		out := p.Next()
		if out == nil {
			t.Fatalf("nil result from pasing %s", test.Input)
		}

		// Cast to the right value
		d, ok := out.(Data)
		if !ok {
			t.Fatalf("didn't get an Data structure: %v", out)
		}

		// Length matches?
		if len(d.Contents) != len(test.Data) {
			t.Fatalf("data length didn't match expectation")
		}

		// Content matches?
		for i, x := range d.Contents {
			if test.Data[i] != x {
				t.Fatalf("data mismatch at offset %d", i)
			}
		}
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
