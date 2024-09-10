package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	khotruyenclub "github.com/teng231/back4app/khotruyen.club"
)

func (r *Serve) renderComicHome(c *gin.Context) {
	comics, _ := r.dbc.ListComics(&khotruyenclub.ComicRequest{
		Limit: 6,
		Page:  1,
		Order: "last_chapter_updated desc",
	})
	comicsDone, _ := r.dbc.ListComics(&khotruyenclub.ComicRequest{
		Limit:  6,
		Page:   1,
		Order:  "last_chapter_updated desc",
		Status: "Completed",
	})

	r.BindTags(comics...)
	r.BindCategories(comics...)

	r.BindTags(comicsDone...)
	r.BindCategories(comicsDone...)
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"comics":        comics,
		"comicsDone":    comicsDone,
		"mapCategories": mcategories,
	})
}
func (r *Serve) renderComicPage(c *gin.Context) {
	comic, _ := r.dbc.FindComic(&khotruyenclub.Comic{
		Shortcut: c.Param("comic_shortcut"),
	})

	chapters, _ := r.dbc.ListChapters(&khotruyenclub.ChapterRequest{
		ComicID: comic.ID,
		Cols:    []string{"id", "title", "shortcut"},
	})
	r.BindTags(comic)
	r.BindCategories(comic)
	similarComics, _ := r.dbc.ListSimilarComics(comic, 4)
	c.HTML(http.StatusOK, "comic-details.tmpl", gin.H{
		"comic":         comic,
		"chapters":      chapters,
		"similarComics": similarComics,
	})
}

func (r *Serve) renderComicChapterPage(c *gin.Context) {
	comic, _ := r.dbc.FindComic(&khotruyenclub.Comic{
		Shortcut: c.Param("comic_shortcut"),
	})
	chapter, _ := r.dbc.FindChapter(&khotruyenclub.Chapter{
		ComicID:  comic.ID,
		Shortcut: c.Param("chapter_shortcut"),
	})
	nextChapter, _ := r.dbc.FindNextChapter(chapter)
	previousChapter, _ := r.dbc.FindPreviousChapter(chapter)
	chapter.ImageURLsArr = strings.Split(chapter.ImageURLs, ",")

	c.HTML(http.StatusOK, "comic-watching.tmpl", gin.H{
		"comic":           comic,
		"chapter":         chapter,
		"nextChapter":     nextChapter,
		"previousChapter": previousChapter,
	})
}
