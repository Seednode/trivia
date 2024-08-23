## About

A basic trivia webapp, for my personal use (and yours, if you want)!

Feature requests, code criticism, bug reports, general chit-chat, and unrelated angst accepted at `trivia@seedno.de`.

Static binary builds available [here](https://cdn.seedno.de/builds/trivia).

I only test the linux/amd64, linux/arm64, and windows/amd64 builds, the rest are all best-effort™.

Dockerfile available [here](https://raw.githubusercontent.com/Seednode/trivia/master/docker/Dockerfile).

An example instance with most features enabled can be found [here](https://trivia.seedno.de/).

## File format
The app expects newline-delimited text files in the following format (category is optional):
```
<question>|<answer>|[category]
<question>|<answer>|[category]
<question>|<answer>|[category]
[...]
```

For example:
```
What is the current year?|2024|History
How many inches are in a foot?|12|Measurement
Is mayonnaise an instrument?|No, Patrick, mayonnaise is not an instrument|Cartoons
```

You can specify as many input files as you'd like, by repeating the `-f|--question-file` flag.

## Exporting
If the `--export` flag is passed, an additional `/export` endpoint is registered.

The trivia database can be viewed by calling the `/export` endpoint.

The output will be in the following format:
```
Category: History
Question: What is the current year?
Answer: 2024
[...]
```

## Reloading
If the `--reload` flag is passed, an additional `/reload` endpoint is registered.

The trivia database can be live-reloaded from all files passed in the `-f|--question-file` flags by calling this endpoint.

Scheduled index rebuilds can be enabled via the `--reload-interval <duration>` flag, which accepts [time.Duration](https://pkg.go.dev/time#ParseDuration) strings.

## Usage output
```
Serves a basic trivia web frontend.

Usage:
  trivia [flags]

Flags:
  -b, --bind string              address to bind to (default "0.0.0.0")
      --exit-on-error            shut down webserver on error, instead of just printing the error
      --export                   allow exporting of trivia database
  -h, --help                     help for trivia
  -p, --port uint16              port to listen on (default 8080)
      --profile                  register net/http/pprof handlers
  -f, --question-file strings    path to file containing trivia questions (can be supplied multiple times)
      --reload                   allow live-reload of questions
      --reload-interval string   interval at which to rebuild question list (e.g. "5m" or "1h")
  -v, --verbose                  log requests to stdout
  -V, --version                  display version and exit
```

## Building the Docker image
From inside the cloned repository, build the image using the following command:

`REGISTRY=<registry url> LATEST=yes TAG=alpine ./build-docker.sh`