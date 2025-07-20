package main

import (
	"bufio"
	"encoding/hex"
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
	inMacroDef := ""
	// Line in file
	codeLine := uint16(0)
	// Line in raw assembly section
	chunkLine := uint16(0)
	for scanner.Scan() {
		codeLine++
		instruction, meta, err := avrassembler.ParseLine(scanner.Text())
		if err != nil {
			fmt.Printf("Error on line %d, %s\n ", chunkLine, err)
			os.Exit(1)
		}
		instruction.Address = int(chunkLine)
		instruction.Line = int(codeLine)
		for _, m := range meta {
			if m.Operation == "label" {
				if inMacroDef != "" {
					fmt.Printf("[E] Labels cannot be created in macros\n")
					os.Exit(1)
				}
				//fmt.Printf("%s added to labelmap at %d\n", m.Args, line+(startAddress/2))
				avrassembler.LabelMap[m.Args] = chunkLine + (startAddress / 2)
			}

			if m.Operation == "org" {
				if inMacroDef != "" {
					fmt.Printf("[E] Cannot define origin inside macros\n")
					os.Exit(1)
				}
				avrassembler.RawAssemblySections[startAddress] = instructions
				instructions = []avrassembler.Instruction{}
				startAddress, err = parseImmidiateUints(m.Args)
				chunkLine = 0
				if err != nil {
					fmt.Printf("Error parsing address %s, %s\n ", m.Args, err)
					os.Exit(1)
				}
				if (startAddress % 2) != 0 {
					fmt.Printf("[W] Address %s is not 16 bit aligned!\n ", m.Args)
				}
			}

			if m.Operation == "db" {
				if inMacroDef != "" {
					fmt.Printf("[E] Cannot define data blob in macro\n")
					os.Exit(1)
				}
				avrassembler.RawAssemblySections[startAddress] = instructions
				instructions = []avrassembler.Instruction{}
				startAddress = startAddress + (chunkLine * 2)
				chunkLine = 0

				// Implementing strings only, more data later
				data := []byte(m.Args)
				entry := avrassembler.DataBlob{
					Data:    data,
					Address: startAddress + (chunkLine * 2),
				}
				avrassembler.DbSections = append(avrassembler.DbSections, entry)
				startAddress += uint16((len(data) % 2) + len(data))
			}

			if m.Operation == "macro" {
				if inMacroDef != "" {
					fmt.Printf("[E] Cannot define macro inside another macro\n")
					os.Exit(1)
				}
				inMacroDef = m.Args
				avrassembler.RawAssemblySections[startAddress] = instructions
				instructions = []avrassembler.Instruction{}
			}

			if m.Operation == "endmacro" {
				if inMacroDef == "" {
					fmt.Printf("[E] No macro to complete\n")
					os.Exit(1)
				}
				avrassembler.RawMacroSections[inMacroDef] = instructions
				instructions = []avrassembler.Instruction{}
				inMacroDef = ""
			}

			if m.Operation == "import" {
				if inMacroDef != "" {
					fmt.Printf("[E] Cannot import inside macro definition\n")
					os.Exit(1)
				}
				inMacroDef = m.Args
				avrassembler.RawAssemblySections[startAddress] = instructions
				instructions = []avrassembler.Instruction{}
			}

			if m.Operation == "invokeMacro" {
				macroExpansion := avrassembler.RawMacroSections[m.Args]
				for _, instr := range macroExpansion {
					instr.Address = int(startAddress + chunkLine)
					instructions = append(instructions, instr)
					chunkLine++
				}
				continue
			}
		}

		// if white space, comment, or meta skip instruction logic
		if instruction.Mnemonic == "" {
			continue
		}
		instructions = append(instructions, instruction)
		if inMacroDef == "" {
			chunkLine++
		}
	}

	if inMacroDef != "" {
		fmt.Printf("[E] Macro definition was never closed...\n")
		os.Exit(1)
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
			operands := []string{}
			for _, o := range instructionSection[i].Operands {
				operands = append(operands, o.Value)
			}

			ops, err := encodingFunc(operands, instructionSection[i].Address)
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
	for _, dataBlob := range avrassembler.DbSections {
		dataBlobString := []string{hex.EncodeToString(dataBlob.Data)}
		fileContent, err := avrassembler.ToIntelHex(dataBlobString, int(dataBlob.Address))
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
