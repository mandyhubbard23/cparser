package main

import (
	"os"
	"flag"
	"fmt"
	"bufio"
	"log"
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
	err := getCoverage(file, target)

	if err != nil {
		for k,v := range err {

			log.Fatal(k, v)
		}
		os.Exit(1)
	}

	os.Exit(0)
}

func getCoverageTarget(file string) float64 {

	lines, err := readLines(file)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error in readLines: %s", err))
	}

	if len(lines) == 0 || len(lines)> 1 {
		log.Fatal("The coverage target file should contain only one line.")
	}

	target, err := strconv.ParseFloat(lines[0], 64)
	if err != nil {

		log.Fatal(fmt.Sprintf("Could not convert coverage target - %s - to a Float", err))
	}

	return target
}

func getCoverage(file string, target float64) map[string]string {

	var (
		cmdOut  []byte
		err     error
		percent float64
		match   []string
	)

	errorMap := make(map[string]string)
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
			panic(err)
		}
		if percent < target {

			errorMap[match[1]] = fmt.Sprintf("Target Code Coverage=%f, Actual Code Coverage=%f", target, percent)
		}
	}
	return errorMap
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