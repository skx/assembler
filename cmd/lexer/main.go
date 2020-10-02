package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/assembler/lexer"
	"github.com/skx/assembler/token"
)

func main() {
	//
	// Ensure we have an argument
	//
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: lexer input.asm\n")
		return
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("error:%s\n", err.Error())
		return
	}

	l := lexer.New(string(data))

	tok := l.NextToken()
	for tok.Type != token.EOF {
		fmt.Printf("%v\n", tok)
		tok = l.NextToken()
	}
}
