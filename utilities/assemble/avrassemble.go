package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	avrassembler "avrassembler"

	simplelog "github.com/ReidRise/simplelogger"
)

// CmdArgs holds input and output file names
type cmdArgs struct {
	InputFile  string
	OutputFile string
	LogLevel   string
}

var logLevelMap = map[string]simplelog.Level{
	"error": simplelog.ErrorLevel,
	"warn":  simplelog.WarnLevel,
	"info":  simplelog.InfoLevel,
	"debug": simplelog.DebugLevel,
	"trace": simplelog.TraceLevel,
}

// ParseCmdArgs parses command-line arguments for input and output files
func parseCmdArgs() (*cmdArgs, error) {
	input := flag.String("i", "", "Input assembly file (.S)")
	output := flag.String("o", "output.hex", "Output binary file (.hex)")
	loglevel := flag.String("l", "info", "Log level for assembler")

	flag.Parse()

	if *input == "" {
		return nil, fmt.Errorf("input file must be specified with -i")
	}

	// Check if input file exists
	if _, err := os.Stat(*input); os.IsNotExist(err) {
		return nil, fmt.Errorf("input file does not exist: %s", *input)
	}

	return &cmdArgs{
		InputFile:  *input,
		OutputFile: *output,
		LogLevel:   *loglevel,
	}, nil
}

func main() {
	args, err := parseCmdArgs()
	if err != nil {
		simplelog.Error(err.Error())
		os.Exit(1)
	}

	level, ok := logLevelMap[args.LogLevel]
	if !ok {
		simplelog.Warn(fmt.Sprintf("log level %s does not exist defaulting to info", args.LogLevel))
		level = simplelog.InfoLevel
	}

	avrassembler.SetLogLevel(level)

	avrassembler.ParseFile(args.InputFile, 0x0000)
	if err != nil {
		slog.Error(err.Error())
		avrassembler.DumpLabelMap()
		os.Exit(1)
	}

	err = avrassembler.WriteToFile(args.OutputFile)
	if err != nil {
		slog.Error(err.Error())
		avrassembler.DumpLabelMap()
		os.Exit(1)
	}
}
