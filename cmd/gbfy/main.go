package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/abtrout/gbfy"
)

var (
	codeFile   = flag.String("f", "", "Path to Brainfuck file to interpret")
	inputFile  = flag.String("i", "", "Path to file containing input data for program")
	launchREPL = flag.Bool("repl", false, "Launch interactive interpreter before exit")
)

func main() {
	flag.Parse()

	var (
		err   error
		cmds  []byte
		input []byte
	)
	if *codeFile != "" {
		cmds, err = os.ReadFile(*codeFile)
		if err != nil {
			log.Fatalf("Failed to read code file: %v", err)
		}
	}
	if *inputFile != "" {
		input, err = os.ReadFile(*inputFile)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}
	}

	var output bytes.Buffer
	bf := gbfy.New(bytes.NewBuffer(input), &output)
	if err != nil {
		log.Fatalf("Failed to initialize interpreter: %v", err)
	}

	if len(cmds) > 0 {
		for _, cmd := range cmds {
			if err := bf.Eval(cmd); err != nil {
				log.Fatalf("Failed to Eval command %q: %v", cmd, err)
			}
		}
		log.Printf("Output from execution: %v\n", output.Bytes())
	}

	if *launchREPL {
		log.Println("Starting interactive REPL")
		repl(bf)
	}
}

const (
	welcomeMsg = `Go Brainfuck Yourself!
 :q[uit] to exit REPL loop 
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state`

	prompt = "gbfy> "
)

func repl(bf *gbfy.Brainfuck) {
	fmt.Println(welcomeMsg)
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		line, err := in.ReadString('\n')
		if err == io.EOF {
			fmt.Println("")
			return
		} else if err != nil {
			log.Fatalf("Error! %v", err)
		}

		if len(line) > 0 && line[0] == ':' {
			switch line[1] {
			case 'd':
				replDump(bf)
			case 'f':
				bf.Reset()
				fmt.Println("Reset interpreter!")
			case 'q':
				return
			}
		} else {
			for _, cmd := range []byte(line) {
				if err := bf.Eval(cmd); err != nil {
					log.Fatalf("Failed to Eval command %q: %v", cmd, err)
				}
			}
		}
	}
}

func replDump(bf *gbfy.Brainfuck) {
	d, cells, i, cmds, out := bf.Dump()
	fmt.Printf("Cells; d: %d, current cell value: %x\n", d, cells[d])
	fmt.Printf("Cmds: i: %d, current command: %q\n", i, cmds[i-1])
	fmt.Printf("Out: %v\n", out)
}
