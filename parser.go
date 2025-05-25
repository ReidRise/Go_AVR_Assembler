package avrassembler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type PointerRegister byte

const (
	X PointerRegister = iota
	Y
	Z
)

type Instruction struct {
	Mnemonic string
	Operands []string
	Line     int // for error reporting
}

type Meta struct {
	Operation  string
	Args       string
	NewSection bool
}

func isMacro(macro string) (meta Meta) {
	_, ok := RawMacroSections[macro]
	if ok {
		meta.Operation = "invokeMacro"
		meta.Args = macro
	}
	return meta
}

func ParseQuotedString(input string) (string, error) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return "", errors.New("input too short to be quoted")
	}

	start := input[0]
	end := input[len(input)-1]

	println(input)

	if (start != '"' && start != '\'') || start != end {
		return "", errors.New("string must be enclosed in matching single or double quotes")
	}

	unquoted, err := strconv.Unquote(input)
	if err != nil {
		return "", err
	}

	return unquoted, nil
}

func parseLabel(line string) (label string, label_present bool) {
	label = ""
	label_arr := strings.Split(line, ":")
	label_present = strings.Contains(line, ":")
	if label_present {
		label = label_arr[0]
	}
	return label, label_present
}

func parseData(dataIn string) (dataOut string, err error) {
	data, err := ParseQuotedString(dataIn)
	if err == nil {
		return data, nil
	}
	println(err)
	num, err := parseImmidiateUints(dataIn)
	if err == nil {
		data = strconv.FormatUint(uint64(num), 16)
		return data, nil
	}
	return "", fmt.Errorf("unable to parse DataBlob")
}

func parseMeta(line string) (meta Meta, err error) {
	meta = Meta{}
	parts := strings.Split(line, " ")
	println(line)
	switch parts[0] {
	case ".db": // Use a offset for each db and then put it at end of code (maybe allow to be placed between orgs?)
		meta.Operation = "db"
		meta.Args, err = parseData(strings.Join(parts[1:], " "))
		if err != nil {
			return meta, err
		}
	case ".org": // Set starting address for code after it
		meta.Operation = "org"
		meta.Args = parts[1]
	case ".macro": // Setup a macro to be inserted into code
		if len(parts) == 1 {
			return meta, fmt.Errorf("no macro name provided")
		}
		meta.Operation = "macro"
		meta.Args = parts[1]
	case ".endmacro": // End a macro
		meta.Operation = "endmacro"
	default:
		label, present := parseLabel(line)
		if present {
			meta.Operation = "label"
			meta.Args = label
		} else {
			meta = isMacro(parts[0])
		}
	}
	return meta, nil
}

func ParseLine(line string) (Instruction, Meta, error) {
	// Remove comments and trim whitespace
	var parts = []string{}
	clean := strings.Split(line, ";")[0]

	meta, err := parseMeta(clean)
	if err != nil {
		fmt.Println("[E] Failed to parse metadata")
		return Instruction{}, meta, err // empty line
	}

	if meta.Operation != "" {
		if meta.Operation == "label" {
			parsed := strings.Split(clean, ":")
			parts = strings.Fields(parsed[1])
		} else {
			return Instruction{}, meta, nil
		}
	} else {
		parts = strings.Fields(clean)
	}

	// Pipe this back or just dup the work?
	// One of them is less bad
	if len(parts) == 0 {
		return Instruction{}, meta, nil // empty line
	}

	mnemonic := strings.ToUpper(parts[0])
	operands := []string{}
	if len(parts) > 1 {
		// Join everything after the mnemonic, split by ','
		ops := strings.Join(parts[1:], " ")
		operands = strings.Split(ops, ",")
		for i := range operands {
			operands[i] = strings.TrimSpace(operands[i])
		}
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

func parseRegister5bits(reg_str string) (reg_uint uint16, err error) {
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

	println(fmt.Sprintf("0x%04x 0x%04x", label_addr, line_addr))
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

	println(fmt.Sprintf("0x%04x 0x%04x", label_addr, line_addr))
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

	println(fmt.Sprintf("0x%04x 0x%04x", label_addr, line_addr))
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
