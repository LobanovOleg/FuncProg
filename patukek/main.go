package main

import (
	"patukek/internal/patukek_compiler"
	"patukek/internal/patukek_parser"
	"patukek/internal/patukek_vm"
	"flag"
	"fmt"
	"os"
)

func readFile(fname string) []byte {
	b, err := os.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return b
}

func compile(path string) (bc *patukek_compiler.Bytecode, err error) {
	input := string(readFile(path))
	res, _ := patukek_parser.Parse(path, input)

	c := patukek_compiler.New()
	c.SetFileInfo(path, input)
	if err = c.Compile(res); err != nil {
		return
	}

	return c.Bytecode(), nil
}

func execFileVM(f string) (err error) {
	var bytecode *patukek_compiler.Bytecode
	bytecode, err = compile(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	tvm := patukek_vm.New(f, bytecode)
	if err = tvm.Run(); err != nil {
		fmt.Println(err)
		return
	}
	return
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("See instructions to run patukek programs in REPORT.md")
	} else {
		_ = execFileVM(flag.Arg(0))
	}
}