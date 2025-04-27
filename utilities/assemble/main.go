package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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

	file, err := os.Open(args.InputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	instructions := []avrassembler.Instruction{}

	// Collect Instuctions and Labels
	line := uint16(0)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		instruction, label, err := avrassembler.ParseLine(scanner.Text(), line)
		if err != nil {
			fmt.Printf("Error on line %d, %s\n ", line, err)
			os.Exit(1)
		}
		if label != "" {
			avrassembler.LabelMap[label] = line
		}
		// if white space, comment, or label skip instruction logic
		if instruction.Mnemonic == "" {
			continue
		}
		instructions = append(instructions, instruction)

		line++
	}

	compiled_assembly := []string{}
	// Parse Operands with context of all labels
	for i := 0; i < len(instructions); i++ {
		encodingFunc, ok := avrassembler.InstructionParse[instructions[i].Mnemonic]
		if !ok {
			fmt.Printf("[E] Parsing function not found for %s not found on line %d\n", instructions[i].Mnemonic, instructions[i].Line)
			os.Exit(1)
		}
		ops, err := encodingFunc(instructions[i].Operands, instructions[i].Line)
		if err != nil {
			fmt.Printf("[E] %s, Found on line %d\n", err, instructions[i].Line)
			os.Exit(1)
		}
		ins, ok := avrassembler.InstructionSet[instructions[i].Mnemonic]
		if !ok {
			fmt.Printf("[E] Encoding function not found for %s on line %d\n", instructions[i].Mnemonic, instructions[i].Line)
			os.Exit(1)
		}

		enc := ins.Encode(ins.ByteCode, ops[0], ops[1])

		le_enc := ((enc[0] >> 8) & 0x00ff) | ((enc[0] << 8) & 0xff00)
		hex := fmt.Sprintf("%x", le_enc)
		hex = fmt.Sprintf("%04s", hex)
		compiled_assembly = append(compiled_assembly, hex)
		fmt.Printf("%6s %04s\n", instructions[i].Mnemonic, hex)
	}

	file_content, err := avrassembler.ToIntelHex(compiled_assembly)
	if err != nil {
		println(err.Error())
	}
	println(file_content)
	os.Remove(args.OutputFile)
	f, err := os.Create(args.OutputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString(file_content)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}
