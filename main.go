package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const usage = `Usage:
    datasubst (--json-data DATA_INPUT | --yaml-data DATA_INPUT | --env-data) [-i INPUT] [-o OUTPUT]

Options:
    -j, --json-data DATA_INPUT   Input data source in JSON format.
    -y, --yaml-data DATA_INPUT   Input data source in YAML format.
    -e, --env-data               Input data source comes from environment variables.
    -i, --input INPUT            Input template file in go template format.
    -o, --output OUTPUT          Write the output to the file at OUTPUT.
    -s, --strict                 Strict mode (causes an error if a key is missing)
        --help                   Display this help and exit.
        --version                Output version information and exit.

INPUT defaults to standard input and OUTPUT defaults to standard output.

Examples:
    $ datasubst --input examples/basic-input.txt --json-data examples/basic-data.json
    $ echo "v3: {{ .key2.first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml
    $ TEST1="hello" TEST2="world" datasubst --input examples/basic-input-env.txt --env-data`

var Version string

func parseYAML(yamlDataFile string) (interface{}, error) {
	var data interface{}
	dataFile, err := os.Open(filepath.Clean(yamlDataFile))
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()
	err = yaml.NewDecoder(dataFile).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseJSON(jsonDataFile string) (interface{}, error) {
	var data interface{}
	dataFile, err := os.Open(filepath.Clean(jsonDataFile))
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()
	err = json.NewDecoder(dataFile).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseEnv() (interface{}, error) {
	data := make(map[string]string)
	for _, v := range os.Environ() {
		envKv := strings.Split(v, "=")
		data[envKv[0]] = envKv[1]
	}
	return data, nil
}

func countTrue(b ...bool) int {
	n := 0
	for _, v := range b {
		if v {
			n++
		}
	}
	return n
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	if len(os.Args) == 1 {
		log.Fatalf("%s\n", usage)
	}

	var (
		inputFile, outputFile, jsonDataFile, yamlDataFile string
		envFlag, strictFlag, helpFlag, versionFlag        bool
	)

	flag.StringVar(&inputFile, "input", "", "input template file in go template format")
	flag.StringVar(&inputFile, "i", "", "input template file in go template format")
	flag.StringVar(&jsonDataFile, "json-data", "", "input data source in JSON format")
	flag.StringVar(&jsonDataFile, "j", "", "input data source in JSON format")
	flag.BoolVar(&envFlag, "env-data", false, "input data source comes from environment variables")
	flag.BoolVar(&envFlag, "e", false, "input data source comes from environment variables")
	flag.StringVar(&outputFile, "output", "", "write the output to the file at OUTPUT")
	flag.StringVar(&outputFile, "o", "", "write the output to the file at OUTPUT")
	flag.StringVar(&yamlDataFile, "yaml-data", "", "input data source in YAML format")
	flag.StringVar(&yamlDataFile, "y", "", "input data source in YAML format")
	flag.BoolVar(&strictFlag, "strict", false, "strict mode (causes an error if a key is missing)")
	flag.BoolVar(&strictFlag, "s", false, "strict mode (causes an error if a key is missing)")
	flag.BoolVar(&versionFlag, "version", false, "output version information and exit")
	flag.BoolVar(&helpFlag, "help", false, "display this help and exit")
	flag.Parse()

	if versionFlag {
		if Version != "" {
			fmt.Println(Version)
			return
		}
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(buildInfo.Main.Version)
			return
		}
		fmt.Println("(unknown)")
		return
	}

	if helpFlag {
		fmt.Println(usage)
		return
	}

	if countTrue(jsonDataFile != "", yamlDataFile != "", envFlag) != 1 {
		log.Fatal("Error: please specify --json-data, --yaml-data or --env-data")
	}
	// Read input
	in := os.Stdin
	if inputFile != "" && inputFile != "-" {
		f, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("Error opening input file: %v\n", err)
		}
		defer f.Close()
		in = f
	}
	tplStr, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("Error reading input file: %v\n", err)
	}
	// Read and Parse data file
	var data interface{}
	if jsonDataFile != "" {
		data, err = parseJSON(jsonDataFile)
	} else if yamlDataFile != "" {
		data, err = parseYAML(yamlDataFile)
	} else {
		data, err = parseEnv()
	}
	if err != nil {
		log.Fatalf("Error opening data file: %v\n", err)
	}
	// Parse Input
	missingkey := "default"
	if strictFlag {
		missingkey = "error"
	}
	tpl, err := template.New("template").Option(fmt.Sprintf("missingkey=%s", missingkey)).Parse(string(tplStr))
	if err != nil {
		log.Fatalf("Error parsing template: %v\n", err)
	}
	// Render
	out := os.Stdout
	if outputFile != "" && outputFile != "-" {
		out, err = os.Create(outputFile)
		if err != nil {
			log.Fatalf("Error creating output file: %v\n", err)
		}
		defer out.Close()
	}
	err = tpl.Execute(out, data)
	if err != nil {
		log.Fatalf("Error rendering template: %v\n", err)
	}
}
