package avrassembler

type EncoderFunc func(bytecode uint16, rd uint16, rr uint16) [1]uint16

type InstructionDef struct {
	Operands int
	ByteCode uint16
	Encode   EncoderFunc
}

/*
NOTES TO SELF:

Some instructions have several forms with variable operands and operand types.
Below I've compiled a list of these instructions and the valid syntax patterns they
can appear in:

ELPM - Extended Load Program Memory
* ELPM								Zero oeprand, R0 implied
* ELPM Rd, Z					Load/store operation, 0 <= d <= 31
* ELPM Rd, Z+					Load/store operation, 0 <= d <= 31, Z post-increment

LPM - Load Program Memory
* LPM									Zero operand, R0 implied
* LPM Rd, Z						Load/store operation, 0 <= d <= 31
* LPM Rd, Z+					Load/store operation, 0 <= d <= 31, Z post-increment

SPM - Store Program Memory
* SPM									Zero operand
* SPM Z+							Zero operand, Z post-increment

*/

/*
NOTE: q = Extend program memory address with RAMPZ (0 = 0:Z, 1 = RAMPZ:Z)

LPM/ELPM zero-operand form, R0 implied
1 0 0 1 | 0 1 0 1 | 1 1 0 q | 1 0 0 0
                  ^ opcode  ^

LPM/ELPM load/store form, 0 <= d <= 31
1 0 0 1 | 0 0 0 d | d d d d | 0 1 q 0
                            ^ opcode  ^

LPM/ELPM load/store form, 0 <= d <= 31, Z post-increment
1 0 0 1 | 0 0 0 d | d d d d | 0 1 q 1
                            ^ opcode  ^
*/

var InstructionSet = map[string]InstructionDef{

	// Arithmetic and Logic Instructions
	"ADD":  {Operands: 2, ByteCode: 0b0000110000000000, Encode: EncodeTwoRegs},
	"ADC":  {Operands: 2, ByteCode: 0b0001110000000000, Encode: EncodeTwoRegs},
	"ADIW": {Operands: 2, ByteCode: 0b1001011000000000, Encode: EncodeWordImm},
	"SUB":  {Operands: 2, ByteCode: 0b0001100000000000, Encode: EncodeTwoRegs},
	"SUBI": {Operands: 2, ByteCode: 0b0101000000000000, Encode: EncodeRegImm},
	"SBC":  {Operands: 2, ByteCode: 0b0000100000000000, Encode: EncodeTwoRegs},
	"SBCI": {Operands: 2, ByteCode: 0b0100000000000000, Encode: EncodeRegImm},
	"SBIW": {Operands: 2, ByteCode: 0b1001011100000000, Encode: EncodeWordImm},
	"AND":  {Operands: 2, ByteCode: 0b0010000000000000, Encode: EncodeTwoRegs},
	"ANDI": {Operands: 2, ByteCode: 0b0111000000000000, Encode: EncodeRegImm},
	"OR":   {Operands: 2, ByteCode: 0b0010100000000000, Encode: EncodeTwoRegs},
	"ORI":  {Operands: 2, ByteCode: 0b0110000000000000, Encode: EncodeRegImm},
	"EOR":  {Operands: 2, ByteCode: 0b0010010000000000, Encode: EncodeTwoRegs},
	"COM":  {Operands: 1, ByteCode: 0b1001010000000000, Encode: EncodeReg},
	"NEG":  {Operands: 1, ByteCode: 0b1001010000000001, Encode: EncodeReg},
	"SBR":  {Operands: 2, ByteCode: 0b0110000000000000, Encode: EncodeRegImm},
	"CBR":  {Operands: 2, ByteCode: 0b0111000000000000, Encode: EncodeRegImm}, // Same as ANDI compliment K
	"INC":  {Operands: 1, ByteCode: 0b1001010000000011, Encode: EncodeReg},
	"DEC":  {Operands: 1, ByteCode: 0b1001010000001010, Encode: EncodeReg},
	"TST":  {Operands: 2, ByteCode: 0b0010000000000000, Encode: EncodeTwoRegs}, // Technically its Rd & Rd
	"CLR":  {Operands: 2, ByteCode: 0b0010010000000000, Encode: EncodeTwoRegs}, // Technically its Rd ^ Rd
	"SER":  {Operands: 1, ByteCode: 0b1110111100001111, Encode: EncodeRegGP},
	"MUL":  {Operands: 2, ByteCode: 0b1001110000000000, Encode: EncodeTwoRegs},
	"MULS": {Operands: 2, ByteCode: 0b0000001000000000, Encode: EncodeAdvMath},
	// "MULSU":
	// "FMUL":
	// "FMULS":
	// "FMULSU":
	// "DES":

	// Change of Flow Instructions
	"RJMP": {Operands: 1, ByteCode: 0b1100000000000000, Encode: EncodeRelBranch},
	// "IJMP":
	// "EIJMP":
	"RCALL": {Operands: 1, ByteCode: 0b1101000000000000, Encode: EncodeRelBranch},
	// "ICALL":
	// "EICALL":
	// "CALL":
	"RET":  {Operands: 0, ByteCode: 0b1001010100001000, Encode: EncodeConstant},
	"RETI": {Operands: 0, ByteCode: 0b1001010100011000, Encode: EncodeConstant},
	"CPSE": {Operands: 2, ByteCode: 0b0001000000000000, Encode: EncodeTwoRegs},
	"CP":   {Operands: 2, ByteCode: 0b0001010000000000, Encode: EncodeTwoRegs},
	"CPC":  {Operands: 2, ByteCode: 0b0000010000000000, Encode: EncodeTwoRegs},
	"CPI":  {Operands: 2, ByteCode: 0b0011000000000000, Encode: EncodeRegImm},
	"SBRC": {Operands: 2, ByteCode: 0b1111110000000000, Encode: EncodeSkipBit},
	"SBRS": {Operands: 2, ByteCode: 0b1111111000000000, Encode: EncodeSkipBit},
	"SBIC": {Operands: 2, ByteCode: 0b1001100100000000, Encode: EncodeSkipBitIO},
	"SBIS": {Operands: 2, ByteCode: 0b1001101100000000, Encode: EncodeSkipBitIO},
	"BRBS": {Operands: 2, ByteCode: 0b1111000000000000, Encode: EncodeBranchSreg},
	"BRBC": {Operands: 2, ByteCode: 0b1111010000000000, Encode: EncodeBranchSreg},
	"BREQ": {Operands: 2, ByteCode: 0b1111000000000001, Encode: EncodeBranchSreg}, // s = 0b001
	"BRNE": {Operands: 2, ByteCode: 0b1111010000000001, Encode: EncodeBranchSreg}, // s = 0b001
	"BRCS": {Operands: 2, ByteCode: 0b1111000000000000, Encode: EncodeBranchSreg}, // s = 0b000
	"BRCC": {Operands: 2, ByteCode: 0b1111010000000000, Encode: EncodeBranchSreg}, // s = 0b000
	"BRSH": {Operands: 2, ByteCode: 0b1111010000000000, Encode: EncodeBranchSreg}, // s = 0b000
	"BRLO": {Operands: 2, ByteCode: 0b1111000000000000, Encode: EncodeBranchSreg}, // s = 0b000
	"BRMI": {Operands: 2, ByteCode: 0b1111000000000010, Encode: EncodeBranchSreg}, // s = 0b010
	"BRPL": {Operands: 2, ByteCode: 0b1111010000000010, Encode: EncodeBranchSreg}, // s = 0b010
	"BRGE": {Operands: 2, ByteCode: 0b1111010000000100, Encode: EncodeBranchSreg}, // s = 0b100
	"BRLT": {Operands: 2, ByteCode: 0b1111000000000100, Encode: EncodeBranchSreg}, // s = 0b100
	"BRHS": {Operands: 2, ByteCode: 0b1111000000000101, Encode: EncodeBranchSreg}, // s = 0b101
	"BRHC": {Operands: 2, ByteCode: 0b1111010000000101, Encode: EncodeBranchSreg}, // s = 0b101
	"BRTS": {Operands: 2, ByteCode: 0b1111000000000110, Encode: EncodeBranchSreg}, // s = 0b110
	"BRTC": {Operands: 2, ByteCode: 0b1111010000000110, Encode: EncodeBranchSreg}, // s = 0b110
	"BRVS": {Operands: 2, ByteCode: 0b1111000000000011, Encode: EncodeBranchSreg}, // s = 0b011
	"BRVC": {Operands: 2, ByteCode: 0b1111010000000011, Encode: EncodeBranchSreg}, // s = 0b011
	"BRIE": {Operands: 2, ByteCode: 0b1111000000000111, Encode: EncodeBranchSreg}, // s = 0b111
	"BRID": {Operands: 2, ByteCode: 0b1111010000000111, Encode: EncodeBranchSreg}, // s = 0b111

	// Data Transfer Instructions
	"MOV":  {Operands: 2, ByteCode: 0b0010110000000000, Encode: EncodeTwoRegs},
	"MOVW": {Operands: 2, ByteCode: 0b0000000100000000, Encode: EncodeAdvMath},
	"LDI":  {Operands: 2, ByteCode: 0b1110000000000000, Encode: EncodeRegImm},
	// "LDS":  {Operands: 2, ByteCode: 0b1001000000000000, Encode: EncodeLoadMemory},
	// "LD":
	// "LDD":
	// "STS": {Operands: 2, ByteCode: 0b1001001000000000, Encode: EncodeLoadMemory},
	// "ST":
	// "STD":
	"LPM":  {Operands: 2, ByteCode: 0b1001_0000_0000_0000, Encode: EncodeLPM}, // zo-form: 1001_0101_110q_1000 (opcode nibble 3) | ls-form: 1001_000d_dddd_01q0 (opcode nibble 4)
	"ELPM": {Operands: 2, ByteCode: 0b1001_0000_0000_0000, Encode: EncodeLPM},
	// "SPM":
	"IN":   {Operands: 2, ByteCode: 0b1011000000000000, Encode: EncodeIOpsIn},
	"OUT":  {Operands: 2, ByteCode: 0b1011100000000000, Encode: EncodeIOpsOut},
	"PUSH": {Operands: 2, ByteCode: 0b1001001000001111, Encode: EncodeReg},
	"POP":  {Operands: 2, ByteCode: 0b1001000000001111, Encode: EncodeReg},
	// "XCH":
	// "LAS":
	// "LAC":
	// "LAT":

	// Bit and Bit-Test Instructions
	// "LSL":
	"LSR": {Operands: 1, ByteCode: 0b1001010000000110, Encode: EncodeShift},
	// "ROL":
	"ROR": {Operands: 1, ByteCode: 0b1001010000000111, Encode: EncodeShift},
	"ASR": {Operands: 1, ByteCode: 0b1001010000000101, Encode: EncodeShift},
	// "SWAP":
	// "SBI":
	// "CBI":
	// "BST":
	// "BLD":
	"BSET": {Operands: 1, ByteCode: 0b1001010000001000, Encode: EncodeSREGBitOp},
	"BCLR": {Operands: 1, ByteCode: 0b1001010010001000, Encode: EncodeSREGBitOp},
	// "SEC":
	// "CLC":
	// "SEN":
	// "CLN":
	// "SEZ":
	// "CLZ":
	// "CLI":
	// "SES":
	// "CLS":
	// "SEV":
	// "CLV":
	// "SET":
	// "CLT":
	// "SEH":
	// "CLH":

	// MCU Control Instructions
	"BREAK": {Operands: 0, ByteCode: 0b1001010110011000, Encode: EncodeConstant},
	"NOP":   {Operands: 0, ByteCode: 0b0000000000000000, Encode: EncodeConstant},
	"SLEEP": {Operands: 0, ByteCode: 0b1001010110001000, Encode: EncodeConstant},
	"WDR":   {Operands: 0, ByteCode: 0b1001010110101000, Encode: EncodeConstant},
}

func EncodeRelBranch(bytecode uint16, kk uint16, _ uint16) [1]uint16 {
	encoded := bytecode
	encoded |= kk
	return [1]uint16{encoded}
}

func EncodeLoadMemory(bytecode uint16, rd uint16, kk uint16) [2]uint16 {
	// Base opcode for LDS
	encoded := bytecode
	encoded |= ((rd & 0x1F) << 4)
	return [2]uint16{encoded, kk}
}

func EncodeBranchSreg(bytecode uint16, ss uint16, kk uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (kk & 0x7f) << 3
	encoded |= (ss & 0x07)
	return [1]uint16{encoded}
}

func EncodeSkipBitIO(bytecode uint16, aa uint16, bb uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (aa & 0x1f) << 3
	encoded |= (bb & 0x03)
	return [1]uint16{encoded}
}

func EncodeSkipBit(bytecode uint16, rr uint16, bb uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rr & 0x1f) << 4
	encoded |= (bb & 0x03)
	return [1]uint16{encoded}
}

func EncodeIOpsIn(bytecode uint16, rr uint16, aa uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rr & 0x1f) << 4
	encoded |= (aa & 0x30) << 5
	encoded |= (aa & 0x0f)
	return [1]uint16{encoded}
}

func EncodeIOpsOut(bytecode uint16, aa uint16, rr uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rr & 0x1f) << 4
	encoded |= (aa & 0x30) << 5
	encoded |= (aa & 0x0f)
	return [1]uint16{encoded}
}

func EncodeConstant(bytecode uint16, _a uint16, _b uint16) [1]uint16 {
	return [1]uint16{bytecode}
}

func EncodeAdvMath(bytecode uint16, rd uint16, rr uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x0f) << 4
	encoded |= (rr & 0x0f)
	return [1]uint16{encoded}
}

func EncodeRegGP(bytecode uint16, rd uint16, _ uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0xf) << 4
	return [1]uint16{encoded}
}

func EncodeReg(bytecode uint16, rd uint16, _ uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x1f) << 4
	return [1]uint16{encoded}
}

func EncodeTwoRegs(bytecode uint16, rd uint16, rr uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x1f) << 4
	encoded |= (rr & 0x10) << 5
	encoded |= (rr & 0x0f)
	return [1]uint16{encoded}
}

func EncodeRegImm(bytecode uint16, rd uint16, kk uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x0f) << 4
	encoded |= (kk & 0xf0) << 4
	encoded |= (kk & 0x0f)
	return [1]uint16{encoded}
}

func EncodeShift(bytecode uint16, rd uint16, _ uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x1f) << 4
	return [1]uint16{encoded}
}

func EncodeWordImm(bytecode uint16, rd uint16, kk uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (rd & 0x03) << 4
	encoded |= (kk & 0x30) << 2
	encoded |= (kk & 0x0f)
	return [1]uint16{encoded}
}

func EncodeSREGBitOp(bytecode uint16, s uint16, _ uint16) [1]uint16 {
	encoded := bytecode
	encoded |= (s & 0x07) << 4
	return [1]uint16{encoded}
}

// zero operand form (z = 0), load/store form: (z = 1)
// Z no post-increment (i = 0), Z post-increment (i = 1)
// LPM (q = 0), ELPM (q = 1)
// NOTE: rd does NOT tell whether the instruction is zo-form or ls-form
//
//	 Ex: lpm r0, Z	; explicit rd = r0
//			 lpm				; implied rd = r0
func EncodeLPM(bytecode uint16, rd uint16, zqi uint16) [1]uint16 {
	encoded := bytecode
	// check form
	if (zqi & 0b100) != 0 {
		// zero-operand form
		// current: 1001_0000_0000_0000
		// changed: 1001_0101_110q_1000

		// encode the opcode and q
		encoded |= 0b0000_0101_1100_1000
		encoded |= (zqi & 0b010) << 3 // q (q = 1 => ELPM)

		return [1]uint16{encoded}
	}

	// load-store form
	// current: 1001_0000_0000_000i
	// changed: 1001_000d_dddd_01qi

	// encode ddddd first
	encoded = EncodeReg(encoded, rd, zqi)[0]

	// encode q and i
	encoded |= 0b0100
	encoded |= (zqi & 0b010) // q (q = 1 => ELPM)
	encoded |= (zqi & 0x01)  // i (i = 1 => Z+)

	return [1]uint16{encoded}
}
