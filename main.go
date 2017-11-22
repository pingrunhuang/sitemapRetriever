package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// func xml_retriever(url string) ([]byte, error){

// }

// SitemapIndex is a customized sitemap object
type SitemapIndex struct {
	// how to access xml tree
	Locations []string `xml:"sitemap>loc"`
}

func (s SitemapIndex) printSitemap() {
	for _, loc := range s.Locations {
		fmt.Println(loc)
	}
}

// News is used as an aggregate for the each item from sitemap
type News struct {
	Title    []string `xml:"url>n:news>n:title"`
	URL      []string `xml:"url>loc"`
	Keywords []string `xml:"url>n:news>Keywords"`
}

func (n News) printNews() {
	fmt.Printf("Title:%s\n", n.Title)
	fmt.Printf("Keywords:%s\n", n.Keywords)
	fmt.Printf("URL:%s\n", n.URL)
	fmt.Println()
}

func generateError(url string, err error) error {
	return fmt.Errorf("Error accessing %s, %v ", url, err)
}

func retrieveXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, generateError(url, err)
	}
	bodyByte, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return bodyByte, err
}

type newsDetailPage struct {
	Title string
	Links []string
}

var newsLink SitemapIndex
var newsDetail = make(map[string]News)

func fetchNews() {
	url := "https://www.washingtonpost.com/news-sitemap-index.xml"
	bodyByte, err := retrieveXML(url)
	if err != nil {
		log.Fatalf("Error retrieving xml from sitemap %s, %v", url, err)
	}
	// how to stringify the byte array data
	// stringBody := string(bodyByte)

	xml.Unmarshal(bodyByte, &newsLink)

	var n News
	for _, u := range newsLink.Locations {
		bodyByte, err := retrieveXML(u)
		if err != nil {
			log.Fatal(generateError(u, err))
		}
		xml.Unmarshal(bodyByte, &n)
		newsDetail[u] = News{URL: n.URL, Keywords: n.Keywords, Title: n.Title}
	}
}

func router() {
	fetchNews()
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// way to instantiate a struct object
	data := newsDetailPage{Title: "Daily news from Washinton Post", Links: newsLink.Locations}
	temp, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("%v", err)
	}
	temp.Execute(w, data)
}

func main() {
	router()
}
