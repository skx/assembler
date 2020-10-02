package lexer

import (
	"testing"

	"github.com/skx/assembler/token"
)

func TestComment(t *testing.T) {

	n := New("; This is a comment")

	tok := n.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected end of file")
	}
}

func TestMov(t *testing.T) {

	input := `
;; Two move instructions
mov rax, rcx
mov rbx, 33
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INSTRUCTION, "mov"},
		{token.REGISTER, "rax"},
		{token.COMMA, ","},
		{token.REGISTER, "rcx"},

		{token.INSTRUCTION, "mov"},
		{token.REGISTER, "rbx"},
		{token.COMMA, ","},
		{token.NUMBER, "33"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}

}
