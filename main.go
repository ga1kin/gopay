package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	URI   = "https://npfb.ru/grafik-vyplaty-pensii.php"
	AGENT = "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/110.0"
)

var (
	usage = `Укажите команду для выполнения:
	
	Московский
	Приволжский
	Восточно-Сибирский
	Северный
	Горьковский
	Юго-Восточный
	Свердловский
	Южно-Уральский
	Северо-Кавказский
	Забайкальский
	Красноярский
	Западно-Сибирский
	Куйбышевский
	Дальневосточный
	Октябрьский
	Калининградское
	`
)

func main() {
	client := httpClient()

	req, err := makeRequest(URI, AGENT)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := fetchHTML(client, req)
	if err != nil {
		log.Fatal(err)
	}

	td, err := getData(doc)
	if err != nil {
		log.Fatal(err)
	}

	text := extractText(td)
	fmt.Println(text)
}

// httpClient configure custom HTTP client
func httpClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}

	return client
}

// makeRequest create and modify HTTP request before sending
func makeRequest(url, userAgent string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

// fetchHTML fetch the provided request and return the response body
func fetchHTML(client *http.Client, req *http.Request) (io.ReadCloser, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code was %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// getData returns text contained in td tags as slice
func getData(body io.ReadCloser) ([]string, error) {
	tokenizer := html.NewTokenizer(body)
	defer body.Close()

	var data []string

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error tokenizing HTML: %w", tokenizer.Err())
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "td" {
				tokenType = tokenizer.Next()
				if tokenType == html.TextToken {
					token := tokenizer.Token()
					data = append(data, strings.TrimSpace(token.String()))
				}
			}
		}
	}

	return data, nil
}

func extractText(data []string) string {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	command := flag.Args()[0]

	var txt string
	for i, d := range data {
		if strings.Contains(d, command) {
			lst := data[i : i+2]
			txt = strings.Join(lst, " ")
		}
	}

	return txt
}
