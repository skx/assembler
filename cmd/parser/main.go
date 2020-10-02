package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/assembler/parser"
	"github.com/skx/assembler/token"
)

func main() {
	//
	// Ensure we have an argument
	//
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: parser input.asm\n")
		return
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("error:%s\n", err.Error())
		return
	}

	p := parser.New(string(data))

	i := p.NextToken()
	for i.Instruction.Type != token.EOF {
		fmt.Printf("%v\n", i)

		// Illegal?  Then stop
		if i.Instruction.Type == token.ILLEGAL {
			return
		}
		i = p.NextToken()
	}
}
