package avrassembler

type DataBlob struct {
	Data    []byte
	Address uint16
}

var LabelMap = map[string]uint16{}

var RawAssemblySections = map[uint16][]Instruction{}
var RawMacroSections = map[string][]Instruction{}

var DbSections = []DataBlob{}
