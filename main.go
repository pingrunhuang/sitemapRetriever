package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"./mocking"
	"./model"
)

func retrieveXML(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("Error accessing %s, %v ", url, err)
	}
	bodyByte, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return bodyByte, err
}

var newsLink model.SitemapIndex
var newsDetail = []model.News{}
var newsDetailMock = []model.News{}

// improvement of the previous method with concurrency
var wg sync.WaitGroup

func handlePanic() {
	if r := recover(); r != nil {
		log.Fatal("Panic being recovered: ", r)
	}
}

// this method could be imporved with concurrency
func fetchNews(isMocking bool) {
	var n model.News
	if isMocking {
		xml.Unmarshal(mocking.PoliticsXMLMocking, &n)
		newsDetailMock = append(newsDetailMock, model.News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
	} else {
		url := "https://www.washingtonpost.com/news-sitemap-index.xml"
		bodyByte, err := retrieveXML(url)
		if err != nil {
			log.Fatalf("Error retrieving xml from sitemap %s, %v", url, err)
		}
		// how to stringify the byte array data
		// stringBody := string(bodyByte)

		xml.Unmarshal(bodyByte, &newsLink)
		newsLink.PrintSitemap()

		for _, newsURL := range newsLink.Locations {
			bodyByte, err := retrieveXML(newsURL)
			if err != nil {
				log.Fatalf("Error accessing %s, %v ", url, err)
			}
			xml.Unmarshal(bodyByte, &n)
			newsDetail = append(newsDetail, model.News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
		}
	}
}

func fetchNewsCon(isMocking bool) {
	var n model.News
	if isMocking {
		xml.Unmarshal(mocking.PoliticsXMLMocking, &n)
		newsDetailMock = append(newsDetailMock, model.News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
	} else {
		url := "https://www.washingtonpost.com/news-sitemap-index.xml"
		bodyByte, err := retrieveXML(url)
		if err != nil {
			log.Fatalf("Error retrieving xml from sitemap %s, %v", url, err)
		}
		xml.Unmarshal(bodyByte, &newsLink)
		newsChan := make(chan model.News, len(newsLink.Locations))
		for _, newsLoc := range newsLink.Locations {
			wg.Add(1)
			go func(newsLoc string) {
				defer wg.Done()
				defer handlePanic()
				bodyByte, err := retrieveXML(newsLoc)
				if err != nil {
					panic(fmt.Errorf("Error retrieving %s", newsLoc))
				}
				xml.Unmarshal(bodyByte, &n)
				newsChan <- n
			}(newsLoc)
		}
		wg.Wait()
		close(newsChan)
		for news := range newsChan {
			newsDetail = append(newsDetail, news)
		}
	}
}

func router() {
	// fetchNews(false)
	fetchNewsCon(false)
	log.Println("Starting listening on port 8000")
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/newsDetail", newsDetailHandler)
	http.ListenAndServe(":8000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// way to instantiate a struct object
	data := model.IndexPage{Title: "Daily news from Washinton Post", Links: newsLink.Locations}
	temp, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("%v", err)
	}
	temp.Execute(w, data)
}

func newsDetailHandler(w http.ResponseWriter, r *http.Request) {
	// data := DetailPage{Details: newsDetailMock}
	data := model.DetailPage{Details: newsDetail}
	temp, err := template.ParseFiles("template/newsDetail.html")
	if err != nil {
		log.Fatalf("%v", err)
	}
	temp.Execute(w, data)
}

func main() {
	router()
}
