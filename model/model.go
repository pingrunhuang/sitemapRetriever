package model

import (
	"fmt"
	"log"
)

// SitemapIndex is a customized sitemap object
type SitemapIndex struct {
	// how to access xml tree
	Locations []string `xml:"sitemap>loc"`
}

// News is used as an aggregation for holding items fetched from sitemap
type News struct {
	URLs []string `xml:"url>loc"`
	// the sub tag in the xml tree starting with 'n:', but somehow I can't access it with 'n:news' but 'news'
	Titles   []string `xml:"url>news>title"`
	Keywords []string `xml:"url>news>keywords"`
}

// NewsItem is an item of a news that show up in the sitemap
// different from the News, it is used to store the info of one item
type NewsItem struct {
	Title    string
	URL      string
	Keywords string
}

// IndexPage is used to store the data that should be rendered on the index page
type IndexPage struct {
	Title string
	Links []string
}

// DetailPage is used to store data passed to detail page
type DetailPage struct {
	Details []News
}

func (s SitemapIndex) PrintSitemap() {
	for _, loc := range s.Locations {
		log.Printf("Fetched url: %s", loc)
	}
}

func (n News) PrintNews() {
	fmt.Printf("Title:%s\n", n.Titles)
	fmt.Printf("Keywords:%s\n", n.Keywords)
	fmt.Printf("URL:%s\n", n.URLs)
	fmt.Println()
}
