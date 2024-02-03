package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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
	urlsFile := flag.String("u", "", "Archivo de texto con las URLs a las que se les realizará la petición GET.")
	dataFile := flag.String("d", "", "Archivo de texto con los datos que se agregarán a las URLs.")
	cookie := flag.String("c", "", "Cookie a incluir en la petición GET.")
	flag.Parse()

	if *urlsFile == "" || *dataFile == "" {
		fmt.Println("Debe proporcionar archivos de URLs y datos.")
		flag.PrintDefaults()
		return
	}

	urlsContent, err := ioutil.ReadFile(*urlsFile)
	if err != nil {
		fmt.Println("Error leyendo el archivo de URLs:", err)
		return
	}
	urls := string(urlsContent)
	urlList := splitLines(urls)

	dataContent, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		fmt.Println("Error leyendo el archivo de datos:", err)
		return
	}
	data := string(dataContent)
	dataList := splitLines(data)

	for _, url := range urlList {
		for _, d := range dataList {
			success, urlWithData, responseTime, errorMessage := performRequest(url, d, *cookie)

			if success && responseTime <= 20 {
				fmt.Printf("\033[1;32mURL %s - %.2f segundos\033[0m\n", urlWithData, responseTime)
			} else {
				fmt.Printf("\033[1;31mURL %s - %.2f segundos - Error: %v\033[0m\n", urlWithData, responseTime, errorMessage)
			}
		}
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}
