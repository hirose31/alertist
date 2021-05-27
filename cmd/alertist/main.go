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
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	debug "github.com/visionmedia/go-debug"
)

const version = "0.0.1"

var revision = "HEAD"
var debugf = debug.Debug("alertist")

var (
	showVersion = flag.Bool("version", false, "print version information")
	configFile  = flag.String("c", "", "config file")
	target      = flag.String("t", "default", "target in config file")
	retry       = flag.Int("r", 1, "number of retries")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Version: %s
Usage:
  %s [OPTIONS] COMMAND ARGS...
Options:
`,
			version,
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()
	if *showVersion {
		fmt.Printf(
			"alertist %s (rev: %s/%s)\n",
			version,
			revision,
			runtime.Version(),
		)
		os.Exit(1)
	}

	config := loadConfig()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "missing command to execute")
		flag.Usage()
		os.Exit(1)
	}

	var stdout, stderr string
	var code int
	var err error
	for *retry > 0 {
		stdout, stderr, code, err = execute(args)
		debugf("CODE:%d\n", code)
		fmt.Printf("STDOUT:\n%s\nSTDERR:\n%s\n", stdout, stderr)

		if err == nil {
			break
		}

		*retry--
		time.Sleep(7 * time.Second)
	}

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

		for _, file := range []string{home + "/.alertist.yaml", "/etc/alertist.yaml"} {
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
