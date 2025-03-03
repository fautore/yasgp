# YASGP - Yet Another Silly Go Proxy

YASGP is, as it's name implies, a proxy written in pure go

This project is meant to be a personal use, simple proxy to help deal with that _sweet sweet_, _awesome_ and _easily configurable_ WSL2 LAN networking via a simple approach: ignoring it entirely.

# Installation

Build from source

```shell
git clone git@github.com:fautore/yasgp.git
cd yasgp
go build .
```

...or via `go install` directly

```shell
go install github.com/fautore/yasgp@latest
```

# Usage

1. Write your configuration file to `config.yasgp`, currently yasgp looks for `config.yasgp` in the project root, aka your cwd.
2. Run
3. Enjoy your proxy!

# Configuration

yasgp uses a configuration file separated in lines, each line is a rule; below an example:

```yasgp
http://10.9.8.24:3118 to http://localhost:3118
http://10.9.8.24:3018 to http://localhost:3018
http://10.9.8.24:3020 to http://localhost:3020
```

Currently, it is required to restart yasgp when you update your configuration.

# Contribution

Contributions welcome! Feel free to open an Issue or a PR.
