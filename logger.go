package avrassembler

import simplelog "github.com/ReidRise/simplelogger"

func SetLogLevel(level simplelog.Level) {
	simplelog.LogLevel = level
}
