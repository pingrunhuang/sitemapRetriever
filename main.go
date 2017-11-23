package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var politicsXMLMocking = []byte(
	`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:n="http://www.google.com/schemas/sitemap-news/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd http://www.google.com/schemas/sitemap-news/0.9 http://www.google.com/schemas/sitemap-news/0.9/sitemap-news.xsd">
	<url>
	<loc>
	https://www.washingtonpost.com/politics/the-latest-roy-moore-thanks-trump-in-new-fundraising-appeal/2017/11/22/bd14fa82-cff2-11e7-a87b-47f14b73162a_story.html
	</loc>
	<changefreq>hourly</changefreq>
	<n:news>
	<n:publication>
	<n:name>Washington Post</n:name>
	<n:language>en</n:language>
	</n:publication>
	<n:publication_date>2017-11-23T02:05:12Z</n:publication_date>
	<n:title>
	The Latest: Roy Moore thanks Trump in new fundraising appeal
	</n:title>
	<n:keywords>
	US-Trump-Moore-The Latest,Trump Moore,United States Senate,U.S. Republican Party,United States Congress,United States government,Roy Moore,Donald Trump,Mitch McConnell,Dean Young,Alabama,United States,North America,District of Columbia,General news,Government and politics,Senate elections,Campaigns,Elections,election news,election news 2017,2017 election news,elections,election latest news,Political fundraising,Campaign finance,Legislature,Political corruption,Political issues,2016 United States presidential election,United States presidential election,2017 elections,election news,2017 elections news,elections,election news,presidential elections 2017,election news 2017,Special elections,Sexual misconduct,Sex in society,Social issues,Social affairs,Presidential elections,National elections,Liberalism
	</n:keywords>
	</n:news>
	</url>
	<url>
	<loc>
	https://www.washingtonpost.com/politics/courts_law/democrats-face-hot-potato-politics-of-sexual-predation-too/2017/11/22/f2e2991e-cfeb-11e7-a87b-47f14b73162a_story.html
	</loc>
	<changefreq>hourly</changefreq>
	<n:news>
	<n:publication>
	<n:name>Washington Post</n:name>
	<n:language>en</n:language>
	</n:publication>
	<n:publication_date>2017-11-23T01:53:31Z</n:publication_date>
	<n:title>
	Democrats face hot-potato politics of sexual predation, too
	</n:title>
	<n:keywords>
	US-Sexual Misconduct-Democrats,United States Congress,District of Columbia,United States,North America,United States Senate,United States government,U.S. Democratic Party,Anita Hill,Clarence Thomas,Bill Clinton,Al Franken,Roy Moore,Kirsten Gillibrand,Jackie Speier,Hillary Clinton,John Conyers,Donald Trump,Government and politics,Sexual assault,Violent crime,Crime,General news,Political parties,Political organizations,Political corruption,Political issues,Bills,Legislation,Legislature,Political scandals,Political ethics,Sexual misconduct,Sex in society,Social issues,Social affairs
	</n:keywords>
	</n:news>
	</url>
	<url>
	<loc>
	https://www.washingtonpost.com/politics/courts_law/sessions-orders-review-of-background-check-system-for-guns/2017/11/22/e6ed179e-cfe4-11e7-a87b-47f14b73162a_story.html
	</loc>
	<changefreq>hourly</changefreq>
	<n:news>
	<n:publication>
	<n:name>Washington Post</n:name>
	<n:language>en</n:language>
	</n:publication>
	<n:publication_date>2017-11-23T00:26:09Z</n:publication_date>
	<n:title>
	Sessions orders review of background check system for guns
	</n:title>
	<n:keywords>
	US-Church Shooting-Justice Department,United States government,Jeff Sessions,U.S. Department of Justice,U.S. Department of Defense,General news,Government and politics,Religious strife,Religious issues,Religion,Social affairs,Social issues,Texas church shooting,Shootings,Violent crime,Crime,Armed forces,Military and defense,Air force
	</n:keywords>
	</n:news>
	</url>
	<url>
	<loc>
	https://www.washingtonpost.com/local/virginia-politics/on-eve-of-certification-democrats-file-third-lawsuit-in-disputed-va-house-race/2017/11/21/f2875a02-ceda-11e7-81bc-c55a220c8cbe_story.html
	</loc>
	<changefreq>hourly</changefreq>
	<n:news>
	<n:publication>
	<n:name>Washington Post</n:name>
	<n:language>en</n:language>
	</n:publication>
	<n:publication_date>2017-11-23T00:03:00Z</n:publication_date>
	<n:title>
	Federal judge rejects Democrats’ request to block certification of Va. races but leaves door open for new election
	</n:title>
	<n:keywords>
	joshua cole, cole, virginia house of delegates, robert thomas, virginia house of delegates ballots, stafford county ballots, absentee ballots virginia,virginia elections
	</n:keywords>
	</n:news>
	</url>
	</urlset>`)

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

func (s SitemapIndex) printSitemap() {
	for _, loc := range s.Locations {
		log.Printf("Fetched url: %s", loc)
	}
}

func (n News) printNews() {
	fmt.Printf("Title:%s\n", n.Titles)
	fmt.Printf("Keywords:%s\n", n.Keywords)
	fmt.Printf("URL:%s\n", n.URLs)
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

var newsLink SitemapIndex
var newsDetail = []News{}
var newsDetailMock = []News{}

// this method could be imporved with concurrency
func fetchNews(isMocking bool) {
	var n News
	if isMocking {
		xml.Unmarshal(politicsXMLMocking, &n)
		newsDetailMock = append(newsDetailMock, News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
	} else {
		url := "https://www.washingtonpost.com/news-sitemap-index.xml"
		bodyByte, err := retrieveXML(url)
		if err != nil {
			log.Fatalf("Error retrieving xml from sitemap %s, %v", url, err)
		}
		// how to stringify the byte array data
		// stringBody := string(bodyByte)

		xml.Unmarshal(bodyByte, &newsLink)
		newsLink.printSitemap()

		for _, newsURL := range newsLink.Locations {
			bodyByte, err := retrieveXML(newsURL)
			if err != nil {
				log.Fatal(generateError(newsURL, err))
			}
			xml.Unmarshal(bodyByte, &n)
			newsDetail = append(newsDetail, News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
		}
	}
}

// improvement of the previous method with concurrency
var wg sync.WaitGroup

func handlePanic() {
	if r := recover(); r != nil {
		log.Fatal("Panic being recovered: ", r)
	}
}
func fetchNewsCon(isMocking bool) {

	var n News
	if isMocking {
		xml.Unmarshal(politicsXMLMocking, &n)
		newsDetailMock = append(newsDetailMock, News{URLs: n.URLs, Keywords: n.Keywords, Titles: n.Titles})
	} else {
		url := "https://www.washingtonpost.com/news-sitemap-index.xml"
		bodyByte, err := retrieveXML(url)
		if err != nil {
			log.Fatalf("Error retrieving xml from sitemap %s, %v", url, err)
		}
		xml.Unmarshal(bodyByte, &newsLink)
		newsChan := make(chan News, len(newsLink.Locations))
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
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/newsDetail", newsDetailHandler)
	http.ListenAndServe(":8000", nil)
	log.Println("Starting listening on port 8000")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// way to instantiate a struct object
	data := IndexPage{Title: "Daily news from Washinton Post", Links: newsLink.Locations}
	temp, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("%v", err)
	}
	temp.Execute(w, data)
}

func newsDetailHandler(w http.ResponseWriter, r *http.Request) {
	// data := DetailPage{Details: newsDetailMock}
	data := DetailPage{Details: newsDetail}
	temp, err := template.ParseFiles("template/newsDetail.html")
	if err != nil {
		log.Fatalf("%v", err)
	}
	temp.Execute(w, data)
}

func main() {
	router()
}
