package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	khotruyenclub "github.com/teng231/back4app/khotruyen.club"
)

func (r *Serve) handleHomelander(ctx *gin.Context) {
	ctx.String(http.StatusOK, "[4client]👉Hope for the best, prepare for the worst")
}

func (r *Serve) handlePing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *Serve) handleListComics(ctx *gin.Context) {
	var current, pageSize int

	currentPageQ := ctx.Query("current")
	curr, err := strconv.Atoi(currentPageQ)
	if err == nil {
		current = curr
	}
	pageSizeQ := ctx.Query("pageSize")
	curr2, err := strconv.Atoi(pageSizeQ)
	if err == nil {
		pageSize = curr2
	}
	idStr := ctx.Query("id")
	id, _ := strconv.Atoi(idStr)
	comics, err := r.dbc.ListComics(&khotruyenclub.ComicRequest{
		Limit:  pageSize,
		Page:   current,
		Order:  ctx.Query("order"),
		Status: ctx.Query("status"),
		ID:     int64(id),
	})
	count, err := r.dbc.CountComics(&khotruyenclub.ComicRequest{ID: int64(id)})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"comics": comics,
		"total":  count,
	})
}
func (r *Serve) handlerListChapters(ctx *gin.Context) {
	comicIDstr := ctx.Query("comic_id")
	comicID, _ := strconv.Atoi(comicIDstr)
	if comicID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "not found comic_id"})
		return
	}
	var current, pageSize int

	currentPageQ := ctx.Query("current")
	curr, err := strconv.Atoi(currentPageQ)
	if err == nil {
		current = curr
	}
	pageSizeQ := ctx.Query("pageSize")
	curr2, err := strconv.Atoi(pageSizeQ)
	if err == nil {
		pageSize = curr2
	}
	chapters, err := r.dbc.ListChapters(&khotruyenclub.ChapterRequest{
		ComicID: int64(comicID),
		Limit:   pageSize,
		Page:    current,
	})
	count, err := r.dbc.CountChapters(&khotruyenclub.ChapterRequest{ComicID: int64(comicID)})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"chapters": chapters,
		"total":    count,
	})
}

func (r *Serve) handlerListCategories(ctx *gin.Context) {

	categories, err := r.dbc.ListCategories(&khotruyenclub.CategoryRequest{
		// ComicID: int64(comicID),
		// Limit:   pageSize,
		// Page:    current,
	})
	// count, err := r.dbc.CountChapters(&khotruyenclub.ChapterRequest{ComicID: int64(comicID)})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"categories": categories,
		// "total":    count,
	})
}

func (r *Serve) handleListTags(ctx *gin.Context) {

	tags, err := r.dbc.ListTags(&khotruyenclub.TagRequest{
		// ComicID: int64(comicID),
		// Limit:   pageSize,
		// Page:    current,
	})
	// count, err := r.dbc.CountChapters(&khotruyenclub.ChapterRequest{ComicID: int64(comicID)})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"tags": tags,
		// "total":    count,
	})
}
