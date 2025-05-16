package avrassembler

import (
	"encoding/hex"
	"fmt"
)

func ToIntelHex(compiledAssembly []string, startingAddress int) (string, error) {
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
		//fmt.Println("error decoding hex string: %s", err)
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
