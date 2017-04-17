package main

import (
	"./evaluator"
	"./lexer"
	"./object"
	"./parser"
	"fmt"
)

func main() {
	//input := `(B5 - B4) / (B5 + B4);`
	input := `B5 # (B5 == 4);`
	//input := `((B4+1) / B5) + 1;`
	//input := `((B5 + 1) / B3) + 1;`
	l := lexer.New(input)

	p := parser.New(l)
	fmt.Println(p)
	prog := p.ParseProgram()
	for _, s := range prog.Statements {
		fmt.Println(s)
	}
	env := object.NewEnvironment()
	obj := evaluator.Eval(prog, env)
	fmt.Println("AAAA", obj.Type(), obj.Inspect(), env)
}
