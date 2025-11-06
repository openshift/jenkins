package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var re = regexp.MustCompile(`[Tt]o exclude this rule add\s+"([^"]+)"`)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: knflx-ecp-gen <log-file> <template-file>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 100*1024*1024), 100*1024*1024)

	var start int
	for n := 1; s.Scan(); n++ {
		if strings.Contains(s.Text(), "step-report-json") {
			start = n + 1
			break
		}
	}
	if err := s.Err(); err != nil {
		log.Fatalf("Error reading log file: %v", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		log.Fatalf("Error seeking file: %v", err)
	}

	s = bufio.NewScanner(f)
	s.Buffer(make([]byte, 100*1024*1024), 100*1024*1024)

	var js strings.Builder
	for n := 1; s.Scan(); n++ {
		if n >= start {
			js.WriteString(s.Text() + "\n")
			var d interface{}
			if json.Unmarshal([]byte(js.String()), &d) == nil {
				t, err := template.ParseFiles(os.Args[2])
				if err != nil {
					log.Fatalf("Failed to parse template: %v", err)
				}
				if err := t.Execute(os.Stdout, extract(d)); err != nil {
					log.Fatalf("Failed to execute template: %v", err)
				}
				fmt.Fprintln(os.Stderr, "Use yq command to merge the exclusions:")
				fmt.Fprintln(os.Stderr, "REPLACE=foo.yaml")
				fmt.Fprintln(os.Stderr, "INPUT=bar.yaml")
				fmt.Fprintln(os.Stderr, `yq eval ".spec.sources[0].volatileConfig = load(env(REPLACE)).volatileConfig" "$INPUT"`)
				return
			}
		}
	}
	if err := s.Err(); err != nil {
		log.Fatalf("Error reading log file: %v", err)
	}
	log.Fatal("Failed to find valid JSON in log file")
}

func extract(d interface{}) (r []string) {
	m := map[string]bool{}
	var walk func(interface{})
	walk = func(v interface{}) {
		switch x := v.(type) {
		case map[string]interface{}:
			for _, v := range x {
				walk(v)
			}
		case []interface{}:
			for _, v := range x {
				walk(v)
			}
		case string:
			for _, s := range re.FindAllStringSubmatch(x, -1) {
				if len(s) > 1 {
					m[s[1]] = true
				}
			}
		}
	}
	walk(d)
	for k := range m {
		r = append(r, k)
	}
	sort.Strings(r)
	return
}
