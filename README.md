# datasubst

A simple [go template](https://golang.org/pkg/text/template/) based tool based that uses structured data like JSON and YAML as data source.

This tool has been initially written as an alternative to `envsubst` in order to support YAML files as data sources. Since it is powered by go template, [built-in functions](https://golang.org/pkg/text/template/#hdr-Functions), loops, conditionals, etc can be used offering extra flexibility.

## Installation

On macOS or Linux, a Homebrew formula is available:

```shell
brew tap marcelocarlos/tap
brew install datasubst
```

Alternatively, you can install a development version using:

```shell
go get github.com/marcelocarlos/datasubst
```

## Usage

```shell
# Using JSON as data source
datasubst --json-data examples/basic-data.json -i examples/basic-input.txt
# Using YAML as data source
datasubst --yaml-data examples/basic-data.yaml -i examples/basic-input.txt

# Using stdin - JSON
echo "v1: {{ .key1 }}" | datasubst --json-data examples/basic-data.json
# Using stdin - YAML
echo "v3: {{ .key2.first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml
```

See [examples](./examples/) for more.

## Build from source

```shell
go build -ldflags=-X=main.Version=$(git describe --tags)
```
