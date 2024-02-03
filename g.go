package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

func performRequest(url string, data string, cookie string) (bool, string, float64, error) {
	urlWithData := fmt.Sprintf("%s%s", url, data)
	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlWithData, nil)
	if err != nil {
		return false, urlWithData, 0, err
	}

	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}

	response, err := client.Do(req)
	if err != nil {
		return false, urlWithData, 0, err
	}
	defer response.Body.Close()

	responseTime := time.Since(startTime).Seconds()

	return true, urlWithData, responseTime, nil
}

func main() {
	urlsFile := flag.String("l", "", "url list GET.")
	dataFile := flag.String("p", "", "paylode list.")
	cookie := flag.String("c", "", "Cookie GET.")
	outputFile := flag.String("o", "", "output 20 second.")
	flag.Parse()

	if *urlsFile == "" || *dataFile == "" {
		fmt.Println("Debe proporcionar archivos de URLs y datos.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	urlsContent, err := ioutil.ReadFile(*urlsFile)
	if err != nil {
		fmt.Println("Error leyendo el archivo de URLs:", err)
		os.Exit(1)
	}
	urls := string(urlsContent)
	urlList := splitLines(urls)

	dataContent, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		fmt.Println("Error leyendo el archivo de datos:", err)
		os.Exit(1)
	}
	data := string(dataContent)
	dataList := splitLines(data)

	var outputLines []string
	var vulnerableURLs []string

	for _, url := range urlList {
		// Skip testing payloads on URLs that are already marked as vulnerable
		if contains(vulnerableURLs, url) {
			fmt.Printf("Skipping payloads on vulnerable URL: %s\n", url)
			continue
		}

		for _, d := range dataList {
			success, urlWithData, responseTime, errorMessage := performRequest(url, d, *cookie)

			if success && responseTime <= 20 {
				fmt.Printf("\033[1;32mURL %s - %.2f second\033[0m\n", urlWithData, responseTime)
			} else {
				outputLine := fmt.Sprintf("\033[1;31mURL %s - %.2f second - Error: %v\033[0m", urlWithData, responseTime, errorMessage)
				fmt.Println(outputLine)
				outputLines = append(outputLines, outputLine)

				if success && responseTime > 20 {
					fmt.Printf("\033[1;33mURL %s is vulnerable!\033[0m\n", urlWithData)
					vulnerableURLs = append(vulnerableURLs, url)
					break // Stop testing payloads on this URL once it's marked as vulnerable
				}
			}
		}
	}

	// Write results to the output file if specified
	if *outputFile != "" {
		outputContent := strings.Join(outputLines, "\n")
		err := ioutil.WriteFile(*outputFile, []byte(outputContent), 0644)
		if err != nil {
			fmt.Println("Error writing results to the output file:", err)
			os.Exit(1)
		}
		fmt.Printf("Results written to %s\n", *outputFile)
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}

func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
