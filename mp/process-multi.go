package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type Token struct {
	Value string
}

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

	tokenQueue := make(chan Token, numTokens)
	wg.Add(1)
	go enqueueTokens(fileHandle, tokenQueue, numTokens)

	stateChannelCount := 0
	totalCount := 0
	threadCount := lineCount / 5 //10 // Number of threads in the threadpool
	results := make(chan int)

	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go runQfabCLI(tokenQueue, outFile, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		stateChannelCount += result
		totalCount++
		fmt.Printf("\rProcessing %d of %d", totalCount, numTokens)
	}

	finals, err := os.OpenFile(resultsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer finals.Close()
	summary := fmt.Sprintf("\nstate channel = %d \n total = %d", stateChannelCount, totalCount)
	fmt.Println("\nSuummary")
	fmt.Println(summary)
	finals.WriteString(summary)
}

func enqueueTokens(fileHandle *os.File, tokenQueue chan<- Token, numTokens int) {
	defer wg.Done()

	scanner := bufio.NewScanner(fileHandle)
	i := 0
	for scanner.Scan() {
		if i == numTokens {
			break
		}
		tokens := scanner.Text()
		tokenQueue <- Token{Value: tokens}
		i++
	}

	close(tokenQueue)
}

func runQfabCLI(tokenQueue <-chan Token, outFile *os.File, results chan<- int) {
	defer wg.Done()

	for token := range tokenQueue {
		stateChannelCount, _ := executeQfabCLI(token.Value, outFile)
		results <- stateChannelCount
	}
}

func executeQfabCLI(formattedLine string, outFile *os.File) (int, int) {
	command := exec.Command("qfab_cli", "tools", "decode", formattedLine)
	stateChannelCount := 0
	totalCount := 0

	output, err := command.Output()
	if err != nil {
		log.Fatal(err)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "legacy token: ") {
		lines := strings.Split(outputStr, "\n")
		fileEntry := ""
		for _, line := range lines {
			if strings.Contains(line, "MAPPED PREFIX") {
				fileEntry = fmt.Sprintf("legacy token: MAPPED PREFIX=%s\n", strings.TrimSpace(strings.TrimPrefix(line, "MAPPED PREFIX")))
				break
			}
		}
		outFile.WriteString(fileEntry)
		return 0, 1
	}

	lines := strings.Split(outputStr, "\n")
	writeOutput := "state-channel:"
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "PREFIX") {
			prefix := strings.TrimSpace(strings.TrimPrefix(line, "PREFIX"))
			if strings.Contains(prefix, "asc=state-channel") {
				stateChannelCount++
			}
			writeOutput += fmt.Sprintf(" PREFIX=%s", prefix)
			totalCount++
		} else if strings.HasPrefix(strings.TrimSpace(line), "json") {
			json := strings.TrimSpace(strings.TrimPrefix(line, "json"))
			writeOutput += fmt.Sprintf(" json=%s", json)
		}
	}
	writeOutput += "\n"
	outFile.WriteString(writeOutput)

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
