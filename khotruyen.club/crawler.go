package khotruyenclub

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/golang-module/carbon/v2"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func convertToTimestamp(text string) int64 {
	// Tạo thời gian hiện tại
	now := carbon.Now()

	// Xử lý chuỗi text
	if strings.Contains(text, "phút trước") {
		minutes := extractNumber(text)
		return now.SubMinutes(minutes).Timestamp()
	} else if strings.Contains(text, "giờ trước") {
		hours := extractNumber(text)
		return now.SubHours(hours).Timestamp()
	} else if strings.Contains(text, "ngày trước") {
		days := extractNumber(text)
		return now.SubDays(days).Timestamp()
	}

	return now.Timestamp()
}

// Hàm này giúp extract số từ chuỗi text
func extractNumber(text string) int {
	var number int
	fmt.Sscanf(text, "%d", &number)
	return number
}

// // Chapter struct
// type Chapter struct {
// 	Name      string
// 	URL       string
// 	ImageURLs []string
// }

// // Comic struct
// type Comic struct {
// 	Name       string
// 	URL        string
// 	Chapters   []*Chapter
// 	Tags       []string
// 	Categories []string
// }

type Spider struct {
}

func Crawl() {
	// Initialize the main collector
	c := colly.NewCollector(
		colly.AllowedDomains("khotruyen.club"),
		colly.Async(true),
	)

	// Initialize the collector for chapters and images
	chapterCollector := c.Clone()
	imageCollector := c.Clone()

	// Set a rate limit to avoid overwhelming the server
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*khotruyen.club*",
		Parallelism: 1,
		Delay:       1 * time.Millisecond,
	})

	// Set a rate limit for the chapter collector
	chapterCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*khotruyen.club*",
		Parallelism: 1000,
		Delay:       10 * time.Millisecond,
	})

	// Set a rate limit for the image collector
	imageCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*khotruyen.club*",
		Parallelism: 1000,
		Delay:       5 * time.Millisecond,
	})

	// Initialize a map to store comics and their chapters
	comics := make(map[string]*Comic)

	// Define the callback for the main collector
	c.OnHTML(".page-item-detail.manga", func(e *colly.HTMLElement) {
		title := e.ChildText(".post-title.font-title")
		url := e.ChildAttr(".post-title.font-title a", "href")

		// Initialize a new Comic struct and store it in the map
		comic := &Comic{
			Title:     title,
			SourceURL: url,
		}
		if len(comics) == 1 {
			return
		}
		comics[url] = comic

		log.Print("Cralwer commic ", title)
		// Visit the comic URL to get the chapters
		chapterCollector.Visit(e.Request.AbsoluteURL(url))
	})

	// Define the callback for the chapter collector
	chapterCollector.OnHTML(".main.version-chap.no-volumn", func(e *colly.HTMLElement) {
		log.Print(111)
		// Get the comic URL from the referrer
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]

		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}

		e.ForEach("li.wp-manga-chapter", func(i int, h *colly.HTMLElement) {
			chapterName := h.ChildText("a")
			chapterURL := h.ChildAttr("a", "href")
			releaseDate := h.ChildText(".chapter-release-date")
			created := carbon.ParseByFormat(releaseDate, "d/m/Y", carbon.Saigon).StartOfDay().Timestamp()

			if created == 0 {
				created = convertToTimestamp(releaseDate)
			}
			// Initialize a new Chapter struct and append it to the comic's Chapters slice
			chapter := &Chapter{
				Title:     chapterName,
				SourceURL: chapterURL,
				CreatedAt: created,
			}
			log.Print("+ chapter ", chapterName)

			comic.Chapters = append(comic.Chapters, chapter)

			// Visit the chapter URL to get the images
			imageCollector.Visit(e.Request.AbsoluteURL(chapterURL))
		})
	})

	chapterCollector.OnHTML(".description-summary .summary__content p", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}
		log.Print(222)

		comic.Description = e.Text
	})

	chapterCollector.OnHTML(".summary-content .artist-content", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}
		log.Print(333)
		comic.Artist = e.ChildText("a")
	})
	chapterCollector.OnHTML(".post-content_item .summary-content", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}
		log.Print(4444)

		comic.Status = strings.TrimLeft(e.Text, "\n")
	})
	// Extract categories and tags
	chapterCollector.OnHTML(".tab-summary .summary_image ", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}
		log.Print(5555)

		comic.CoverImageURL = e.ChildAttr("img.img-responsive", "data-src")
	})

	chapterCollector.OnHTML(".tags-content", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}

		tags := strings.Split(e.Text, ",")
		tagItems := make([]*Tag, 0)
		log.Print(66666)

		for _, tag := range tags {
			tagItems = append(tagItems, &Tag{Title: strings.TrimSpace(tag)})
		}
		comic.Tags = tagItems
	})

	// Extract categories and tags
	chapterCollector.OnHTML(".genres-content", func(e *colly.HTMLElement) {
		comicURL := e.Request.URL.String()
		comic, exists := comics[comicURL]
		if !exists {
			log.Println("Comic not found for URL:", comicURL)
			return
		}

		categories := strings.Split(e.Text, ",")
		cateItems := make([]*Category, 0)
		for _, category := range categories {
			// categories[i] = strings.TrimSpace(category)
			cateItems = append(cateItems, &Category{Title: strings.TrimSpace(category)})
		}
		comic.Categories = cateItems
	})

	// Define the callback for the image collector
	imageCollector.OnHTML(".page-break.no-gaps", func(e *colly.HTMLElement) {

		// Get the chapter URL from the referrer
		chapterURL := e.Request.URL.String()

		e.ForEach("img", func(i int, h *colly.HTMLElement) {
		lbComic:
			for _, comic := range comics {
				for _, chapter := range comic.Chapters {
					if chapter.SourceURL == chapterURL {
						if len(chapter.ImageURLsArr) == 0 {
							chapter.ImageURLsArr = make([]string, 0)
						}
						chapter.ImageURLsArr = append(chapter.ImageURLsArr, strings.TrimLeft(h.Attr("data-src"), "\t\n\t\t\t"))
						break lbComic
					}
				}
			}
		})
		// Find the corresponding comic and chapter

	})

	page := 1
	maxPages := 1 // Set a limit for the number of pages to crawl

	for page <= maxPages {
		// Define the POST request
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("accept", "*/*")
			r.Headers.Set("accept-language", "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3")
			r.Headers.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
			r.Headers.Set("cookie", "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0")
			r.Headers.Set("origin", "https://khotruyen.club")
			r.Headers.Set("priority", "u=1, i")
			r.Headers.Set("referer", "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga")
			r.Headers.Set("sec-ch-ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
			r.Headers.Set("sec-ch-ua-mobile", "?0")
			r.Headers.Set("sec-ch-ua-platform", `"macOS"`)
			r.Headers.Set("sec-fetch-dest", "empty")
			r.Headers.Set("sec-fetch-mode", "cors")
			r.Headers.Set("sec-fetch-site", "same-origin")
			r.Headers.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
			r.Headers.Set("x-requested-with", "XMLHttpRequest")
		})

		// Send the POST request
		err := c.Post("https://khotruyen.club/wp-admin/admin-ajax.php", map[string]string{
			"action":                           "madara_load_more",
			"page":                             strconv.Itoa(page),
			"template":                         "madara-core/content/content-archive",
			"vars[paged]":                      "0",
			"vars[orderby]":                    "date",
			"vars[template]":                   "archive",
			"vars[sidebar]":                    "full",
			"vars[post_type]":                  "wp-manga",
			"vars[post_status]":                "publish",
			"vars[meta_query][relation]":       "AND",
			"vars[manga_archives_item_layout]": "big_thumbnail",
		},
		)

		if err != nil {
			log.Println("Failed to make POST request:", err)
		}

		page++
		c.Wait() // Wait for all requests to finish before proceeding to the next page
	}

	// Start the collector
	c.Visit("https://khotruyen.club/truyen-tranh/")
	c.Wait() // Wait for all requests to complete
	chapterCollector.Wait()
	imageCollector.Wait()
	// Print the collected data

	data, _ := json.MarshalIndent(comics, " ", " ")
	os.WriteFile("khotruyen.json", data, 0777)
	// for _, comic := range comics {
	// 	fmt.Printf("Comic: %s, URL: %s\n", comic.Name, comic.URL)
	// 	fmt.Printf("  Categories: %v\n", comic.Categories)
	// 	fmt.Printf("  Tags: %v\n", comic.Tags)
	// 	for _, chapter := range comic.Chapters {
	// 		fmt.Printf("    Chapter: %s, URL: %s\n", chapter.Name, chapter.URL)
	// 		fmt.Printf("    Images: %v\n", chapter.ImageURLs)
	// 	}
	// }
}

type FnChapter func(*Chapter) *Chapter
type FnComic func(*Comic) *Comic

// func CrawlToDb(cComic chan *Comic, cChapter chan *Chapter, findComic FnComic, findChapter FnChapter) {
// 	// Initialize the main collector
// 	c := colly.NewCollector(
// 		colly.AllowedDomains("khotruyen.club"),
// 		colly.Async(true),
// 	)

// 	// Initialize the collector for chapters and images
// 	chapterCollector := c.Clone()
// 	imageCollector := c.Clone()
// 	// Set a rate limit to avoid overwhelming the server
// 	c.Limit(&colly.LimitRule{
// 		DomainGlob:  "*khotruyen.club*",
// 		Parallelism: 1,
// 		Delay:       1 * time.Millisecond,
// 	})

// 	// Set a rate limit for the chapter collector
// 	chapterCollector.Limit(&colly.LimitRule{
// 		DomainGlob:  "*khotruyen.club*",
// 		Parallelism: 1000,
// 		Delay:       10 * time.Millisecond,
// 	})

// 	// Set a rate limit for the image collector
// 	imageCollector.Limit(&colly.LimitRule{
// 		DomainGlob:  "*khotruyen.club*",
// 		Parallelism: 1000,
// 		Delay:       5 * time.Millisecond,
// 	})

// 	// Initialize a map to store comics and their chapters
// 	// comics := make(map[string]*Comic)

// 	// Define the callback for the main collector
// 	c.OnHTML(".page-item-detail.manga", func(e *colly.HTMLElement) {
// 		title := e.ChildText(".post-title.font-title")
// 		url := e.ChildAttr(".post-title.font-title a", "href")

// 		// Initialize a new Comic struct and store it in the map
// 		comic := &Comic{
// 			Title:     title,
// 			SourceURL: url,
// 		}
// 		// if len(comics) == 1 {
// 		// 	return
// 		// }
// 		// comics[url] = comic
// 		cComic <- comic

// 		log.Print("Cralwer commic ", title)
// 		// Visit the comic URL to get the chapters
// 		chapterCollector.Visit(e.Request.AbsoluteURL(url))
// 	})

// 	// Define the callback for the chapter collector
// 	chapterCollector.OnHTML(".main.version-chap.no-volumn", func(e *colly.HTMLElement) {
// 		// Get the comic URL from the referrer
// 		comicURL := e.Request.URL.String()
// 		// comic, exists := comics[comicURL]
// 		comic := findComic(&Comic{SourceURL: comicURL})
// 		if comic == nil {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}

// 		e.ForEach("li.wp-manga-chapter", func(i int, h *colly.HTMLElement) {
// 			chapterName := h.ChildText("a")
// 			chapterURL := h.ChildAttr("a", "href")
// 			releaseDate := h.ChildText(".chapter-release-date")
// 			created := carbon.ParseByFormat(releaseDate, "d/m/Y", carbon.Saigon).StartOfDay().Timestamp()

// 			if created == 0 {
// 				created = convertToTimestamp(releaseDate)
// 			}
// 			// Initialize a new Chapter struct and append it to the comic's Chapters slice
// 			chapter := &Chapter{
// 				Title:     chapterName,
// 				SourceURL: chapterURL,
// 				CreatedAt: created,
// 			}
// 			log.Print("+ chapter ", chapterName)

// 			comic.Chapters = append(comic.Chapters, chapter)

// 			// Visit the chapter URL to get the images
// 			imageCollector.Visit(e.Request.AbsoluteURL(chapterURL))
// 		})
// 	})

// 	chapterCollector.OnHTML(".description-summary .summary__content p", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		// comic, exists := comics[comicURL]
// 		// if !exists {
// 		// 	log.Println("Comic not found for URL:", comicURL)
// 		// 	return
// 		// }

// 		comic := findComic(&Comic{SourceURL: comicURL})
// 		if comic == nil {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}

// 		comic.Description = e.Text
// 	})

// 	chapterCollector.OnHTML(".summary-content .artist-content", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		log.Print(e)
// 		comic, exists := comics[comicURL]
// 		if !exists {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}
// 		comic.Artist = e.ChildText("a")
// 	})
// 	chapterCollector.OnHTML(".post-content_item .summary-content", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		log.Print(e)
// 		comic, exists := comics[comicURL]
// 		if !exists {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}
// 		comic.Status = strings.TrimLeft(e.Text, "\n")
// 	})
// 	// Extract categories and tags
// 	chapterCollector.OnHTML(".tab-summary .summary_image ", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		comic, exists := comics[comicURL]
// 		if !exists {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}
// 		comic.CoverImageURL = e.ChildAttr("img.img-responsive", "data-src")
// 	})

// 	chapterCollector.OnHTML(".tags-content", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		comic, exists := comics[comicURL]
// 		if !exists {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}

// 		tags := strings.Split(e.Text, ",")
// 		tagItems := make([]*Tag, 0)

// 		for _, tag := range tags {
// 			tagItems = append(tagItems, &Tag{Name: strings.TrimSpace(tag)})
// 		}
// 		comic.Tags = tagItems
// 	})

// 	// Extract categories and tags
// 	chapterCollector.OnHTML(".genres-content", func(e *colly.HTMLElement) {
// 		comicURL := e.Request.URL.String()
// 		comic, exists := comics[comicURL]
// 		if !exists {
// 			log.Println("Comic not found for URL:", comicURL)
// 			return
// 		}

// 		categories := strings.Split(e.Text, ",")
// 		cateItems := make([]*Category, 0)
// 		for _, category := range categories {
// 			// categories[i] = strings.TrimSpace(category)
// 			cateItems = append(cateItems, &Category{Name: strings.TrimSpace(category)})
// 		}
// 		comic.Categories = cateItems
// 	})

// 	// Define the callback for the image collector
// 	imageCollector.OnHTML(".page-break.no-gaps", func(e *colly.HTMLElement) {

// 		// Get the chapter URL from the referrer
// 		chapterURL := e.Request.URL.String()

// 		e.ForEach("img", func(i int, h *colly.HTMLElement) {
// 		lbComic:
// 			for _, comic := range comics {
// 				for _, chapter := range comic.Chapters {
// 					if chapter.SourceURL == chapterURL {
// 						if len(chapter.ImageURLsArr) == 0 {
// 							chapter.ImageURLsArr = make([]string, 0)
// 						}
// 						chapter.ImageURLsArr = append(chapter.ImageURLsArr, strings.TrimLeft(h.Attr("data-src"), "\t\n\t\t\t"))
// 						break lbComic
// 					}
// 				}
// 			}
// 		})
// 		// Find the corresponding comic and chapter

// 	})

// 	page := 1
// 	maxPages := 1 // Set a limit for the number of pages to crawl

// 	for page <= maxPages {
// 		// Define the POST request
// 		c.OnRequest(func(r *colly.Request) {
// 			r.Headers.Set("accept", "*/*")
// 			r.Headers.Set("accept-language", "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3")
// 			r.Headers.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
// 			r.Headers.Set("cookie", "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0")
// 			r.Headers.Set("origin", "https://khotruyen.club")
// 			r.Headers.Set("priority", "u=1, i")
// 			r.Headers.Set("referer", "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga")
// 			r.Headers.Set("sec-ch-ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
// 			r.Headers.Set("sec-ch-ua-mobile", "?0")
// 			r.Headers.Set("sec-ch-ua-platform", `"macOS"`)
// 			r.Headers.Set("sec-fetch-dest", "empty")
// 			r.Headers.Set("sec-fetch-mode", "cors")
// 			r.Headers.Set("sec-fetch-site", "same-origin")
// 			r.Headers.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
// 			r.Headers.Set("x-requested-with", "XMLHttpRequest")
// 		})

// 		// Send the POST request
// 		err := c.Post("https://khotruyen.club/wp-admin/admin-ajax.php", map[string]string{
// 			"action":                           "madara_load_more",
// 			"page":                             strconv.Itoa(page),
// 			"template":                         "madara-core/content/content-archive",
// 			"vars[paged]":                      "0",
// 			"vars[orderby]":                    "date",
// 			"vars[template]":                   "archive",
// 			"vars[sidebar]":                    "full",
// 			"vars[post_type]":                  "wp-manga",
// 			"vars[post_status]":                "publish",
// 			"vars[meta_query][relation]":       "AND",
// 			"vars[manga_archives_item_layout]": "big_thumbnail",
// 		},
// 		)

// 		if err != nil {
// 			log.Println("Failed to make POST request:", err)
// 		}

// 		page++
// 		c.Wait() // Wait for all requests to finish before proceeding to the next page
// 	}

// 	// Start the collector
// 	c.Visit("https://khotruyen.club/truyen-tranh/")
// 	c.Wait() // Wait for all requests to complete
// 	chapterCollector.Wait()
// 	imageCollector.Wait()
// 	// Print the collected data

// 	data, _ := json.MarshalIndent(comics, " ", " ")
// 	os.WriteFile("khotruyen.json", data, 0777)
// 	// for _, comic := range comics {
// 	// 	fmt.Printf("Comic: %s, URL: %s\n", comic.Name, comic.URL)
// 	// 	fmt.Printf("  Categories: %v\n", comic.Categories)
// 	// 	fmt.Printf("  Tags: %v\n", comic.Tags)
// 	// 	for _, chapter := range comic.Chapters {
// 	// 		fmt.Printf("    Chapter: %s, URL: %s\n", chapter.Name, chapter.URL)
// 	// 		fmt.Printf("    Images: %v\n", chapter.ImageURLs)
// 	// 	}
// 	// }
// }
