package main

import (
	"os"
	"flag"
	"fmt"
	"bufio"
	"strconv"
	"os/exec"
	"strings"
	"regexp"
)

var file string
var coverageFile string

func main() {

	flag.StringVar(&file, "report", "", "The fully qualified path of the coverage report output file to parse.")
	flag.StringVar(&coverageFile, "target", "", "The fully qualified path of the file containing the code coverage target.")
	flag.Parse()

	if file == ""   {
		flag.PrintDefaults()
		os.Exit(1)

	}

	target := getCoverageTarget(coverageFile)
	err, failed := getCoverage(file, target)

	if failed {

		fmt.Println(err)
		for k,v := range err {

			fmt.Fprintln(os.Stderr, k, v)
		}
		fmt.Fprintln(os.Stderr, "Exiting because coverage was below target: ", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "Success")
	os.Exit(0)
}

func getCoverageTarget(file string) float64 {

	lines, err := readLines(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in readLines:", err)
	}

	if len(lines) == 0 || len(lines)> 1 {
		fmt.Fprintln(os.Stderr, "The coverage target file should contain only one line.")
	}

	target, err := strconv.ParseFloat(lines[0], 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not convert coverage target - %s - to a Float", err)
	}

	return target
}

func getCoverage(file string, target float64) (map[string]string, bool) {

	var (
		cmdOut  []byte
		err     error
		percent float64
		match   []string
	)

	errorMap := make(map[string]string)
	failed := false

	cmd := "cat " + file + "| pup [id=files] option text{}"

	if cmdOut, err = exec.Command("bash", "-c", cmd).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running the command: ", err)
		os.Exit(1)
	}

	values := strings.Split(string(cmdOut), "\n")
	r := regexp.MustCompile(`(.*)\((.*?)\)`)

	for _, v := range values {
		match = r.FindStringSubmatch(v)
		if match == nil {
			continue
		}
		if percent, err = strconv.ParseFloat(strings.Replace(string(match[2]), "%", "", 1), 64); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing target percent into float: ", err)
		}
		if percent < target {

			errorMap[match[1]] = fmt.Sprintf("Target Code Coverage=%f, Actual Code Coverage=%f", target, percent)
			failed = true
		}
	}
	return errorMap, failed
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}