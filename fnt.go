package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"bufio"
	"io"
	"log"
	"errors"
)

func main() {
	if (len(os.Args) < 2) {
		fmt.Printf("File Name Template Utility --\n\n")
		fmt.Printf("A command line utility to replace {{keys}} inside of a template file to\n")
		fmt.Printf("the associated values.\n\n")
		fmt.Printf("Usage:\t%s <template_file> value=key value1=key1 ...\n", os.Args[0])
		fmt.Printf("\tcat <key_value_file> | %s <template_file>\n\n", os.Args[0])
		fmt.Printf("Use %s --help to learn more.\n", os.Args[0])
		return
	}

	// if --help is specified, give the rundown
	for _, opt := range os.Args {
		if (opt == "--help" || opt == "-h") {
			display_full_help()
			return
		}
	}

	document, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	data, _ := getstdin()
	if (data != nil) {
		// fmt.Println("[+] Found stdin. Parsing...")
		process_keysets(document, data)
	} else {
		// fmt.Println("[ ] No stdin. Searching for keys in command-line arguments...")
		rendered_document := replace_keys_with_args(document)
		fmt.Println(string(rendered_document))
	}
}

// processes the key format file (from stdin or cli)
func process_keysets(document []byte, stdin_data *[]byte) {
	key_values := make(map[string]string)

	separator_re := regexp.MustCompile("\\-{3}") // seperators in the custom file are three dashes on their own line
	nl_re := regexp.MustCompile("\n")
	key_re := regexp.MustCompile("\\=")

	states_array := separator_re.Split(string(*stdin_data), -1)
	for _, state := range states_array {
		single_key_value_pair := nl_re.Split(state, -1)
		for _, line := range single_key_value_pair {
			key_val := key_re.Split(line, -1)
			if (len(key_val) == 2) {
				// fmt.Printf("found key: %s -> %s\n", key_val[0], key_val[1])
				key_values[string(key_val[0])] = string(key_val[1])
			}
		}

		render_and_save(document, key_values)
	}
}

func display_full_help() {
	fmt.Printf("File name utility is a program that replaces keys from within a file\n")
	fmt.Printf("with the appropriate value.\n\n")

	fmt.Printf("To create a template, create a file, and use the {{key}} syntax anywhere.\n")
	fmt.Printf("(e.g.) template.txt: '{{GREETING}} my name is {{NAME}}'\n\n")

	fmt.Printf("Next, use '%s template.txt GREETING=Hi! NAME=Robert'\n\n", os.Args[0])
	fmt.Printf("Output for single replacements is directed to the console.\nUse '>' to save it to a file.\n\n")

	fmt.Printf("File name utlity also supports importing key 'states'.\n")
	fmt.Printf("A key state is saved in a seperate file and looks like:\n\n")
	fmt.Println("GREETING=Hello\nNAME=Bob")
	fmt.Println("---")
	fmt.Println("GREETING=Hey,\nNAME=Alice")
	fmt.Println("---")
	fmt.Println("GREETING=Hi\nNAME=Ted\n")

	fmt.Printf("To use the key state, use 'cat <keystate_file> | %s <template_file>'\n", os.Args[0])
	fmt.Printf("Using a key state will write all output files to the working directory\n")
	fmt.Printf("with unique filenames that describe their values.\n")
}

// render and save a document
func render_and_save(document []byte, key_store map[string]string) {
	re := regexp.MustCompile("{{ *\\w+ *}}") // matches: {{key}} {{under_score}} {{ key  }}
	refine := regexp.MustCompile("\\w+") //matches: key under_store key (from above)

	matches := re.FindAll(document, -1)
	for _, match := range matches {
		stripped_key := refine.Find(match)
		stripped_key_value := key_store[string(stripped_key)] // find the key in the key_store map[string]string

		// someone who actually knows golang should probably fix this
		document = []byte(strings.Replace(string(document), string(match), string(stripped_key_value), -1))
	}

	f, err := os.Open(os.Args[1])
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}

	key_uniq_filename := fi.Name()
	for key, val := range key_store {
		key_uniq_filename += fmt.Sprintf("_%s=%s", key, strings.Replace(val, "/", "", -1))
	}

	fmt.Printf("Writing to %s\n", key_uniq_filename)
	ioutil.WriteFile(key_uniq_filename, document, 0777)
}

// processes stdin if it exists and return a memory address to the data
func getstdin() (stdin_data *[]byte, err error){
	s, _ := os.Stdin.Stat()
	if (s.Mode() & os.ModeCharDevice) != 0 {
		// there is no stdin, do nothing
		return nil, errors.New("[-] No stdin")
	}

	// there is stdin, process it
	nBytes, nChunks := int64(0), int64(0)
	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 10*1024) // 10MB -- CAN BE ADJUSTED IF PEOPLE ACTUALLY USE THIS

	// read from stdin
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		nChunks++
		nBytes += int64(len(buf))

		return &buf, nil

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
	}

	return nil, errors.New("[-] Missed buffer? Possibly out of memory?")
}

// attempts to parse a key=value pair from command line arguments
func get_key(key string) (value string, err error) {
	arg_array := os.Args
	re := regexp.MustCompile("\\=")
	for _, arg := range arg_array {
		key_val := re.Split(arg, 2)
		if (len(key_val) == 2) {
			// found a valid key-value pair in the args
			if (key_val[0] == key) {
				return key_val[1], nil
			}
		}
	}

	return "", fmt.Errorf("attempted to replace undefined key -> {{" + key + "}}")
}

// replaces {{ keys }} with their matching command line argument value
func replace_keys_with_args(document []byte) (rendered_document []byte) {
	re := regexp.MustCompile("{{ *\\w+ *}}") // matches: {{key}} {{under_score}} {{ key  }}
	refine := regexp.MustCompile("\\w+") // matches: key under_score key (from above)

	matches := re.FindAll(document, -1)
	for _, match := range matches {
		stripped_key := refine.Find(match)
		stripped_key_value, err := get_key(string(stripped_key))
		if err != nil {
			panic(err)
		}

		document = []byte(strings.Replace(string(document), string(match), string(stripped_key_value), -1))
	}

	return document
}
