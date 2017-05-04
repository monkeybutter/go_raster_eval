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
	input := `B5 # (BQA == 32768);`
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
	//fmt.Println("AAAA", obj.Type(), obj.Inspect(), env)
	fmt.Println("AAAA", obj.(*object.Raster).Value)
}


/*
package main

import "fmt"

func main() {
    val := uint16(24576)
    isCloud := uint16(32768)
    fmt.Printf("%b\n", val)
    val = val & isCloud
    fmt.Println(val > 0)
}
*/
