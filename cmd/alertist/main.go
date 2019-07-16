package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"

	debug "github.com/visionmedia/go-debug"
)

var debugf = debug.Debug("alertist")

var (
	configFile = flag.String("c", "", "configuration file")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage:
  %s [OPTIONS] ARGS...
Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	fmt.Printf("%#v\n", flag.Args())
	fmt.Printf("%s\n", *configFile)

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing command to execute")
		flag.Usage()
		os.Exit(1)
	}

	stdout, stderr, code, err := execute(args)
	if err != nil {
		fmt.Println("ERR")
	}

	fmt.Printf("OUT:%s\nERR:%s\nCODE:%d\n", stdout, stderr, code)

}

func execute(args []string) (stdout string, stderr string, code int, err error) {
	debugf("execute args: %s", args)
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	var oer bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &oer
	err = cmd.Run()

	return out.String(), oer.String(), cmd.ProcessState.ExitCode(), err
}
