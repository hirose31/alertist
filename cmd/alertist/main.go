package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"

	debug "github.com/visionmedia/go-debug"
)

var debugf = debug.Debug("fixme")

var (
	configFile = flag.String("c", "/etc/alertist.yaml", "configuration file")
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

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "missing command to execute")
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	var oer bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &oer
	err := cmd.Run()
	if err != nil {
		fmt.Println("ERR")
		fmt.Println(err)
	}
	fmt.Printf("STDOUT: %s\n", out.String())
	fmt.Printf("STDERR: %s\n", oer.String())
	fmt.Printf("exit with %d\n", cmd.ProcessState.ExitCode())
}
