package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/hirose31/alertist"
	debug "github.com/visionmedia/go-debug"
)

var debugf = debug.Debug("alertist")

var (
	configFile = flag.String("c", "", "config file")
	target     = flag.String("t", "default", "target in config file")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Version: %s
Usage:
  %s [OPTIONS] COMMAND ARGS...
Options:
`,
			alertist.Version,
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	config := loadConfig()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing command to execute")
		flag.Usage()
		os.Exit(1)
	}

	stdout, stderr, code, err := execute(args)
	debugf("OUT:%s\nERR:%s\nCODE:%d\n", stdout, stderr, code)
	if err != nil {
		if targetConfig, ok := config[*target]; ok {
			notify(args, stdout, stderr, code, targetConfig)
		}
	}
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

func loadConfig() (config map[string]map[string]interface{}) {
	config = map[string]map[string]interface{}{}
	config["default"] = map[string]interface{}{}

	var _configFile string

	if *configFile != "" {
		_configFile = *configFile
	} else {
		user, _ := user.Current()
		home := user.HomeDir
		debugf("home: %s", home)

		for _, file := range []string{"/etc/alertist.yaml", home + "/.alertist.yaml"} {
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

func notify(args []string, stdout string, stderr string, code int, config map[string]interface{}) {
	debugf("notify config: %s", config)

	if slackConfig, ok := config["slack"]; ok {
		debugf("notify by slack")
		type Slack struct {
			Text      string `json:"text"`
			Username  string `json:"username"`
			IconEmoji string `json:"icon_emoji"`
			Channel   string `json:"channel"`
		}

		text := fmt.Sprintf(`
command: %s
stdout: %s
stderr: %s
code: %d
`,
			strings.Join(args, " "),
			stdout,
			stderr,
			code,
		)

		payload, _ := json.Marshal(Slack{
			"```" + text + "```",
			"alertist",
			":mega:",
			slackConfig.(map[interface{}]interface{})["channel"].(string),
		})
		debugf("payload: %s", payload)

		resp, _ := http.PostForm(
			slackConfig.(map[interface{}]interface{})["hook"].(string),
			url.Values{"payload": {string(payload)}},
		)
		debugf("status code: %d", resp.StatusCode)
	}
}
