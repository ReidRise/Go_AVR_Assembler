package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	avrassembler "avrassembler"
)

func parseImmidiateUints(num string) (imm uint16, err error) {
	im, err := strconv.ParseUint(num, 10, 16)
	if err == nil {
		return uint16(im), nil
	} else if num[0:2] == "0b" {
		imm, err := strconv.ParseUint(num[2:], 2, 16)
		if err != nil {
			return 0, err
		}
		return uint16(imm), nil
	} else if num[0:2] == "0x" {
		imm, err := strconv.ParseUint(num[2:], 16, 16)
		if err != nil {
			return 0, err
		}
		return uint16(imm), nil
	} else {
		return 0, fmt.Errorf(" unable to parse [%s] into uint", num)
	}
}

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

	startAddress := uint16(0x00)
	instructions := []avrassembler.Instruction{}
	// Collect Instuctions and Labels
	line := uint16(0)
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		instruction, meta, err := avrassembler.ParseLine(scanner.Text(), line)
		if err != nil {
			fmt.Printf("Error on line %d, %s\n ", line, err)
			os.Exit(1)
		}

		if meta.Operation == "label" {
			avrassembler.LabelMap[meta.Args] = line + (startAddress / 2)
		}

		if meta.Operation == "org" {
			avrassembler.RawAssemblySections[startAddress] = instructions
			instructions = []avrassembler.Instruction{}
			startAddress, err = parseImmidiateUints(meta.Args)
			line = 0
			if err != nil {
				fmt.Printf("Error parsing address %s, %s\n ", meta.Args, err)
				os.Exit(1)
			}
		}
		// if white space, comment, or meta skip instruction logic
		if instruction.Mnemonic == "" {
			continue
		}
		instructions = append(instructions, instruction)

		line++
	}
	avrassembler.RawAssemblySections[startAddress] = instructions

	fileOut := ""
	// Parse Operands with context of all labels
	for addr, instructionSection := range avrassembler.RawAssemblySections {
		compiledAssembly := []string{}
		for i := 0; i < len(instructionSection); i++ {
			encodingFunc, ok := avrassembler.InstructionParse[instructionSection[i].Mnemonic]
			if !ok {
				fmt.Printf("[E] Parsing function not found for %s not found on line %d\n", instructionSection[i].Mnemonic, instructionSection[i].Line)
				os.Exit(1)
			}
			ops, err := encodingFunc(instructionSection[i].Operands, instructionSection[i].Line)
			if err != nil {
				fmt.Printf("[E] %s, Found on line %d\n", err, instructionSection[i].Line)
				os.Exit(1)
			}
			ins, ok := avrassembler.InstructionSet[instructionSection[i].Mnemonic]
			if !ok {
				fmt.Printf("[E] Encoding function not found for %s on line %d\n", instructionSection[i].Mnemonic, instructionSection[i].Line)
				os.Exit(1)
			}

			enc := ins.Encode(ins.ByteCode, ops[0], ops[1])

			le_enc := ((enc[0] >> 8) & 0x00ff) | ((enc[0] << 8) & 0xff00)
			hex := fmt.Sprintf("%x", le_enc)
			hex = fmt.Sprintf("%04s", hex)
			compiledAssembly = append(compiledAssembly, hex)
			fmt.Printf("%6s %04s\n", instructionSection[i].Mnemonic, hex)
		}

		fileContent, err := avrassembler.ToIntelHex(compiledAssembly, int(addr))
		if err != nil {
			println(err.Error())
		}
		fileOut += fileContent
	}
	fileOut += ":00000001FF"
	println(fileOut)
	os.Remove(args.OutputFile)
	f, err := os.Create(args.OutputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString(fileOut)
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
