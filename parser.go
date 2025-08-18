package avrassembler

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type PointerRegister byte

const (
	X PointerRegister = iota
	Y
	Z
)

type Instruction struct {
	Mnemonic string
	Operands []Token
	Address  int // Tracking jumps and branches
	Line     int // For error reporting
}

type Meta struct {
	Operation  string
	Args       string
	NewSection bool
}

func isMacro(macro string) (meta Meta, exists bool) {
	_, ok := RawMacroSections[macro]
	if ok {
		meta.Operation = "invokeMacro"
		meta.Args = macro
	}
	return meta, ok
}

func ParseFile(fn string, startAddress uint16) (err error, handoverAddress uint16) {
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	instructions := []Instruction{}
	// Collect Instuctions and Labels
	inMacroDef := ""
	// Line in file
	codeLine := uint16(0)
	// Line in raw assembly section
	chunkLine := uint16(0)

	fmt.Printf("Entering File %s at starting address 0x%04x\n", fn, startAddress/2)
	for scanner.Scan() {
		codeLine++
		instruction, meta, err := parseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("[E] error in file %s on line %d, %s ", fn, codeLine, err), 0
		}
		instruction.Address = int(chunkLine + (startAddress / 2))
		instruction.Line = int(codeLine)

		for _, m := range meta {
			if m.Operation == "label" {
				if inMacroDef != "" {
					return fmt.Errorf("[E] labels cannot be created in macros"), 0
				}
				//fmt.Printf("%s added to labelmap at %d\n", m.Args, line+(startAddress/2))
				LabelMap[m.Args] = chunkLine + (startAddress / 2)
			}

			if m.Operation == "org" {
				if inMacroDef != "" {
					return fmt.Errorf("[E] cannot define origin inside macros"), 0
				}
				RawAssemblySections[startAddress] = instructions
				instructions = []Instruction{}
				startAddress, err = parseImmidiateUints(m.Args)
				chunkLine = 0
				if err != nil {
					return fmt.Errorf("[E] error parsing address %s, %s ", m.Args, err), 0
				}
				if (startAddress % 2) != 0 {
					return fmt.Errorf("[W] address %s is not 16 bit aligned ", m.Args), 0
				}
			}

			if m.Operation == "db" {
				if inMacroDef != "" {
					return fmt.Errorf("[E] cannot define data blob in macro"), 0
				}
				RawAssemblySections[startAddress] = instructions
				instructions = []Instruction{}
				startAddress = startAddress + (chunkLine * 2)
				chunkLine = 0

				// Implementing strings only, more data later
				data := []byte(m.Args)
				entry := DataBlob{
					Data:    data,
					Address: startAddress + (chunkLine * 2),
				}
				DbSections = append(DbSections, entry)
				startAddress += uint16((len(data) % 2) + len(data))
			}

			if m.Operation == "macro" {
				if inMacroDef != "" {
					return fmt.Errorf("[E] Cannot define macro inside another macro"), 0
				}
				inMacroDef = m.Args
				RawAssemblySections[startAddress] = instructions
				instructions = []Instruction{}
			}

			if m.Operation == "endmacro" {
				if inMacroDef == "" {
					return fmt.Errorf("[E] No macro to complete"), 0
				}
				RawMacroSections[inMacroDef] = instructions
				instructions = []Instruction{}
				inMacroDef = ""
			}

			if m.Operation == "import" {
				if inMacroDef != "" {
					return fmt.Errorf("[E] cannot import inside macro definition"), 0
				}
				importFileName := m.Args
				RawAssemblySections[startAddress] = instructions
				instructions = []Instruction{}
				err, startAddress = ParseFile(importFileName, uint16(startAddress+(chunkLine*2)))
				chunkLine = 0
				if err != nil {
					return err, 0
				}
			}

			if m.Operation == "invokeMacro" {
				macroExpansion := RawMacroSections[m.Args]
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
		fmt.Printf("Parsing Instruction %s in file %s at line %d at address 0x%04x\n", instruction.Mnemonic, fn, instruction.Line, instruction.Address)
		if inMacroDef == "" {
			chunkLine++
		}
	}

	if inMacroDef != "" {
		return fmt.Errorf("[E] macro definition was never close"), 0

	}

	RawAssemblySections[startAddress] = instructions
	return nil, uint16(startAddress + (chunkLine * 2))
}

func parseMeta(tokens []Token) (meta []Meta, parsedTokens int, err error) {
	for i := 0; i < len(tokens); i++ {
		switch tokens[i].Type {
		case "Label":
			parsedTokens++
			meta = append(meta, Meta{})
			meta[i].Operation = "label"
			meta[i].Args = tokens[i].Value[:len(tokens[i].Value)-1]
		case "Operand":
			macro, exists := isMacro(tokens[i].Value)
			if exists {
				parsedTokens++
				meta = append(meta, macro)
			}
		case "MetaTag":
			parsedTokens++
			meta = append(meta, Meta{})
			switch tokens[i].Value {
			case ".db": // Use a offset for each db and then put it at end of code (maybe allow to be placed between orgs?)
				parsedTokens++
				meta[i].Operation = "db"
				if len(tokens) < i+1 {
					return meta, 0, fmt.Errorf("no data provided")
				}
				meta[i].Args = tokens[i+1].Value
				i++
			case ".org": // Set starting address for code after it
				parsedTokens++
				meta[i].Operation = "org"
				if len(tokens) < i+1 {
					return meta, 0, fmt.Errorf("no origin provided")
				}
				if tokens[i+1].DataType != "Integer" {
					return meta, 0, fmt.Errorf("origin provided is not integer")
				}
				meta[i].Args = tokens[i+1].Value
				i++
			case ".macro": // Setup a macro to be inserted into code
				parsedTokens++
				if len(tokens) <= i+1 {
					return meta, 0, fmt.Errorf("no macro name provided")
				}
				meta[i].Operation = "macro"
				meta[i].Args = tokens[i+1].Value
				i++
			case ".endmacro": // End a macro
				meta[i].Operation = "endmacro"
			case ".import":
				parsedTokens++
				if len(tokens) <= i+1 {
					return meta, 0, fmt.Errorf("no macro name provided")
				}
				meta[i].Operation = "import"
				meta[i].Args = tokens[i+1].Value
				i++
			}
		}
	}
	return meta, parsedTokens, nil
}

type Token struct {
	Type     string
	Value    string
	DataType string // Add data type
	Line     int
	Column   int
}

func tokenizeLine(code string) (tokens []Token, err error) {
	tokens = []Token{}
	line := 0
	for i := 0; i < len(code); i++ {
		r := rune(code[i])
		if unicode.IsSpace(r) {
			continue
		}

		if r == ';' {
			// Comment detected
			// Hacky workaround
			i = len(code)
		}

		if r == '.' && len(tokens) == 0 {
			buf := ""
			for ; i < len(code) && !unicode.IsSpace(rune(code[i])); i++ {
				buf += string(code[i])
			}
			tokens = append(tokens, Token{Type: "MetaTag", Value: buf, DataType: "String", Line: line, Column: i})
		}

		if unicode.IsLetter(r) {
			buf := ""
			for ; i < len(code) && (!unicode.IsSpace(rune(code[i]))) && (code[i] != ','); i++ {
				buf += string(code[i])
			}
			if buf[len(buf)-1] != ':' {
				tokens = append(tokens, Token{Type: "Operand", Value: buf, DataType: "String", Line: line, Column: i})
			} else {
				tokens = append(tokens, Token{Type: "Label", Value: buf, DataType: "String", Line: line, Column: i})
			}
		}

		if r == '"' {
			// Handle string literals
			buf := ""
			i++
			for ; i < len(code) && code[i] != '"'; i++ {
				buf += string(code[i])
			}

			if code[len(code)-1] != '"' {
				return tokens, fmt.Errorf("found string without matching \" on line %d col %d", line, i)
			}

			if i < len(code) && code[i] == '"' {
				tokens = append(tokens, Token{Type: "StringLiteral", Value: buf, DataType: "String", Line: line, Column: i})
				i++ // Advance past the closing quote
			}
		}

		if r == '0' {
			buf := ""
			tokenType := ""
			if len(code) <= i+1 && unicode.IsLetter(rune(code[i+1])) {
				return tokens, fmt.Errorf("incomplete number on line %d col %d", line, i)
			} else if code[i+1] == 'x' {
				tokenType = "Hexidecimal"
				i = i + 2
				buf += "0x"
				for ; i < len(code); i++ {
					if unicode.IsSpace(rune(code[i])) || rune(code[i]) == ',' {
						break
					}
					if !unicode.IsDigit(rune(code[i])) && !strings.ContainsAny(strings.ToUpper(string(code[i])), "A | B | C | D | E | F") {
						return tokens, fmt.Errorf("non-hex char [%s] on line %d col %d", string(code[i]), line, i)
					}
					buf += string(code[i])
				}
			} else if code[i+1] == 'b' {
				tokenType = "Binary"
				i = i + 2
				buf += "0b"
				for ; i < len(code) && rune(code[i]) != ','; i++ {
					if unicode.IsSpace(rune(code[i])) {
						break
					}
					if !strings.ContainsAny(strings.ToUpper(string(code[i])), "1 | 0") {
						return tokens, fmt.Errorf("non-binary char on line %d col %d", line, i)
					}
					buf += string(code[i])
				}
			} else {
				tokenType = "Decimal"
				buf := ""
				for ; i < len(code) && unicode.IsDigit(rune(code[i])); i++ {
					buf += string(code[i])
				}
			}
			tokens = append(tokens, Token{Type: tokenType, Value: buf, DataType: "Integer", Line: line, Column: i})

		} else if unicode.IsDigit(rune(r)) {
			// Integer Detection:  Very basic - needs refinement
			buf := ""
			for ; i < len(code) && unicode.IsDigit(rune(code[i])); i++ {
				buf += string(code[i])
			}
			tokens = append(tokens, Token{Type: "Decimal", Value: buf, DataType: "Integer", Line: line, Column: i})
		}
	}
	return tokens, nil
}

func parseLine(line string) (Instruction, []Meta, error) {
	// Remove comments and trim whitespace
	tokens, err := tokenizeLine(line)
	if err != nil {
		return Instruction{}, []Meta{}, err
	}
	if len(tokens) == 0 {
		return Instruction{}, []Meta{}, nil
	}

	meta, parsedTokens, err := parseMeta(tokens)
	if err != nil {
		return Instruction{}, []Meta{}, err
	}

	instructionTokens := tokens[parsedTokens:]
	// Pipe this back or just dup the work?
	// One of them is less bad
	if len(instructionTokens) == 0 {
		return Instruction{}, meta, nil // empty line
	}

	mnemonic := strings.ToUpper(instructionTokens[0].Value)
	operands := []Token{}
	if len(instructionTokens) > 1 {
		operands = instructionTokens[1:]
	}

	return Instruction{
		Mnemonic: mnemonic,
		Operands: operands,
	}, meta, nil
}

type ParserFunc func(args []string, line_addr int) ([2]uint16, error)

var InstructionParse = map[string]ParserFunc{
	// Arithmetic and Logic Instructions
	"ADC": parseTwoRegs,
	"ADD": parseTwoRegs,
	//ADIW
	"COM": parseOneReg,
	"DEC": parseOneReg,
	"SUB": parseTwoRegs,
	//SUBI
	"SBC":   parseTwoRegs,
	"SBIS":  parseSkipBit,
	"LDI":   parseRegImm,
	"IN":    parseIOpsIn,
	"OUT":   parseIOpsOut,
	"CPI":   parseRegImm,
	"POP":   parseOneReg,
	"PUSH":  parseOneReg,
	"BRBC":  pasrseBranchSreg,
	"BRNE":  pasrseBranchStaticSreg,
	"BREQ":  pasrseBranchStaticSreg,
	"BRBS":  pasrseBranchSreg,
	"RJMP":  parseRelBranch,
	"RCALL": parseRelBranch,
	"RET":   parseConst,
	"LPM":   parseLPM,
	"ELPM":  parseELPM,
	"NOP":   parseConst,
	"TST":   parseTST,
}

// Helper Functions

func parsePointerRegisters(reg_str string) (reg_uint uint16, ok bool, err error) {
	reg_parts := strings.Split(reg_str, "(")
	reg_letter := reg_parts[0]
	switch strings.ToUpper(reg_letter) {
	case "X":
		reg_uint = uint16(26)
	case "Y":
		reg_uint = uint16(28)
	case "Z":
		reg_uint = uint16(30)
	default:
		return 0, ok, nil
	}
	if len(reg_parts) > 1 {
		if reg_parts[1] == "HIGH)" {
			reg_uint += 1
		} else if reg_parts[1] == "LOW)" {
		} else {
			return 0, false, fmt.Errorf("unknown suffix (%s", reg_parts[1])
		}
	}
	return reg_uint, true, nil
}

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
		labelParsed := strings.Split(num, "(")
		imm, err := getLabelAddress(labelParsed[0])
		if err != nil {
			return 0, err
		} else {
			if len(labelParsed) > 1 {
				if strings.Split(labelParsed[1], ")")[0] == "HIGH" {
					imm &= 0xff00
				} else if strings.Split(labelParsed[1], ")")[0] == "LOW" {
					imm &= 0x00ff
				} else {
					return 0, fmt.Errorf("unknown suffix (%s", labelParsed[1])
				}
				return uint16(imm), nil
			}
		}
		return 0, fmt.Errorf(" unable to parse [%s] indwto uint", num)
	}
}

func parseRegister5bits(reg_str string) (reg_uint uint16, err error) {
	reg_uint, ok, err := parsePointerRegisters(reg_str)
	if err != nil {
		return 0, err
	} else if ok {
		return uint16(reg_uint), nil
	}
	if strings.ToUpper(reg_str[0:1]) != "R" {
		return 0, fmt.Errorf(" argument [%s] is not regiter rXX", reg_str)
	}
	reg_num, err := strconv.ParseUint(reg_str[1:], 10, 16)
	if err != nil {
		return 0, err
	}
	if reg_num > 31 {
		return 0, fmt.Errorf(" register [%s] does not exist", reg_str)
	}
	return uint16(reg_num), nil
}

func parseRegister4bits(reg_str string) (reg_uint uint16, err error) {
	if strings.ToUpper(reg_str[0:1]) != "R" {
		return 0, fmt.Errorf(" argument [%s] is not regiter rXX", reg_str)
	}
	reg_num, err := strconv.ParseUint(reg_str[1:], 10, 16)
	if err != nil {
		return 0, err
	}
	if reg_num > 31 || reg_num < 16 {
		return 0, fmt.Errorf(" register [%s] does not exist", reg_str)
	}
	return uint16(reg_num), nil
}

func parsePointerRegister(reg_str string) (reg PointerRegister, post_inc bool, err error) {
	reg, post_inc, err = 0, false, nil

	switch strings.ToUpper(reg_str[0:1]) {
	case "X":
		reg = X
	case "Y":
		reg = Y
	case "Z":
		reg = Z
	default:
		err = fmt.Errorf(" argument [%s] is not X, Y or Z", reg_str)
	}

	if strings.Contains(reg_str, "+") {
		post_inc = true
	}

	return
}

// Arg Parser

func getLabelAddress(label string) (addr uint16, err error) {
	addr, ok := LabelMap[label]
	if !ok {
		return 0, fmt.Errorf("label [%s] not found", label)
	}
	return addr, nil
}

func parseConst(args []string, line_addr int) (ops [2]uint16, err error) {
	return [2]uint16{0, 0}, nil
}

func parseSkipBit(args []string, line_addr int) (ops [2]uint16, err error) {
	ops[0], err = parseImmidiateUints(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[0] > 31 {
		return [2]uint16{0, 0}, fmt.Errorf("uint value [%d] is not a valid flag [0-31]", ops[0])
	}

	ops[0], err = parseImmidiateUints(args[1])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[0] > 7 {
		return [2]uint16{0, 0}, fmt.Errorf("uint value [%d] is not a valid flag [0-7]", ops[0])
	}
	return ops, nil
}

func pasrseBranchStaticSreg(args []string, line_addr int) (ops [2]uint16, err error) {
	label_addr, err := getLabelAddress(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	rel_addr := int(label_addr) - line_addr - 1
	if rel_addr > 2047 || rel_addr < -2048 {
		return [2]uint16{0, 0}, fmt.Errorf("relative address [%d] is not in range of +/- 2k", rel_addr)
	}

	ops[0] = uint16(0)
	ops[1] = uint16(rel_addr)

	return ops, nil
}

func pasrseBranchSreg(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseImmidiateUints(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[0] > 7 {
		return [2]uint16{0, 0}, fmt.Errorf("uint value [%d] is not a valid flag [0-7]", ops[0])
	}

	label_addr, err := getLabelAddress(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	rel_addr := int(label_addr) - line_addr - 1
	if rel_addr > 2047 || rel_addr < -2048 {
		return [2]uint16{0, 0}, fmt.Errorf("relative address [%d] is not in range of +/- 2k", rel_addr)
	}
	ops[1] = uint16(rel_addr)

	return ops, nil
}

func parseRelBranch(args []string, line_addr int) (ops [2]uint16, err error) {
	label_addr, err := getLabelAddress(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	rel_addr := int(label_addr) - line_addr - 1
	if rel_addr > 2047 || rel_addr < -2048 {
		return [2]uint16{0, 0}, fmt.Errorf("relative address [%d] is not in range of +/- 2k", rel_addr)
	}
	ops[0] = uint16(rel_addr) & 0x0fff
	return ops, nil
}

func parseIOpsIn(args []string, line_addr int) (ops [2]uint16, err error) {
	ops[0], err = parseRegister5bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	ops[1], err = parseImmidiateUints(args[1])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[1] > 63 {
		return [2]uint16{0, 0}, fmt.Errorf(" invalid io space [%d] > 63", ops[1])
	}
	return ops, nil
}

func parseIOpsOut(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseImmidiateUints(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[0] > 63 {
		return [2]uint16{0, 0}, fmt.Errorf(" invalid io space [%d] > 63", ops[1])
	}

	ops[1], err = parseRegister5bits(args[1])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	return ops, nil
}

func parseOneReg(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseRegister5bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	ops[1] = 0
	return ops, nil
}

func parseTwoRegs(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseRegister5bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	ops[1], err = parseRegister5bits(args[1])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	return ops, nil
}

func parseRegImm(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseRegister4bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	if ops[0] < 16 || ops[0] > 31 {
		return [2]uint16{0, 0}, fmt.Errorf(" register r%d is not 16 ≤ Rd ≤ 31", ops[0])
	}
	ops[0] = ops[0] - 16

	ops[1], err = parseImmidiateUints(args[1])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	return ops, nil
}

func parseLPM(args []string, line_addr int) (ops [2]uint16, err error) {
	// no arguments provided, this should be in zero-operand form
	if len(args) == 0 {
		ops, err = parseConst(args, line_addr)
		// each encoder function takes 3 arguments, but I need 5
		// pieces of information for this, so we're gonna encode
		// it into the 3rd argument, ops[1] (aka zqi)
		// set z bit to 1
		ops[1] = 0b100
		return
	}

	// arguments provided, assumed to be load/store form
	ops[0], err = parseRegister5bits(args[0])

	if err != nil {
		return
	}

	ptr_reg, post_inc, err := parsePointerRegister(args[1])

	if err != nil {
		return
	}

	// NOTE: I have no idea if this is actually true or not
	//			 I've only seen examples of Z in the docs, so it's
	//			 unclear whether X or Y are allowed here...
	if ptr_reg != Z {
		err = fmt.Errorf("pointer register value must be Z or Z+")
	}

	// set i bit to 1
	if post_inc {
		ops[1] = 0b001
	}

	return
}

// these parsing functions never receive information about the actual instruction
// ELPM needs its own call, and it should reference the LPM parser but set the q bit
// LPM/ELPM can share an encoder though
func parseELPM(args []string, line_addr int) (ops [2]uint16, err error) {
	ops, err = parseLPM(args, line_addr)

	// set the q bit (zqi)
	ops[1] |= 0b010

	return
}

func parseTST(args []string, line_addr int) (ops [2]uint16, err error) {

	ops[0], err = parseRegister5bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}

	ops[1], err = parseRegister5bits(args[0])
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	return ops, nil
}
