package khotruyenclub

import (
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/teng231/executor"
)

func TestCrawlerV2(t *testing.T) {
	comics, err := FechComitPage("https://khotruyen.club/wp-admin/admin-ajax.php", 1)
	data1, _ := json.MarshalIndent(comics, "", " ")
	log.Print(string(data1), err)

	com := comics[1]

	if err = FetchComicDetailChapter(com); err != nil {
		log.Print(err)
	}

	// data2, _ := json.MarshalIndent(com, "", " ")
	// log.Print(string(data2), err)
	for _, chapter := range com.Chapters {
		if err := FetchChapterImgs(chapter); err != nil {
			log.Print(err)
		}
	}

	data, _ := json.MarshalIndent(com, "", " ")
	log.Print(string(data), err)
}

func TestCrawlerTool(t *testing.T) {
	page := 1
	exec := executor.RunSafeQueue(&executor.SafeQueueConfig{NumberWorkers: 200, Capacity: 4000})
	for {
		log.Printf("---------------- page %d ---------------------", page)
		comics, err := FechComitPage("https://khotruyen.club/wp-admin/admin-ajax.php", page)
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
		jobs := make([]*executor.Job, 0)
		for _, comic := range comics {
			jobs = append(jobs, &executor.Job{
				Params: []any{comic},
				Exectutor: func(i ...interface{}) (interface{}, error) {
					comic := i[0].(*Comic)

					if err = FetchComicDetailChapter(comic); err != nil {
						log.Print(err)
					}
					for _, chapter := range comic.Chapters {
						if err := FetchChapterImgs(chapter); err != nil {
							log.Print(err)
						}
					}
					log.Print("pulling comic ", comic.Title)
					return nil, nil
				},
			})
		}
		exec.SendWithGroup(jobs...)
		data1, _ := json.MarshalIndent(comics, "", " ")
		log.Print(string(data1), err)
		page++
	}
}

func TestMakeShortcut(t *testing.T) {
	shortcuts := strings.ReplaceAll("https://khotruyen.club/truyen-tranh/tuyen-tap-abo-ngan-cua-nha-sec-2/", "https://", "")
	// khotruyen.club/truyen-tranh/tuyen-tap-abo-ngan-cua-nha-sec-2/
	sc := strings.Split(shortcuts, "/")

	if len(sc) > 2 {
		log.Print(sc[2])
	}
}

func TestCrawlerDetail(t *testing.T) {
	comic := &Comic{SourceURL: "https://khotruyen.club/truyen-tranh/thuan-hoa-quy-co/"}
	FetchComicDetailChapter(comic)
	data1, _ := json.MarshalIndent(comic, "", " ")
	log.Print(string(data1))

}

func TestSplit(t *testing.T) {

	x := strings.Split("https://khotruyen.club/truyen-tranh/chich-den-sang-voi-co-nguoi-yeu-song-chung/chapter-39/", "/")
	log.Print(x[5])
}
