package avrassembler

import (
	"golang.org/x/exp/slog"
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
	slog.Debug("Label Map:\n")
	for key, value := range LabelMap {
		slog.Debug("\t%s @ 0x%04x\n", key, value)
	}
}
