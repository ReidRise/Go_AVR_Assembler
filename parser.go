package avrassembler

import (
	"fmt"
	"strconv"
	"strings"
)

type Instruction struct {
	Mnemonic string
	Operands []string
	Line     int // for error reporting
}

func ParseLine(line string, lineNumber uint16) (Instruction, string, error) {
	// Remove comments and trim whitespace
	clean := strings.Split(line, ";")[0]
	label_arr := strings.Split(clean, ":")
	var label = ""
	var parts = []string{}
	label_present := strings.Contains(clean, ":")
	if label_present {
		label = label_arr[0]
		parts = strings.Fields(label_arr[1])
	} else {
		parts = strings.Fields(clean)
	}
	if len(parts) == 0 {
		return Instruction{}, label, nil // empty line
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
		Line:     int(lineNumber),
	}, label, nil
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
	"LDI":   parseRegImm,
	"IN":    parseIOpsIn,
	"OUT":   parseIOpsOut,
	"CPI":   parseRegImm,
	"BRBC":  pasrseBranchSreg,
	"BRNE":  pasrseBranchStaticSreg,
	"BREQ":  pasrseBranchStaticSreg,
	"BRBS":  pasrseBranchSreg,
	"RJMP":  parseRelBranch,
	"RCALL": parseRelBranch,
	"RET":   parseConst,
	"NOP":   parseConst,
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

// Arg Parser

func parseConst(args []string, line_addr int) (ops [2]uint16, err error) {
	return [2]uint16{0, 0}, nil
}

func pasrseBranchStaticSreg(args []string, line_addr int) (ops [2]uint16, err error) {

	label_addr := int(LabelMap[args[0]])
	rel_addr := label_addr - line_addr - 1
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

	label_addr := int(LabelMap[args[1]])
	rel_addr := label_addr - line_addr
	if rel_addr > 2047 || rel_addr < -2048 {
		return [2]uint16{0, 0}, fmt.Errorf("relative address [%d] is not in range of +/- 2k", rel_addr)
	}
	ops[1] = uint16(rel_addr)

	return ops, nil
}

func parseRelBranch(args []string, line_addr int) (ops [2]uint16, err error) {
	label_addr, ok := LabelMap[args[0]]
	if !ok {
		return [2]uint16{0, 0}, fmt.Errorf("no label '%s'", args[0])
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
