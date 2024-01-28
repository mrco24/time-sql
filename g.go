package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	payloadsPath string
	urlsPath     string
	url          string
	outputPath   string
	verbose      bool
	threads      int
)

func init() {
	flag.StringVar(&payloadsPath, "p", "", "Path to the payloads file")
	flag.StringVar(&urlsPath, "l", "", "Path to the URLs file")
	flag.StringVar(&url, "u", "", "Single URL to test with all payloads")
	flag.StringVar(&outputPath, "o", "", "Path to the output file")
	flag.BoolVar(&verbose, "v", false, "Run in verbose mode")
	flag.IntVar(&threads, "t", 1, "Number of concurrent threads")
	flag.Parse()
}

// checkVulnerability checks if a URL with a specific payload is vulnerable
func checkVulnerability(url string, payload string, outputFile *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", payload)

	start := time.Now()
	res, err := client.Do(req)
	elapsed := time.Since(start).Seconds()

	if err != nil {
		log.Printf("The request was not successful due to: %v\n", err)
		return
	}

	defer res.Body.Close()

	// Create color instances
	urlColor := color.New(color.FgCyan)
	payloadColor := color.New(color.FgMagenta)
	timeColor := color.New(color.FgBlue)
	statusColor := color.New(color.FgGreen)

	// Apply colors to each part of the result
	urlColor.Printf("Testing for URL: %s\n", url)
	payloadColor.Printf("Payload: %s\n", payload)
	timeColor.Printf("Response Time: %.2f seconds\n", elapsed)

	result := ""
	if elapsed >= 25 && elapsed <= 50 {
		statusColor.Printf("Status: Vulnerable\n\n")
		result = "Vulnerable"
	} else {
		statusColor.Printf("Status: Not Vulnerable\n\n")
		result = "Not Vulnerable"
	}

	if outputFile != nil {
		outputFile.WriteString(fmt.Sprintf("URL: %s, Payload: %s, Response Time: %.2f seconds, Status: %s\n",
			url, payload, elapsed, result))
	}
}

func main() {
	if payloadsPath == "" {
		fmt.Println("Error: Payloads file path is required.")
		return
	}

	payloads := readLines(payloadsPath)

	if outputPath != "" {
		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFile.Close()

		if url != "" {
			var wg sync.WaitGroup
			for _, payload := range payloads {
				wg.Add(1)
				go checkVulnerability(url, payload, outputFile, &wg)
			}
			wg.Wait()
		} else if urlsPath != "" {
			urls := readLines(urlsPath)

			var wg sync.WaitGroup
			for _, u := range urls {
				for _, payload := range payloads {
					wg.Add(1)
					go checkVulnerability(u, payload, outputFile, &wg)
				}
			}
			wg.Wait()
		} else {
			fmt.Println("Error: URLs file path is required.")
		}
	} else {
		fmt.Println("Error: Output file path is required.")
	}
}

func readLines(filePath string) []string {
	lines := []string{}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}
