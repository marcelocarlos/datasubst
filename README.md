# datasubst

A simple [go template](https://golang.org/pkg/text/template/) based tool that supports JSON, YAML and environment variables as data sources.

This tool has been written as an alternative to `envsubst` in order to support additional data source formats, such as YAML and JSON files. Since it is powered by go template, [built-in functions](https://golang.org/pkg/text/template/#hdr-Functions), loops, conditionals and more can be used for extra flexibility.

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

If you don't want to install it, you can run `datasubst` via Docker instead:

```shell
docker run --rm ghcr.io/marcelocarlos/datasubst:v0
```

### GitHub Action

There is also a [setup-datasubst](https://github.com/marcelocarlos/setup-datasubst) action available so you can use
datasubst in your GitHub Actions workflow. To use it, just add the following step in you workflow:

```yaml
- name: Setup datasubst
  uses: marcelocarlos/setup-datasubst@v1
  with:
    version: v0.7.0
```

## Usage

```shell
# Using JSON as data source
datasubst --json-data examples/basic-data.json -i examples/basic-input.txt
# Using YAML as data source
datasubst --yaml-data examples/basic-data.yaml -i examples/basic-input.txt
# Using environment variables as data source
TEST1="hello" TEST2="world" datasubst --input examples/basic-input-env.txt --env-data

# Using stdin - JSON
echo "v1: {{ .key1 }}" | datasubst --json-data examples/basic-data.json
# Using stdin - YAML
echo "v3: {{ .key2.first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml
# Using stdin - env
echo "{{ .TEST1 }} {{ .TEST2 }}" | TEST1="hello" TEST2="world" datasubst --env-data

# Specifying JSON subtrees to use (available for JSON and YAML)
echo "{{ .first.key3 }}" | datasubst --json-data examples/basic-data.json --subtree .key2
# Specifying YAML subtrees to use (available for JSON and YAML)
datasubst --json-data examples/basic-data.json --subtree .key2 -i examples/basic-input-subtree.txt

# Using additional options, such -s (strict mode) and -d (change delimiters)
echo "(( .TEST ))" | TEST="hi" datasubst --env-data -d '((:))' -s
```

See [examples](./examples/) for more.

## Build from source

```shell
go build -ldflags=-X=main.Version=$(git describe --tags)
```
