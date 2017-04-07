package main

import (
	"./lexer"
	"./parser"
	"fmt"
	"./evaluator"
	"./object"
)

func main() {
	input := `(B5 - B4) / (B5 + B4);`
	l := lexer.New(input)
	
	p := parser.New(l)

	prog := p.ParseProgram()
	for _, s := range prog.Statements {
		fmt.Println(s)
	}
	env := object.NewEnvironment()
	obj := evaluator.Eval(prog, env)
	fmt.Println("AAAA", obj.Type(), obj.Inspect(), env)
}
