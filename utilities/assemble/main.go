package main

import (
	"flag"
	"fmt"
	"os"

	avrassembler "avrassembler"
)

// CmdArgs holds input and output file names
type cmdArgs struct {
	InputFile  string
	OutputFile string
}

// ParseCmdArgs parses command-line arguments for input and output files
func parseCmdArgs() (*cmdArgs, error) {
	input := flag.String("i", "", "Input assembly file (.S)")
	output := flag.String("o", "output.hex", "Output binary file (.hex)")

	flag.Parse()

	if *input == "" {
		return nil, fmt.Errorf("input file must be specified with -i")
	}

	// Check if input file exists
	if _, err := os.Stat(*input); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", *input)
	}

	return &cmdArgs{
		InputFile:  *input,
		OutputFile: *output,
	}, nil
}

func main() {
	args, err := parseCmdArgs()
	if err != nil {
		fmt.Printf("[E] %s\n", err)
		os.Exit(1)
	}
	avrassembler.ParseFile(args.InputFile, 0x0000)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	err = avrassembler.WriteToFile(args.OutputFile)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
