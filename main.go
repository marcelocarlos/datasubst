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
	"text/template"

	"gopkg.in/yaml.v3"
)

const usage = `Usage:
    datasubst (--json-data DATA_INPUT | --yaml-data DATA_INPUT) [-i INPUT] [-o OUTPUT]

Options:
    -j, --json-data DATA_INPUT   Input data source in JSON format.
    -y, --yaml-data DATA_INPUT   Input data source in YAML format.
    -i, --input INPUT            Input template file in go template format.
    -o, --output OUTPUT          Write the output to the file at OUTPUT.
        --help PATH              Display this help and exit.
        --version                Output version information and exit.

INPUT defaults to standard input and OUTPUT defaults to standard output.

Examples:
    $ datasubst --input examples/basic-input.txt --json-data examples/basic-data.json
    $ echo "v3: {{ .key2.first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml`

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

func main() {
	log.SetFlags(0)
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	if len(os.Args) == 1 {
		log.Fatalf("%s\n", usage)
	}

	var (
		inputFile, outputFile, jsonDataFile, yamlDataFile string
		helpFlag, versionFlag                             bool
	)

	flag.StringVar(&inputFile, "input", "", "input template file in go template format")
	flag.StringVar(&inputFile, "i", "", "input template file in go template format")
	flag.StringVar(&jsonDataFile, "json-data", "", "input data source in JSON format")
	flag.StringVar(&jsonDataFile, "j", "", "input data source in JSON format")
	flag.StringVar(&outputFile, "output", "", "write the output to the file at OUTPUT")
	flag.StringVar(&outputFile, "o", "", "write the output to the file at OUTPUT")
	flag.StringVar(&yamlDataFile, "yaml-data", "", "input data source in YAML format")
	flag.StringVar(&yamlDataFile, "y", "", "input data source in YAML format")
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

	if (jsonDataFile == "" && yamlDataFile == "") || (jsonDataFile != "" && yamlDataFile != "") {
		log.Fatal("Error: please specify --json-data or --yaml-data")
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
	} else {
		data, err = parseYAML(yamlDataFile)
	}
	if err != nil {
		log.Fatalf("Error opening data file: %v\n", err)
	}
	// Parse Input
	tpl, err := template.New("template").Parse(string(tplStr))
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
