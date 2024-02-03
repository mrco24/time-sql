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

func performRequest(url string, data string, cookie string, wg *sync.WaitGroup, ch chan struct{}, vulnerableURLs chan<- string, result chan<- string, foundFlag *bool) {
	defer wg.Done()

	urlWithData := fmt.Sprintf("%s%s", url, data)

	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlWithData, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", urlWithData, err)
		return
	}

	if cookie != "" {
		req.Header.Add("Cookie", cookie)
	}

	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error performing request for %s: %v\n", urlWithData, err)
		return
	}
	defer response.Body.Close()

	responseTime := time.Since(startTime).Seconds()

	mu.Lock()
	defer mu.Unlock()

	if responseTime <= 20 {
		fmt.Printf("\033[1;32mURL %s - %.2f seconds\033[0m\n", urlWithData, responseTime)
	} else {
		fmt.Printf("\033[1;31mURL %s - %.2f seconds - Vulnerable!\033[0m\n", urlWithData, responseTime)
		vulnerableURLs <- urlWithData

		// Save the first vulnerable URL and set the flag
		if !*foundFlag {
			*foundFlag = true

			// Send result data with URL and response time
			result <- fmt.Sprintf("%s|%.2f", urlWithData, responseTime)
		}
	}
	<-ch
}

func writeResultsToFile(outputFile string, result <-chan string, done chan<- struct{}) {
	for {
		select {
		case resultData, ok := <-result:
			if !ok {
				// result channel closed, stop writing
				close(done)
				return
			}

			// Split the resultData into URL and response time
			parts := strings.Split(resultData, "|")
			if len(parts) != 2 {
				fmt.Println("Invalid result format:", resultData)
				continue
			}
			url := parts[0]
			responseTime := parts[1]

			// Append the result to the output file
			err := ioutil.WriteFile(outputFile, []byte(fmt.Sprintf("Vulnerable URL: %s - Response Time: %s seconds\n", url, responseTime)), 0644)
			if err != nil {
				fmt.Println("Error writing result to the output file:", err)
				os.Exit(1)
			}
			fmt.Printf("Vulnerable URL written to %s\n", outputFile)
		}
	}
}

func main() {
	urlsFile := flag.String("l", "", "Archivo de texto con las URLs a las que se les realizará la petición GET.")
	dataFile := flag.String("p", "", "Archivo de texto con los datos que se agregarán a las URLs.")
	cookie := flag.String("c", "", "Cookie a incluir en la petición GET.")
	outputFile := flag.String("o", "", "Archivo de salida para el primer URL vulnerable encontrado.")
	threads := flag.Int("t", 10, "Número de goroutines concurrentes.")
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

	var wg sync.WaitGroup
	ch := make(chan struct{}, *threads)
	vulnerableURLs := make(chan string, len(urlList)*len(dataList))
	result := make(chan string, 1)
	done := make(chan struct{})

	go func() {
		writeResultsToFile(*outputFile, result, done)
	}()

	foundFlag := false

	for _, url := range urlList {
		for _, d := range dataList {
			wg.Add(1)
			ch <- struct{}{}
			go performRequest(url, d, *cookie, &wg, ch, vulnerableURLs, result, &foundFlag)
		}
	}

	wg.Wait()

	// Close the result channel to signal the output writer goroutine to exit
	close(result)

	// If no vulnerable URLs were found, print a message
	if !foundFlag {
		fmt.Println("No vulnerable URLs found.")
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}
