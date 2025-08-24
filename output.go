package avrassembler

import (
	"encoding/hex"
	"fmt"
	"os"
	"slices"

	simplelog "github.com/ReidRise/simplelogger"
)

func toIntelHex(compiledAssembly []string, startingAddress int) (string, error) {
	intel_hex := ""
	intel_header := ""
	intel_line := ""
	intel_length := startingAddress
	for i := 0; i < len(compiledAssembly); i++ {
		if i != 0 && i%8 == 0 || i == (len(compiledAssembly)-1) {
			intel_line = intel_line + compiledAssembly[i]
			hex_len := fmt.Sprintf("%02x", len(intel_line)/2)
			hex_addr := fmt.Sprintf("%04x", intel_length)
			intel_length += len(intel_line) / 2
			intel_header = ":" + hex_len + hex_addr + "00"
			intel_checksum, err := intelHexChecksum(intel_header[1:] + intel_line)
			if err != nil {
				return "", fmt.Errorf("failed to compute intel checksum for %s", intel_line)
			}
			intel_hex += intel_header + intel_line + intel_checksum + "\n"
			intel_line = ""
		} else {
			intel_line = intel_line + compiledAssembly[i]
		}
	}
	//intel_hex += ":00000001FF"
	return intel_hex, nil
}

func intelHexChecksum(line string) (res string, err error) {
	decodedByteArray, err := hex.DecodeString(line)
	if err != nil {
		return "", err
	}
	checksum := uint64(0)
	for i := 0; i < len(decodedByteArray); i++ {
		checksum = checksum + uint64(decodedByteArray[i])
	}
	checksum = checksum & 0xff
	checksum = checksum ^ 0xff
	checksum = (checksum + 1) & 0xff
	return fmt.Sprintf("%02x", checksum), nil
}

func WriteToFile(fn string) (err error) {
	fileOut := ""
	// Parse Operands with context of all labels
	for addr, instructionSection := range RawAssemblySections {
		compiledAssembly := []string{}
		// Split into two loops to write from low->high addr
		for i := 0; i < len(instructionSection); i++ {
			encodingFunc, ok := InstructionParse[instructionSection[i].Mnemonic]
			if !ok {
				return fmt.Errorf("parsing function not found for %s not found on line %d", instructionSection[i].Mnemonic, instructionSection[i].Line)
			}
			operands := []string{}
			for _, o := range instructionSection[i].Operands {
				operands = append(operands, o.Value)
			}

			ops, err := encodingFunc(operands, instructionSection[i].Address)
			if err != nil {
				return fmt.Errorf("%s, Found on line %d", err, instructionSection[i].Line)
			}

			ins, ok := InstructionSet[instructionSection[i].Mnemonic]
			if !ok {
				return fmt.Errorf("encoding function not found for %s on line %d", instructionSection[i].Mnemonic, instructionSection[i].Line)
			}

			enc := ins.Encode(ins.ByteCode, ops[0], ops[1])

			le_enc := ((enc[0] >> 8) & 0x00ff) | ((enc[0] << 8) & 0xff00)
			hex := fmt.Sprintf("%x", le_enc)
			hex = fmt.Sprintf("%04s", hex)
			compiledAssembly = append(compiledAssembly, hex)
			simplelog.Debug(fmt.Sprintf("%6s %04s", instructionSection[i].Mnemonic, hex))

			// Extra handling for 32bit instructions
			if slices.Contains(LongInstructions, instructionSection[i].Mnemonic) {
				ins, ok := InstructionSet["_"+instructionSection[i].Mnemonic]
				if !ok {
					return fmt.Errorf("second encoding function not found for _%s", instructionSection[i].Mnemonic)
				}
				enc := ins.Encode(ins.ByteCode, ops[0], ops[1])
				le_enc := ((enc[0] >> 8) & 0x00ff) | ((enc[0] << 8) & 0xff00)
				hex := fmt.Sprintf("%x", le_enc)
				hex = fmt.Sprintf("%04s", hex) // Pad small strings with 0s
				compiledAssembly = append(compiledAssembly, hex)
			}
		}

		fileContent, err := toIntelHex(compiledAssembly, int(addr))
		if err != nil {
			return err
		}
		fileOut += fileContent
	}
	for _, dataBlob := range DbSections {
		dataBlobString := []string{hex.EncodeToString(dataBlob.Data)}
		fileContent, err := toIntelHex(dataBlobString, int(dataBlob.Address))
		if err != nil {
			return err
		}
		fileOut += fileContent
	}
	fileOut += ":00000001FF"
	simplelog.Debug("\n" + fileOut)
	os.Remove(fn)
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	l, err := f.WriteString(fileOut)
	if err != nil {
		return err
	}
	simplelog.Info(fmt.Sprintf("%d bytes written successfully", l))
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
