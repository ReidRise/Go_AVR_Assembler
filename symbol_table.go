package avrassembler

import (
	"fmt"

	simplelog "github.com/ReidRise/simplelogger"
)

// Data to be loaded to a memory location
type DataBlob struct {
	Data    []byte
	Address uint16
}

// Format for laying out instructions in memory at address
type AssemblySection struct {
	Address  uint16
	Assembly []Instruction
}

// Instruction Sections
var RawAssemblySections = []AssemblySection{}
var RawMacroSections = map[string][]Instruction{}

// Labels in Memory
var LabelMap = map[string]uint16{}

// Data blobs (strings for now) in memory
var DbSections = []DataBlob{}

// Variable Symbols to uint mapping
var VariableMapping = map[string]uint16{}

func DumpLabelMap() {
	simplelog.Trace("Label Map:")
	for key, value := range LabelMap {
		simplelog.Trace(fmt.Sprintf("\t%s @ 0x%04x", key, value))
	}
}
