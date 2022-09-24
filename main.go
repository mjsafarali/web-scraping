package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"log"
	"strings"
)

type Course struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Year        string `json:"year"`
	Consumption string `json:"consumption"`
	Description string `json:"description"`
	Price       string `json:"price"`
}

func main() {
	url := "https://divar.ir/s/tehran/car"

	c := colly.NewCollector(
		colly.AllowedDomains("divar.ir", "www.divar.ir"),
		colly.CacheDir("./divar_cache"),
		colly.MaxDepth(2),
		colly.Debugger(&debug.LogDebugger{}),
	)

	// Create another collector to scrape course details
	detailCollector := c.Clone()
	ads := make([]Course, 0, 200)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if !strings.HasPrefix(link, "/v") || strings.Index(link, "/report") > -1 || strings.Index(link, "/feedback") > -1 {
			return
		}
		err := e.Request.Visit(link)
		if err != nil {
			return
		}
	})

	detailCollector.OnHTML("div[id=app]", func(e *colly.HTMLElement) {
		log.Println("Course found")
		title := e.ChildText(".kt-page-title__title.kt-page-title__title--responsive-sized")

		ad := Course{
			Title: title,
			URL:   e.Request.URL.String(),
		}

		// Iterate over div components and add details to ad
		e.ForEach("div.kt-group-row", func(_ int, el *colly.HTMLElement) {
			//consumption
			consumption := el.ChildText("div.kt-group-row-item:nth-child(1)>span.kt-group-row-item__value")
			ad.Consumption = clean([]byte(faNumToEn(consumption)))

			//year
			year := el.ChildText("div.kt-group-row-item:nth-child(2)>span.kt-group-row-item__value")
			ad.Year = clean([]byte(faNumToEn(year)))
		})

		//price
		price := e.ChildText("div.kt-container > div > div.kt-col-5 > div:nth-child(6) > div:nth-child(15) > div.kt-base-row__end.kt-unexpandable-row__value-box > p")
		// --remove toman word
		price = strings.ReplaceAll(price, "تومان", "")
		ad.Price = clean([]byte(faNumToEn(price)))

		//description
		description := faNumToEn(e.ChildText(".kt-description-row__text"))
		ad.Description = description

		//append to ads
		if ad.URL != url {
			ads = append(ads, ad)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	err := c.Visit(url)
	if err != nil {
		return
	}

	// Convert results to JSON data if the scraping job has finished
	jsonData, err := json.MarshalIndent(ads, "", "  ")
	if err != nil {
		panic(err)
	}

	// Dump json to the standard output (can be redirected to a file)
	fmt.Println(string(jsonData))
}

func faNumToEn(str string) (enNum string) {
	numbers := map[string]string{"۱": "1", "۲": "2", "۳": "3", "۴": "4", "۵": "5", "۶": "6", "۷": "7", "۸": "8", "۹": "9", "۰": "0"}

	for _, c := range str {
		if numbers[string(c)] != "" {
			enNum = enNum + numbers[string(c)]
		} else {
			enNum = enNum + string(c)
		}
	}
	return
}

func clean(s []byte) string {
	j := 0
	for _, c := range s {
		if ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') {
			s[j] = c
			j++
		}
	}
	return string(s[:j])
}
