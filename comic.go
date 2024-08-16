package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/teng231/back4app/db"
	khotruyenclub "github.com/teng231/back4app/khotruyen.club"
	"github.com/teng231/executor"
)

var (
	// limitCallComic        = &executor.Limiter{Limit: 10}
	limitCallFetchChapter = &executor.Limiter{Limit: 100}
)

func chaptersReverse(chapters []*khotruyenclub.Chapter) []*khotruyenclub.Chapter {
	for i := 0; i < len(chapters)/2; i++ {
		// Hoán đổi phần tử đầu với phần tử cuối
		j := len(chapters) - i - 1
		chapters[i], chapters[j] = chapters[j], chapters[i]
	}
	return chapters
}

func upsertCommic(dbc db.IComicDB, comic *khotruyenclub.Comic) error {
	comicData, err := dbc.FindComic(&khotruyenclub.Comic{Title: comic.Title})

	if err != nil {
		respCate, _ := dbc.UpsertCategory(comic.Categories)
		respTag, _ := dbc.UpsertTag(comic.Tags)

		tags := ""

		cates := ""

		for i, cate := range respCate {
			cates += strconv.Itoa(int(cate.ID))
			if i != len(respCate)-1 {
				cates += ","
			}
		}

		for i, tag := range respTag {
			tags += strconv.Itoa(int(tag.ID))
			if i != len(respTag)-1 {
				tags += ","
			}
		}
		comic.CategoryItems = cates
		comic.TagItems = tags
		comic.Type = "truyen_tranh"
		if err := dbc.InsertComic(comic); err != nil {
			return err
		}
		comicData = comic
	}

	chapters, err := dbc.ListChapters(&khotruyenclub.ChapterRequest{ComicID: comicData.ID, Limit: 1, Order: "id desc "})

	if err != nil {
		return nil
	}

	chaptersInserted := []*khotruyenclub.Chapter{}

	if len(chapters) == 0 {
		chaptersInserted = comic.Chapters
	} else {
		for _, chapter := range comic.Chapters {
			if chapter.CreatedAt > chapters[0].CreatedAt {
				chaptersInserted = append(chaptersInserted, chapter)
			}
		}
	}

	for _, chapter := range chaptersInserted {
		chapter.ComicID = comicData.ID
		chapter.ImageURLs = strings.Join(chapter.ImageURLsArr, ",")
	}
	if len(chaptersInserted) == 0 {
		return errors.New("no_inserted")
	}
	chaptersInserted = chaptersReverse(chaptersInserted)
	if err := dbc.InsertChapter(chaptersInserted...); err != nil {
		return err
	}
	lastChapter := chaptersInserted[0]

	dbc.UpdateComic(&khotruyenclub.Comic{LastChapter: lastChapter.ID, LastChapterUpdated: lastChapter.CreatedAt},
		&khotruyenclub.Comic{ID: comicData.ID})
	return nil
}

func MakeCrawlerKhotruyen(dbComic db.IComicDB, exec executor.ISafeQueue, page int) {
	// NOTE: crawler run
	// page := 1
	for {
		log.Printf("---------------- page %d ---------------------", page)
		comics, err := khotruyenclub.FechComitPage("https://khotruyen.club/wp-admin/admin-ajax.php", page)
		// data1, _ := json.MarshalIndent(comics, "", " ")
		// log.Print(string(data1), err)
		// com := comics[1]
		if err != nil {
			log.Print(err)
			page++
			continue
		}

		if len(comics) == 0 {
			break
		}
		log.Printf("comic found %d", len(comics))
		noInserted := 0
		for _, comic := range comics {

			now := time.Now()

			if err = khotruyenclub.FetchComicDetailChapter(comic); err != nil {
				log.Print(err)
			}

			shortcuts := strings.ReplaceAll(comic.SourceURL, "https://", "")
			// khotruyen.club/truyen-tranh/tuyen-tap-abo-ngan-cua-nha-sec-2/
			sc := strings.Split(shortcuts, "/")

			if len(sc) > 2 {
				comic.Shortcut = sc[2]
			}

			jobs := make([]*executor.Job, 0)
			for _, chapter := range comic.Chapters {
				jobs = append(jobs, &executor.Job{
					Params:  []any{chapter},
					Limiter: limitCallFetchChapter,
					Exectutor: func(i ...interface{}) (interface{}, error) {
						chapter := i[0].(*khotruyenclub.Chapter)
						if err := khotruyenclub.FetchChapterImgs(chapter); err != nil {
							log.Print(err)
						}
						return nil, nil
					},
				})
			}
			exec.SendWithGroup(jobs...)
			if err := upsertCommic(dbComic, comic); err != nil && err.Error() == "no_inserted" {
				noInserted++
			}
			log.Printf("pulling comic %s number chapter %d %v", comic.Title, len(comic.Chapters), time.Since(now))
		}
		if noInserted == len(comics) {
			log.Print("Update all doc.")
			break
		}
		page++
	}
}

// func Recrawler(dbComic db.IComicDB, exec executor.ISafeQueue, startWithDate int64) {
// 	// NOTE: crawler run
// 	cComic := make(chan *khotruyenclub.Comic, 1000)
// 	wg := &sync.WaitGroup{}

// 	for i := 0; i < 10; i++ {
// 		go func() {
// 			for {
// 				comic := <-cComic
// 				// lấy tổng số chapter
// 				// lấy tag, category
// 				if err := khotruyenclub.FetchComicDetailChapter(comic); err != nil {
// 					log.Print(err)
// 				}
// 				respCate, _ := dbComic.UpsertCategory(comic.Categories)
// 				respTag, _ := dbComic.UpsertTag(comic.Tags)

// 				tags := ""

// 				cates := ""

// 				for i, cate := range respCate {
// 					cates += strconv.Itoa(int(cate.ID))
// 					if i != len(comic.Categories)-1 {
// 						cates += ","
// 					}
// 				}

// 				for i, tag := range respTag {
// 					tags += strconv.Itoa(int(tag.ID))
// 					if i != len(comic.Categories)-1 {
// 						tags += ","
// 					}
// 				}
// 				comic.CategoryItems = cates
// 				comic.TagItems = tags
// 				if err := dbComic.UpdateComic(comic, &khotruyenclub.Comic{ID: comic.ID}); err != nil {
// 					log.Print(err)
// 				}
// 			}
// 		}()
// 	}

// 	dbComic.ScanComicTable(&khotruyenclub.ComicRequest{Cols: []string{"id", "source_url"}}, cComic, wg)
// 	wg.Wait()
// }
