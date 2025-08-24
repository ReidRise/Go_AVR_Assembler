package avrassembler

import (
	"fmt"

	simplelog "github.com/ReidRise/simplelogger"
)

type DataBlob struct {
	Data    []byte
	Address uint16
}

var LabelMap = map[string]uint16{}

var RawAssemblySections = map[uint16][]Instruction{}
var RawMacroSections = map[string][]Instruction{}

var DbSections = []DataBlob{}

func DumpLabelMap() {
	simplelog.Debug("Label Map:")
	for key, value := range LabelMap {
		simplelog.Debug(fmt.Sprintf("\t%s @ 0x%04x", key, value))
	}
}
