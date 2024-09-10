package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teng231/back4app/db"
	"github.com/teng231/back4app/httpserver"
	khotruyenclub "github.com/teng231/back4app/khotruyen.club"
)

type Serve struct {
	*httpserver.Engine
	dbc *db.ComicDB
}

func newServer() *Serve {
	// comicDb, err := db.NewComicDb(cfg.DbComicDSN)
	// if err != nil {
	// 	log.Print("db connect fail ", err)
	// }
	r := &Serve{Engine: &httpserver.Engine{Engine: gin.New()}}
	// r.dbc = comicDb
	r.SetTrustedProxies(nil)

	return r
}

func (r *Serve) customMiddleware() *Serve {
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(httpserver.UsingCORSMode(cfg.DomainAllowed), httpserver.UsingTimeTracing())
	return r
}

func (r *Serve) registerHandlers() *Serve {

	// r.GET("/", r.handleHomelander)
	r.GET("/ping", r.handlePing)

	r.GETs([]string{"/a", "/b"}, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "[4client]👉Hope for the best, prepare for the worst")
	})

	r.GET("/response-err", httpserver.UsingTimeTracing(), func(ctx *gin.Context) {
		httpserver.JSON(ctx, http.StatusBadRequest, "[4client]👉Hope for the best, prepare for the worst")
	})

	// comicRoute := r.Group("/comic")
	// comicRoute.GET("/comics", httpserver.UsingTimeTracing(), r.handleListComics)
	// comicRoute.GET("/chapters", httpserver.UsingTimeTracing(), r.handlerListChapters)
	// comicRoute.GET("/categories", httpserver.UsingTimeTracing(), r.handlerListCategories)
	// comicRoute.GET("/tags", httpserver.UsingTimeTracing(), r.handleListTags)
	return r
}

var (
	mcategories = map[int64]*khotruyenclub.Category{}
	mtags       = map[int64]*khotruyenclub.Tag{}
)

func (r *Serve) BindCategories(comics ...*khotruyenclub.Comic) {
	comicIds := []int64{}
	for _, comic := range comics {
		comicIds = append(comicIds, comic.ID)
	}
	comicCates, _ := r.dbc.ListComicCategory(&khotruyenclub.ComicCategory{ComicIDs: comicIds})

	for _, comic := range comics {
		comic.Categories = make([]*khotruyenclub.Category, 0)

		for _, comicCate := range comicCates {
			if comicCate.ComicID == comic.ID {
				comic.Categories = append(comic.Categories, mcategories[comicCate.CategoryID])
			}
		}
	}
}
func (r *Serve) BindTags(comics ...*khotruyenclub.Comic) {
	comicIds := []int64{}
	for _, comic := range comics {
		comicIds = append(comicIds, comic.ID)
	}

	comicTags, _ := r.dbc.ListComicTag(&khotruyenclub.ComicTag{ComicIDs: comicIds})

	for _, comic := range comics {
		comic.Tags = make([]*khotruyenclub.Tag, 0)

		for _, comicTag := range comicTags {
			if comicTag.ComicID == comic.ID {
				comic.Tags = append(comic.Tags, mtags[comicTag.TagID])
			}
		}
	}
}

func (r *Serve) registerWeb() *Serve {
	r.LoadHTMLFiles("web/index.tmpl",
		"web/comic-details.tmpl",
		"web/comic-watching.tmpl",
	)
	r.Static("/css", "./web/css")
	r.Static("/js", "./web/js")
	r.Static("/fonts", "./web/fonts")
	r.Static("/img", "./web/img")

	categories, _ := r.dbc.ListCategories(&khotruyenclub.CategoryRequest{})

	for _, cate := range categories {
		mcategories[cate.ID] = cate
	}

	tags, _ := r.dbc.ListTags(&khotruyenclub.TagRequest{})

	for _, tag := range tags {
		mtags[tag.ID] = tag
	}

	r.GET("/", r.renderComicHome)
	r.GET("/comic/:comic_shortcut", r.renderComicPage)
	r.GET("/comic-chapter/:comic_shortcut/:chapter_shortcut", r.renderComicChapterPage)

	return r
}

func (r *Serve) listComicSimilar(comic *khotruyenclub.Comic, total int) []*khotruyenclub.Comic {
	// ComicIDs := []int{}
	// comicCates, _ := r.dbc.FindComicCategory(&khotruyenclub.ComicCategory{ComicID: comic.ID})

	comics, _ := r.dbc.ListComics(&khotruyenclub.ComicRequest{
		// IDs: ComicIDs,
		// CategoryID: pickCateId,
		// TagID:      pickTagId,
		// Order: "last_chapter_updated desc",
		Limit: total,
	})
	return comics
}
