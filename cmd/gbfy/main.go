package main

import (
	"bufio"
	"flag"
	"fmt"
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

	// Read contents of codeFile and inputFile to get commands
	// to execute and any input data the program may expect.
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

	bf, err := gbfy.New(string(cmds), input)
	if err != nil {
		log.Fatalf("Failed to initialize interpreter: %v", err)
	}

	output, err := bf.Run()
	if err != nil {
		log.Printf("Run failed with error: %v", err)
	}

	log.Printf("Output from execution: %q\n", output)

	// Drop user into interactive "REPL" if requested. Execution
	// may continue in the REPL with whatever interpreter state
	// state the above program left it in.
	if *launchREPL {
		log.Println("... starting interactive REPL")
		beInteractive(bf)
	}
}

const welcomeMsg = `Welcome!
 :q[uit] to exit REPL loop 
 :d[ump] to dump interpreter state
 :f[uck] to reset interpreter state`

func beInteractive(bf *gbfy.Brainfuck) {
	fmt.Println(welcomeMsg)
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("gbfy> ")
		line, _ := in.ReadString('\n')
		if line[0] == ':' {
			// Handle REPL commands.
			switch line[1] {
			case 'q': // exit REPL loop.
				return
			case 'd': // dump state to stdout.
				fmt.Println(bf.String())
			case 'f': // reset interpreter.
				bf.Reset()
				fmt.Println("Restarted interpreter...")
			}
		} else {
			// Handle a Brainfuck command(s).
		}
	}
}
