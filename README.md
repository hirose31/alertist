# alertist

A simple CLI utility to send a notification to Slack if a given command failed.

## Usage

```
cat << EOF | sudo tee /etc/alertist.yaml
default:
  slack:
    hook: 'https://hooks.slack.com/services/XXX/XXX'
    channel: '#watchme'
EOF
```

I use alertist to run commands in cron to send alerts if cron commands fail.

```
0 * * * * alertist -- /path/to/script --foo FOO --bar
```

## Installation

```
go build -o alertist cmd/alertist/main.go
```

## More features?

Try [alerty](https://github.com/sonots/alerty).
