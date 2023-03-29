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
    -t, --subtree                JSON and YAML only, use a subtree of the data source instead of the full contents
    -e, --env-data               Input data source comes from environment variables.
    -i, --input INPUT            Input template file or directory containig template(s) in go template format.
    -o, --output OUTPUT          Write the output to the file at OUTPUT.
    -s, --strict                 Strict mode (causes an error if a key is missing)
    -d, --delimiters             Set the delimiters used in the templates in the format <left>:<right> (default: '{{:}}')
        --help                   Display this help and exit.
        --version                Output version information and exit.

INPUT defaults to standard input and OUTPUT defaults to standard output.

Examples:
    $ datasubst --input examples/basic-input.txt --json-data examples/basic-data.json
    $ echo "v3: {{ .key2.first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml
    $ echo "{{ .TEST1 }} {{ .TEST2 }}" | TEST1="hello" TEST2="world" datasubst --env-data
    $ echo "(( .TEST ))" | TEST="hi" datasubst --env-data -d '((:))'
		$ echo "v3: {{ .first.key3 }}" | datasubst --yaml-data examples/basic-data.yaml --subtree .key2`

var Version string

var (
	inputFile, outputFile, jsonDataFile, yamlDataFile, delimiters, subtree string
	envFlag, strictFlag, helpFlag, versionFlag                             bool
)

func main() {
	log.SetFlags(0)
	parseArgs()

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
		if subtree != "" {
			data = getSubTree(data, subtree)
		}
	} else if yamlDataFile != "" {
		data, err = parseYAML(yamlDataFile)
		if subtree != "" {
			data = getSubTree(data, subtree)
		}
	} else {
		data, err = parseEnv()
	}
	if err != nil {
		log.Fatalf("Error opening data file: %v\n", err)
	}

	// Prepare Template
	tpl := template.New("template")
	if strictFlag {
		tpl.Option("missingkey=error")
	}
	if delimiters != "" {
		if strings.Count(delimiters, ":") != 1 || delimiters[len(delimiters)-1:] == ":" || delimiters[0:1] == ":" {
			log.Fatal("Error: invalid delimiter format. Must be '<left>:<right>' and ':'")
		}
		d := strings.Split(delimiters, ":")
		tpl.Delims(d[0], d[1])
	}
	tpl, err = tpl.Parse(string(tplStr))
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

func getSubTree(data interface{}, substree string) interface{} {
	st := strings.Split(subtree, ".")[1:]
	for _, k := range st {
		v := data.(map[string]interface{})
		data = v[k]
	}
	return data
}

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

func parseArgs() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }
	if len(os.Args) == 1 {
		log.Fatalf("%s\n", usage)
	}

	flag.StringVar(&inputFile, "input", "", "input template file or directory containig template(s) in go template format")
	flag.StringVar(&inputFile, "i", "", "input template file or directory containig template(s) in go template format")
	flag.StringVar(&jsonDataFile, "json-data", "", "input data source in JSON format")
	flag.StringVar(&jsonDataFile, "j", "", "input data source in JSON format")
	flag.StringVar(&subtree, "subtree", "", "subtree to be used (e.g. .my_key.my_subkey)")
	flag.StringVar(&subtree, "t", "", "subtree to be used (e.g. .my_key.my_subkey)")
	flag.BoolVar(&envFlag, "env-data", false, "input data source comes from environment variables")
	flag.BoolVar(&envFlag, "e", false, "input data source comes from environment variables")
	flag.StringVar(&outputFile, "output", "", "write the output to the file at OUTPUT")
	flag.StringVar(&outputFile, "o", "", "write the output to the file at OUTPUT")
	flag.StringVar(&yamlDataFile, "yaml-data", "", "input data source in YAML format")
	flag.StringVar(&yamlDataFile, "y", "", "input data source in YAML format")
	flag.StringVar(&delimiters, "delimiters", "", "Set the delimiters used in the templates in the format <left>:<right> (default: '{{:}}')")
	flag.StringVar(&delimiters, "d", "", "Set the delimiters used in the templates in the format <left>:<right> (default: '{{:}}')")
	flag.BoolVar(&strictFlag, "strict", false, "strict mode (causes an error if a key is missing)")
	flag.BoolVar(&strictFlag, "s", false, "strict mode (causes an error if a key is missing)")
	flag.BoolVar(&versionFlag, "version", false, "output version information and exit")
	flag.BoolVar(&helpFlag, "help", false, "display this help and exit")
	flag.Parse()

	if versionFlag {
		if Version != "" {
			fmt.Println(Version)
			os.Exit(0)
		}
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(buildInfo.Main.Version)
			os.Exit(0)
		}
		fmt.Println("(unknown)")
		os.Exit(0)
	}

	if helpFlag {
		fmt.Println(usage)
		os.Exit(0)
	}

	if countTrue(jsonDataFile != "", yamlDataFile != "", envFlag) != 1 {
		log.Fatal("Error: please specify --json-data, --yaml-data or --env-data")
	}
}
