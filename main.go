package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jheredos/rye/interpreter"
)

func main() {
	if len(os.Args) > 2 {
		os.Exit(1)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	file, err := ioutil.ReadFile(path) // read file
	if err != nil {
		panic(err)
	}

	// scan...
	ts := interpreter.Scan(string(file))
	// for _, t := range ts {
	// 	fmt.Println(t.ToString())
	// }

	root, err := interpreter.Parse(ts)
	if err != nil {
		fmt.Println(err)
		return
	}

	// execute...
	env := &interpreter.Environment{
		Parent: &interpreter.Environment{
			// env above the "top-level" for imports
			Consts: interpreter.StdLib,
		},
		Consts: map[string]*interpreter.Node{},
		Vars:   map[string]*interpreter.Node{},
	}
	_, err = interpreter.Interpret(root, env)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)
	env := &interpreter.Environment{
		Parent: &interpreter.Environment{
			// env above the "top-level" for imports
			Consts: interpreter.StdLib,
		},
		Consts: map[string]*interpreter.Node{},
		Vars:   map[string]*interpreter.Node{},
	}

	for {
		fmt.Print("> ")
		inp, err := reader.ReadString('\n') // read line

		// scan...
		ts := interpreter.Scan(inp)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// parse...
		root, err := interpreter.Parse(ts)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
		if root == nil {
			continue
		}

		// execute...
		res, err := interpreter.Interpret(root, env)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(interpreter.Display(res))
	}
}
