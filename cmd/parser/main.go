package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/assembler/parser"
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

	stmt := p.Next()
	for stmt != nil {
		fmt.Printf("%v\n", stmt)

		stmt = p.Next()
	}
}
