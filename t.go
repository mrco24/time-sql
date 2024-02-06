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

func performRequest(url string, data string, cookie string, wg *sync.WaitGroup, ch chan struct{}, vulnerableURLs chan<- string, result chan<- string) {
	defer wg.Done()

	urlWithData := fmt.Sprintf("%s%s", url, data)

	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlWithData, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", urlWithData, err)
		<-ch // Release the channel to avoid deadlock
		return
	}

	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error performing request for %s: %v\n", urlWithData, err)
		<-ch // Release the channel to avoid deadlock
		return
	}
	defer response.Body.Close()

	responseTime := time.Since(startTime).Seconds()

	mu.Lock()
	defer mu.Unlock()

	if responseTime <= 20 {
		fmt.Printf("\033[1;32mURL %s - %.2f seconds\033[0m\n", urlWithData, responseTime)
	} else if responseTime <= 30 { // Skip URLs taking more than 30 seconds
		fmt.Printf("\033[1;31mURL %s - %.2f seconds - Vulnerable!\033[0m\n", urlWithData, responseTime)
		vulnerableURLs <- fmt.Sprintf("Vulnerable URL: %s - Response Time: %.2f seconds", urlWithData, responseTime)
	}

	<-ch
}

func writeResultsToFile(outputFile string, vulnerableURLs <-chan string, done chan<- struct{}) {
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		done <- struct{}{}
		return
	}
	defer file.Close()

	for vulnURL := range vulnerableURLs {
		_, err := file.WriteString(vulnURL + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			done <- struct{}{}
			return
		}
		fmt.Printf("Vulnerable URL written to %s\n", outputFile)
	}

	done <- struct{}{}
}

func main() {
	urlsFile := flag.String("l", "", "File containing URLs to be requested with GET.")
	dataFile := flag.String("p", "", "File containing data to be added to the URLs.")
	cookie := flag.String("c", "", "Cookie to include in the GET request.")
	outputFile := flag.String("o", "", "Output file for vulnerable URLs.")
	threads := flag.Int("t", 10, "Number of concurrent goroutines.")
	flag.Parse()

	if *urlsFile == "" || *dataFile == "" || *outputFile == "" {
		fmt.Println("Please provide valid values for -l, -p, and -o.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	urlsContent, err := ioutil.ReadFile(*urlsFile)
	if err != nil {
		fmt.Println("Error reading the URLs file:", err)
		os.Exit(1)
	}
	urls := string(urlsContent)
	urlList := splitLines(urls)

	dataContent, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		fmt.Println("Error reading the data file:", err)
		os.Exit(1)
	}
	data := string(dataContent)
	dataList := splitLines(data)

	var wg sync.WaitGroup
	ch := make(chan struct{}, *threads)
	vulnerableURLs := make(chan string, len(urlList)*len(dataList))
	done := make(chan struct{})

	go func() {
		writeResultsToFile(*outputFile, vulnerableURLs, done)
	}()

	for _, url := range urlList {
		for _, d := range dataList {
			wg.Add(1)
			ch <- struct{}{}
			go performRequest(url, d, *cookie, &wg, ch, vulnerableURLs, nil)
		}
	}

	wg.Wait()

	// Close the vulnerableURLs channel to signal the output writer goroutine to exit
	close(vulnerableURLs)

	// Wait for the output writer to finish writing results to the file
	<-done
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}
