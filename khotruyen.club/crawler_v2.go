package khotruyenclub

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang-module/carbon/v2"
	"github.com/teng231/gotools/v2/httpclient"
)

func trim(txt string) string {
	return strings.ReplaceAll(strings.ReplaceAll(txt, "\n", ""), "\t", "")
}
func FetchChapterImgs(chapter *Chapter) error {
	resp, err := httpclient.Exec(chapter.SourceURL,
		httpclient.WithTransport(100, 100, 100),
		httpclient.WithHeader(map[string]string{
			"accept":             "*/*",
			"accept-language":    "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3",
			"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
			"cookie":             "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0",
			"origin":             "https://khotruyen.club",
			"priority":           "u=1, i",
			"referer":            "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga",
			"sec-ch-ua":          `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"macOS"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-origin",
			"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"x-requested-with":   "XMLHttpRequest",
		}),
		httpclient.WithTimeout(20*time.Second),
	)
	if err != nil {
		return err
	}
	if resp.HttpCode != 200 {
		return fmt.Errorf("status code error: %d", resp.HttpCode)
	}
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	sector := doc.Find(".page-break.no-gaps img")
	if sector.Length() > 0 {
		wg.Add(sector.Length())
	}
	sector.Each(func(i int, s *goquery.Selection) {
		imageURL, _ := s.Attr("data-src")
		chapter.ImageURLsArr = append(chapter.ImageURLsArr, trim(imageURL))
		wg.Done()
	})
	wg.Wait()
	return nil
}

func FetchComicDetailChapter(comic *Comic) error {
	resp, err := httpclient.Exec(comic.SourceURL,
		httpclient.WithTransport(100, 100, 100),
		httpclient.WithHeader(map[string]string{
			"accept":             "*/*",
			"accept-language":    "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3",
			"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
			"cookie":             "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0",
			"origin":             "https://khotruyen.club",
			"priority":           "u=1, i",
			"referer":            "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga",
			"sec-ch-ua":          `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"macOS"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-origin",
			"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"x-requested-with":   "XMLHttpRequest",
		}),
		httpclient.WithTimeout(20*time.Second),
	)
	if err != nil {
		return err
	}
	if resp.HttpCode != 200 {
		return fmt.Errorf("status code error: %d", resp.HttpCode)
	}
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	// wg.Add(7)
	sector1 := doc.Find(".main.version-chap.no-volumn li.wp-manga-chapter")
	if sector1.Length() > 0 {
		wg.Add(sector1.Length())
	}
	comic.TotalChapter = int32(sector1.Length())
	sector1.Each(func(i int, s *goquery.Selection) {
		chapterName := s.Find("a").Text()
		chapterURL, _ := s.Find("a").Attr("href")
		releaseDate := s.Find(".chapter-release-date").Text()
		created := carbon.ParseByFormat(trim(releaseDate), "d/m/Y", carbon.Saigon).EndOfDay().Timestamp()

		if created == 0 {
			created = convertToTimestamp(trim(releaseDate))
		}

		sourceURL := trim(chapterURL)
		shortcut := ""
		if x := strings.Split(sourceURL, "/"); len(x) >= 6 {
			shortcut = x[5]
		}
		strings.Split(sourceURL, "/")
		chapter := &Chapter{
			Title:     trim(chapterName),
			SourceURL: sourceURL,
			CreatedAt: created,
			Shortcut:  shortcut,
		}
		// log.Print("+ chapter ", chapterName)

		comic.Chapters = append(comic.Chapters, chapter)

		// imageDoc, err := fetchDocument(chapterURL)
		// if err != nil {
		// 	log.Println("Error fetching image document:", err)
		// 	return
		// }

		// imageDoc.Find(".page-break.no-gaps img").Each(func(i int, s *goquery.Selection) {
		// 	imageURL, _ := s.Attr("data-src")
		// 	chapter.ImageURLsArr = append(chapter.ImageURLsArr, strings.TrimLeft(imageURL, "\t\n\t\t\t"))
		// })

		wg.Done()
	})
	sector2 := doc.Find(".description-summary .summary__content p")
	if sector2.Length() > 0 {
		wg.Add(1)
	}
	sector2.Each(func(i int, s *goquery.Selection) {
		comic.Description = trim(s.Text())
		wg.Done()
	})

	sector3 := doc.Find(".summary-content .artist-content a")
	if sector3.Length() > 0 {
		wg.Add(1)
	}
	sector3.Each(func(i int, s *goquery.Selection) {
		comic.Artist = trim(s.Text())
		wg.Done()
	})
	sector4 := doc.Find(".post-content_item .summary-content")
	if sector4.Length() > 0 {
		wg.Add(sector4.Length())
	}
	sector4.Each(func(i int, s *goquery.Selection) {
		comic.Status = strings.TrimSpace(s.Text())
		wg.Done()
	})
	sector5 := doc.Find(".tab-summary .summary_image img.img-responsive")
	if sector5.Length() > 0 {
		wg.Add(1)
	}
	sector5.Each(func(i int, s *goquery.Selection) {
		comic.CoverImageURL, _ = s.Attr("data-src")
		wg.Done()
	})
	sector6 := doc.Find(".tags-content")
	if sector6.Length() > 0 {
		wg.Add(1)
	}

	sector6.Each(func(i int, s *goquery.Selection) {
		tags := strings.Split(s.Text(), ",")
		tagItems := make([]*Tag, 0)
		for _, tag := range tags {
			tagItems = append(tagItems, &Tag{Title: trim(tag)})
		}
		comic.Tags = tagItems
		wg.Done()
	})
	sector7 := doc.Find(".genres-content")
	if sector7.Length() > 0 {
		wg.Add(1)
	}
	sector7.Each(func(i int, s *goquery.Selection) {
		categories := strings.Split(s.Text(), ",")
		cateItems := make([]*Category, 0)
		for _, category := range categories {
			cateItems = append(cateItems, &Category{Title: trim(category)})
		}
		comic.Categories = cateItems
		wg.Done()
	})

	wg.Wait()
	return nil
}

// func FetchComicDetailChapterFix(comic *Comic) error {
// 	resp, err := httpclient.Exec(comic.SourceURL,
// 		httpclient.WithTransport(100, 100, 100),
// 		httpclient.WithHeader(map[string]string{
// 			"accept":             "*/*",
// 			"accept-language":    "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3",
// 			"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
// 			"cookie":             "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0",
// 			"origin":             "https://khotruyen.club",
// 			"priority":           "u=1, i",
// 			"referer":            "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga",
// 			"sec-ch-ua":          `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`,
// 			"sec-ch-ua-mobile":   "?0",
// 			"sec-ch-ua-platform": `"macOS"`,
// 			"sec-fetch-dest":     "empty",
// 			"sec-fetch-mode":     "cors",
// 			"sec-fetch-site":     "same-origin",
// 			"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
// 			"x-requested-with":   "XMLHttpRequest",
// 		}),
// 		httpclient.WithTimeout(20*time.Second),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	if resp.HttpCode != 200 {
// 		return fmt.Errorf("status code error: %d", resp.HttpCode)
// 	}
// 	// Parse the HTML document
// 	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
// 	if err != nil {
// 		return err
// 	}

// 	wg := &sync.WaitGroup{}
// 	sector1 := doc.Find(".main.version-chap.no-volumn li.wp-manga-chapter")
// 	comic.TotalChapter = int32(sector1.Length())

// 	sector6 := doc.Find(".tags-content")
// 	if sector6.Length() > 0 {
// 		wg.Add(1)
// 	}

// 	sector6.Each(func(i int, s *goquery.Selection) {
// 		tags := strings.Split(s.Text(), ",")
// 		tagItems := make([]*Tag, 0)
// 		for _, tag := range tags {
// 			tagItems = append(tagItems, &Tag{Title: trim(tag)})
// 		}
// 		comic.Tags = tagItems
// 		wg.Done()
// 	})
// 	sector7 := doc.Find(".genres-content")
// 	if sector7.Length() > 0 {
// 		wg.Add(1)
// 	}
// 	sector7.Each(func(i int, s *goquery.Selection) {
// 		categories := strings.Split(s.Text(), ",")
// 		cateItems := make([]*Category, 0)
// 		for _, category := range categories {
// 			cateItems = append(cateItems, &Category{Title: trim(category)})
// 		}
// 		comic.Categories = cateItems
// 		wg.Done()
// 	})

// 	wg.Wait()
// 	return nil
// }

func FechComitPage(url string, page int) ([]*Comic, error) {
	resp, err := httpclient.Exec(url,
		httpclient.WithTransport(100, 100, 100),
		httpclient.WithHeader(map[string]string{
			"accept":             "*/*",
			"accept-language":    "vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3",
			"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
			"cookie":             "_ga=GA1.1.1703967820.1721562178; wordpress_test_cookie=WP%20Cookie%20check; wpmanga-reading-history=W3siaWQiOjYwMjgsImMiOiIzOTU4NCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2MDIwfSx7ImlkIjozMjcwNiwiYyI6IjYyNTgxOCIsInAiOjEsImkiOiIiLCJ0IjoxNzIxODU2ODk1fV0%3D; _ga_133RB8PD54=GS1.1.1721842086.14.1.1721842169.0.0.0",
			"origin":             "https://khotruyen.club",
			"priority":           "u=1, i",
			"referer":            "https://khotruyen.club/truyen-tranh/?m_orderby=new-manga",
			"sec-ch-ua":          `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"macOS"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "same-origin",
			"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"x-requested-with":   "XMLHttpRequest",
		}),
		httpclient.WithTimeout(20*time.Second),
		httpclient.WithMethod("POST"),
		httpclient.WithUrlEncode(map[string]string{
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
		}),
	)
	if err != nil {
		return nil, err
	}
	if resp.HttpCode != 200 {
		return nil, fmt.Errorf("status code error: %d", resp.HttpCode)
	}
	comics := make([]*Comic, 0)
	// Parse the HTML document
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		return nil, err
	}
	// numdoc := 0
	selector := doc.Find(".page-item-detail.manga")
	if selector.Length() == 0 {
		return nil, nil
	}

	wg := &sync.WaitGroup{}

	wg.Add(selector.Length())

	selector.Each(func(i int, s *goquery.Selection) {
		defer wg.Done()
		title := s.Find(".post-title.font-title").Text()
		url, _ := s.Find(".post-title.font-title a").Attr("href")

		comic := &Comic{
			Title:     trim(title),
			SourceURL: url,
		}
		// if len(comics) == 1 {
		// 	return
		// }
		comics = append(comics, comic)
		// log.Print("Crawled comic ", title)
	})
	wg.Wait()
	return comics, nil
}
