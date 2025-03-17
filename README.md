## About

A very basic trivia webapp.

Feature requests, code criticism, bug reports, general chit-chat, and unrelated angst accepted at `trivia@seedno.de`.

Static binary builds available [here](https://cdn.seedno.de/builds/trivia).

I only test the linux/amd64, linux/arm64, and windows/amd64 builds, the rest are all best-effort™.

Dockerfile available [here](https://raw.githubusercontent.com/Seednode/trivia/master/docker/Dockerfile).

An example instance with most features enabled can be found [here](https://trivia.seedno.de/).

### Configuration
The following configuration methods are accepted, in order of highest to lowest priority:
- Command-line flags
- Environment variables

## File format
The app expects newline-delimited text files in the following format (category is optional), with the file extension `.trivia` (configurable):
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
[...]
```

If the `--html` flag is passed, HTML can be used for formatting trivia questions:
```
What is the <u>current</u> year?|2024|History
How many inches are in a foot?|12|Measurement
Is mayonnaise an instrument?|No, <em>Patrick</em>, mayonnaise is not an instrument|Cartoons
[...]
```

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
If the `--reload` flag is passed, an additional `/reload` POST endpoint is registered.

The trivia database can be live-reloaded by calling this endpoint.

Scheduled index rebuilds can be enabled via the `--reload-interval <duration>` flag, which accepts [time.Duration](https://pkg.go.dev/time#ParseDuration) strings.

### Colors
A file containing custom hex color mappings for categories can be specified via the `-c|--colors` flag. 

The app expects the following format:
```
Entertainment|#da6ab2
History|#e5cb3a
Arts & Literature|#7a563c
Science & Nature|#157255
Sports & Leisure|#db6327
Global View|#6d6b82
Sound & Screen|#a04251
News|#b37e00
The Written Word|#7a4e34
Innovations|#4f7144
Game Time|#a66231
[...]
```

### Environment variables
Almost all options configurable via flags can also be configured via environment variables. 

The associated environment variable is the prefix `TRIVIA_` plus the flag name, with the following changes:
- Leading hyphens removed
- Converted to upper-case
- All internal hyphens converted to underscores

For example:
- `--colors /home/sinc/trivia/colors.txt` becomes `TRIVIA_COLORS=/home/sinc/trivia/colors.txt`
- `--recursive` becomes `TRIVIA_RECURSIVE=true`

## Usage output
```
Serves a basic trivia web frontend.

Usage:
  trivia [flags]

Flags:
  -b, --bind string              address to bind to (default "0.0.0.0")
  -c, --colors string            file from which to load color schemes
      --exit-on-error            shut down webserver on error, instead of just printing the error
      --export                   allow exporting of trivia database
      --extension string         only process files ending in this extension (leave empty to match all files) (default ".trivia")
  -h, --help                     help for trivia
      --html                     allow arbitrary html tags in input
  -p, --port uint16              port to listen on (default 8080)
      --profile                  register net/http/pprof handlers
  -r, --recursive                recurse into directories
      --reload                   allow live-reload of questions
      --reload-interval string   interval at which to rebuild question list (e.g. "5m" or "1h")
      --settings                 enable settings page at /settings (default true)
      --tls-cert string          path to TLS certificate
      --tls-key string           path to TLS keyfile
  -v, --verbose                  log requests to stdout
  -V, --version                  display version and exit
```

## Building the Docker image
From inside the cloned repository, build the image using the following command:

`REGISTRY=<registry url> LATEST=yes TAG=alpine ./build-docker.sh`
