package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	"gopkg.in/yaml.v2"

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

	config := loadConfig()
	_ = config // fixme

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

func loadConfig() (config map[string]map[string]string) {
	config = map[string]map[string]string{}
	config["default"] = map[string]string{}

	var _configFile string

	if *configFile != "" {
		_configFile = *configFile
	} else {
		user, _ := user.Current()
		home := user.HomeDir
		debugf("home: %s", home)

		for _, file := range []string{"_/etc/alertist.yaml", home + "/.alertist.yaml"} {
			debugf("exists? %s", file)
			if _, err := os.Stat(file); err == nil {
				_configFile = file
				break
			}
		}
	}
	debugf("configFile: %s", _configFile)

	if _configFile != "" {
		content, err := ioutil.ReadFile(_configFile)
		if err != nil {
			debugf("failed to read %s: %s", _configFile, err)
		} else {
			err := yaml.Unmarshal(content, &config)
			if err != nil {
				debugf("failed to parse as yaml: %s", err)
			}
		}
	}

	debugf("config: %s", config)
	return config
}
