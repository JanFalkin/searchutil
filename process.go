package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func countLines(filename string) int {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lineCount
}

func processFile(dataPath string, numTokens int, resultsPath string) {
	i := 0
	lineCount := countLines(dataPath)
	if numTokens < 0 {
		numTokens = lineCount
	}

	fileHandle, err := os.Open(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileHandle.Close()

	outFile, err := os.Create(resultsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	stateChannelCount := 0
	totalCount := 0
	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		if i == numTokens {
			break
		}
		tokens := scanner.Text()
		fmt.Printf("%d out of %d stat: state channel = %d total = %d\r", i, numTokens, stateChannelCount, totalCount)

		sc, t := runQfabCLI(tokens, outFile)
		stateChannelCount += sc
		totalCount += t
		i++
	}

	finals, err := os.OpenFile(resultsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer finals.Close()

	finals.WriteString(fmt.Sprintf("state channel = %d \n total = %d", stateChannelCount, totalCount))
}

func runQfabCLI(formattedLine string, outFile *os.File) (int, int) {
	command := exec.Command("qfab_cli", "tools", "decode", formattedLine)
	stateChannelCount := 0
	totalCount := 0

	output, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "legacy token: ") {
		outFile.WriteString(outputStr + "\n")
		return 0, 1
	}

	splitOut := strings.Join(strings.Split(outputStr, "\n")[2:], "")
	postSplit := strings.Split(splitOut, "TOKEN")

	prefixInfo := strings.Split(postSplit[1], "PREFIX      ")
	if len(prefixInfo) > 1 {
		outFile.WriteString(prefixInfo[1] + "\n")
		if strings.Contains(prefixInfo[1], "asc=state-channel") {
			stateChannelCount++
			totalCount++
		} else {
			totalCount++
		}
	} else {
		ps := strings.Join(postSplit, "")
		fmt.Printf("prefix info unexpected = %v post_split=%v\n", prefixInfo, ps)
	}

	return stateChannelCount, totalCount
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: process <input_file> <num_tokens> <output_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	numTokens, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("num_tokens must be an integer")
		os.Exit(1)
	}
	outputFile := os.Args[3]

	processFile(inputFile, numTokens, outputFile)
}
