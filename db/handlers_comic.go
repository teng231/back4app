package db

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	khotruyenclub "github.com/teng231/back4app/khotruyen.club"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	tblComic         = "comics"
	tblCategory      = "categories"
	tblTag           = "tags"
	tblChapter       = "chapters"
	tblComicTag      = "comic_tags"
	tblComicCategory = "comic_categories"
)

type ComicDB struct {
	engine *gorm.DB
}

type IComicDB interface {
	StatusCheck() error
	Migrate() error

	//db for comic
	InsertComic(comic *khotruyenclub.Comic) error
	UpdateComic(updator, selector *khotruyenclub.Comic) error
	FindComic(in *khotruyenclub.Comic) (*khotruyenclub.Comic, error)
	ListComics(rq *khotruyenclub.ComicRequest) ([]*khotruyenclub.Comic, error)
	ListSimilarComics(rq *khotruyenclub.Comic, limit int) ([]*khotruyenclub.Comic, error)
	CountComics(rq *khotruyenclub.ComicRequest) (int64, error)
	ScanComicTable(cond *khotruyenclub.ComicRequest, vChan chan<- *khotruyenclub.Comic, wg *sync.WaitGroup) error
	ListComicByCategoryId(rq *khotruyenclub.ComicRequest) ([]*khotruyenclub.Comic, error)

	InsertCategory(category *khotruyenclub.Category) error
	UpdateCategory(updator, selector *khotruyenclub.Category) error
	FindCategory(in *khotruyenclub.Category) (*khotruyenclub.Category, error)
	ListCategories(rq *khotruyenclub.CategoryRequest) ([]*khotruyenclub.Category, error)
	UpsertCategory(categories []*khotruyenclub.Category) ([]*khotruyenclub.Category, error)

	FindTag(in *khotruyenclub.Tag) (*khotruyenclub.Tag, error)
	UpdateTag(updator, selector *khotruyenclub.Tag) error
	InsertTag(tag *khotruyenclub.Tag) error
	ListTags(rq *khotruyenclub.TagRequest) ([]*khotruyenclub.Tag, error)
	UpsertTag(tags []*khotruyenclub.Tag) ([]*khotruyenclub.Tag, error)

	FindChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error)
	FindPreviousChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error)
	FindNextChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error)
	InsertChapter(chapters ...*khotruyenclub.Chapter) error
	UpdateChapter(updator, selector *khotruyenclub.Chapter) error
	ListChapters(rq *khotruyenclub.ChapterRequest) ([]*khotruyenclub.Chapter, error)
	CountChapters(rq *khotruyenclub.ChapterRequest) (int64, error)

	InsertComicTag(comicTag ...*khotruyenclub.ComicTag) error
	// DeleteComicTag(selector *khotruyenclub.ComicTag) error

	// DeleteComicCategory(selector *khotruyenclub.ComicCategory) error
	InsertComicCategory(comicCategory ...*khotruyenclub.ComicCategory) error
}
type PreloadQuery struct {
	Table     string
	Condition any
}

func preloadRequest(preloads []*PreloadQuery, ss *gorm.DB) {
	for _, preload := range preloads {
		if preload.Condition == nil {
			ss.Preload(preload.Table)
			continue
		}
		ss.Preload(preload.Table, func(db *gorm.DB) *gorm.DB {
			return db.Table(preload.Table).Where(preload.Condition)
		})
	}
}
func (d *ComicDB) StatusCheck() error {
	conn, err := d.engine.DB()

	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.PingContext(ctx)
}

func (d *ComicDB) Migrate() error {
	if err := d.engine.Table(tblComic).AutoMigrate(khotruyenclub.Comic{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblCategory).AutoMigrate(khotruyenclub.Category{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblTag).AutoMigrate(khotruyenclub.Tag{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblChapter).AutoMigrate(khotruyenclub.Chapter{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblComicTag).AutoMigrate(khotruyenclub.ComicTag{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblComicCategory).AutoMigrate(khotruyenclub.ComicCategory{}); err != nil {
		return err
	}
	return nil
}

func NewComicDb(dsn string) (*ComicDB, error) {
	db, err := gorm.Open(mysql.New(
		mysql.Config{DSN: dsn}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		/*i
		GORM perform write (create/update/delete) operations run inside a transaction to ensure data consistency, you can disable it during initialization if it is not required, you will gain about 30%+ performance improvement after that
		*/
		SkipDefaultTransaction: true,
		// PrepareStmt:            true,
	})
	if err != nil {
		log.Print("connect db fail ", err)
		return nil, err
	}

	return &ComicDB{engine: db}, nil
}

// InsertComic ...
func (d *ComicDB) InsertComic(comic *khotruyenclub.Comic) error {
	result := d.engine.Table(tblComic).Create(comic)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert comic")
	}
	return nil
}

// UpdateComic ...
func (d *ComicDB) UpdateComic(updator, selector *khotruyenclub.Comic) error {
	result := d.engine.Table(tblComic).Where(selector).Updates(map[string]any{
		"last_chapter":         updator.LastChapter,
		"last_chapter_updated": updator.LastChapterUpdated,
		"total_chapter":        gorm.Expr("total_chapter + ?", updator.TotalChapter),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot update comic")
	}
	return nil
}

// FindComic ...
func (d *ComicDB) FindComic(in *khotruyenclub.Comic) (*khotruyenclub.Comic, error) {
	out := &khotruyenclub.Comic{}
	err := d.engine.Table(tblComic).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("comic not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (d *ComicDB) CountComics(rq *khotruyenclub.ComicRequest) (int64, error) {
	ss := d.engine.Table(tblComic)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	var count int64
	err := ss.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (d *ComicDB) ListSimilarComics(rq *khotruyenclub.Comic, limit int) ([]*khotruyenclub.Comic, error) {
	var comics []*khotruyenclub.Comic
	ss := d.engine.Table(tblComic).Select("DISTINCT comics.*")
	ss.Joins("JOIN " + tblComicCategory + " ON " + tblComic + ".id = " + tblComicCategory + ".comic_id")
	ss.Where(tblComicCategory+".category_id IN (?)",
		d.engine.Select("category_id").
			Table(tblComicCategory).
			Where("comic_id = ?", rq.ID)).
		Where("comic_id != ?", rq.ID).Limit(limit).Find(&comics)
	return comics, nil
}

func (d *ComicDB) ListComicByCategoryId(rq *khotruyenclub.ComicRequest) ([]*khotruyenclub.Comic, error) {
	var comics []*khotruyenclub.Comic
	ss := d.engine.Table(tblComic).Select("DISTINCT comics.*")
	ss.Joins("JOIN " + tblComicCategory + " ON " + tblComic + ".id = " + tblComicCategory + ".comic_id")
	ss.Where(tblComicCategory+".category_id = ", rq.CategoryID).
		Limit(rq.Limit).
		Order(tblComic + ".last_chapter_updated desc ")
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	ss.Find(&comics)
	return comics, nil
}
func (d *ComicDB) ListComics(rq *khotruyenclub.ComicRequest) ([]*khotruyenclub.Comic, error) {
	var comics []*khotruyenclub.Comic
	ss := d.engine.Table(tblComic)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	if len(rq.IDs) > 0 {
		ss.Where("ids IN ?", rq.IDs)
	}
	if rq.Status != "" {
		ss.Where("status = ?", rq.Status)
	}
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	if rq.Order != "" {
		ss.Order(rq.Order)
	}
	if len(rq.Cols) > 0 {
		ss.Select(rq.Cols)
	}
	if rq.Order == "" {
		ss.Order("id desc")
	} else {
		ss.Order(rq.Order)
	}
	err := ss.Find(&comics).Error
	if err != nil {
		return nil, err
	}
	return comics, nil
}

func (d *ComicDB) ScanComicTable(cond *khotruyenclub.ComicRequest, vChan chan<- *khotruyenclub.Comic, wg *sync.WaitGroup) error {
	ss := d.engine.Table(tblComic)
	if len(cond.Cols) > 0 {
		ss.Select(cond.Cols)
	}
	rows, err := ss.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	wg.Add(1)
	defer wg.Done()
	bean := new(khotruyenclub.Comic)
	for rows.Next() {
		err := d.engine.ScanRows(rows, bean)
		if err != nil {
			continue
		}
		vChan <- bean
		bean = new(khotruyenclub.Comic)
		wg.Add(1)
	}
	return nil
}

// DeleteComic ...
func (d *ComicDB) DeleteComic(selector *khotruyenclub.Comic) error {
	result := d.engine.Table(tblComic).Where(selector).Delete(&khotruyenclub.Comic{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete comic")
	}
	return nil
}

// InsertCategory ...
func (d *ComicDB) InsertCategory(category *khotruyenclub.Category) error {
	result := d.engine.Table(tblCategory).Create(category)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert category")
	}
	return nil
}

func (d *ComicDB) UpsertCategory(categories []*khotruyenclub.Category) ([]*khotruyenclub.Category, error) {
	if len(categories) == 0 {
		return []*khotruyenclub.Category{}, nil
	}
	// Lấy danh sách các tiêu đề hiện tại
	var titles []string
	for _, category := range categories {
		titles = append(titles, category.Title)
	}

	// Tìm các category đã tồn tại trong DB
	var existingCategories []*khotruyenclub.Category
	err := d.engine.Table(tblCategory).Where("title IN ?", titles).Find(&existingCategories).Error
	if err != nil {
		return nil, err
	}

	// Tạo một map từ tiêu đề để dễ kiểm tra
	existingMap := make(map[string]*khotruyenclub.Category)
	for _, category := range existingCategories {
		existingMap[category.Title] = category
	}

	// Danh sách các category cần thêm mới
	var newCategories []*khotruyenclub.Category
	for _, category := range categories {
		if _, exists := existingMap[category.Title]; !exists {
			newCategories = append(newCategories, category)
		} else {
			// Bổ sung dữ liệu vào slice categories nếu đã tồn tại
			category = existingMap[category.Title]
		}
	}

	// Insert các category chưa tồn tại
	if len(newCategories) > 0 {
		if err := d.engine.Table(tblCategory).Create(&newCategories).Error; err != nil {
			return nil, err
		}
		// Bổ sung ID và dữ liệu vào categories sau khi insert
		for i, category := range categories {
			for _, newCategory := range newCategories {
				if category.Title == newCategory.Title {
					categories[i] = newCategory
				}
			}
		}
	}
	respCate := make([]*khotruyenclub.Category, 0)
	respCate = append(respCate, existingCategories...)
	respCate = append(respCate, newCategories...)
	return respCate, nil
}

// UpdateCategory ...
func (d *ComicDB) UpdateCategory(updator, selector *khotruyenclub.Category) error {
	result := d.engine.Table(tblCategory).Where(selector).Updates(updator)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot update category")
	}
	return nil
}

// FindCategory ...
func (d *ComicDB) FindCategory(in *khotruyenclub.Category) (*khotruyenclub.Category, error) {
	out := &khotruyenclub.Category{}
	err := d.engine.Table(tblCategory).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("category not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (d *ComicDB) ListCategories(rq *khotruyenclub.CategoryRequest) ([]*khotruyenclub.Category, error) {
	var categories []*khotruyenclub.Category
	ss := d.engine.Table(tblCategory)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	if rq.Title != "" {
		ss.Where("title = ?", rq.Title)
	}
	if len(rq.Titles) > 0 {
		ss.Where("title IN ?", rq.Titles)
	}
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	if rq.Order != "" {
		ss.Order(rq.Order)
	}
	if len(rq.Cols) > 0 {
		ss.Select(rq.Cols)
	}
	err := ss.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// DeleteCategory ...
func (d *ComicDB) DeleteCategory(selector *khotruyenclub.Category) error {
	result := d.engine.Table(tblCategory).Where(selector).Delete(&khotruyenclub.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete category")
	}
	return nil
}

// InsertTag ...
func (d *ComicDB) InsertTag(tag *khotruyenclub.Tag) error {
	result := d.engine.Table(tblTag).Create(tag)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert tag")
	}
	return nil
}

// UpdateTag ...
func (d *ComicDB) UpdateTag(updator, selector *khotruyenclub.Tag) error {
	result := d.engine.Table(tblTag).Where(selector).Updates(updator)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot update tag")
	}
	return nil
}

// FindTag ...
func (d *ComicDB) FindTag(in *khotruyenclub.Tag) (*khotruyenclub.Tag, error) {
	out := &khotruyenclub.Tag{}
	err := d.engine.Table(tblTag).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("tag not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (d *ComicDB) ListTags(rq *khotruyenclub.TagRequest) ([]*khotruyenclub.Tag, error) {
	var tags []*khotruyenclub.Tag
	ss := d.engine.Table(tblTag)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	if rq.Title != "" {
		ss.Where("title = ?", rq.Title)
	}
	if len(rq.Titles) > 0 {
		ss.Where("title IN ?", rq.Titles)
	}
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	if rq.Order != "" {
		ss.Order(rq.Order)
	}
	if len(rq.Cols) > 0 {
		ss.Select(rq.Cols)
	}
	err := ss.Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (d *ComicDB) UpsertTag(tags []*khotruyenclub.Tag) ([]*khotruyenclub.Tag, error) {
	if len(tags) == 0 {
		return []*khotruyenclub.Tag{}, nil
	}
	// Lấy danh sách các tiêu đề hiện tại
	var titles []string
	for _, tag := range tags {
		titles = append(titles, tag.Title)
	}

	// Tìm các tag đã tồn tại trong DB
	var existingTags []*khotruyenclub.Tag
	err := d.engine.Table(tblTag).Where("title IN ?", titles).Find(&existingTags).Error
	if err != nil {
		return nil, err
	}

	// Tạo một map từ tiêu đề để dễ kiểm tra
	existingMap := make(map[string]*khotruyenclub.Tag)
	for _, tag := range existingTags {
		existingMap[tag.Title] = tag
	}

	// Danh sách các tag cần thêm mới
	var newTags []*khotruyenclub.Tag
	for _, tag := range tags {
		if _, exists := existingMap[tag.Title]; !exists {
			newTags = append(newTags, tag)
		} else {
			// Bổ sung dữ liệu vào slice tags nếu đã tồn tại
			tag = existingMap[tag.Title]
		}
	}

	// Insert các tag chưa tồn tại
	if len(newTags) > 0 {
		if err := d.engine.Table(tblTag).Create(&newTags).Error; err != nil {
			return nil, err
		}
		// Bổ sung ID và dữ liệu vào tags sau khi insert
		for i, tag := range tags {
			for _, newTag := range newTags {
				if tag.Title == newTag.Title {
					tags[i] = newTag
				}
			}
		}
	}
	respTag := make([]*khotruyenclub.Tag, 0)
	respTag = append(respTag, existingTags...)
	respTag = append(respTag, newTags...)
	return respTag, nil
}

// DeleteTag ...
func (d *ComicDB) DeleteTag(selector *khotruyenclub.Tag) error {
	result := d.engine.Table(tblTag).Where(selector).Delete(&khotruyenclub.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete tag")
	}
	return nil
}

// InsertChapter ...
func (d *ComicDB) InsertChapter(chapters ...*khotruyenclub.Chapter) error {
	result := d.engine.Table(tblChapter).Create(chapters)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert chapter")
	}
	return nil
}

// UpdateChapter ...
func (d *ComicDB) UpdateChapter(updator, selector *khotruyenclub.Chapter) error {
	result := d.engine.Table(tblChapter).Where(selector).Updates(updator)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot update chapter")
	}
	return nil
}

// FindChapter ...
func (d *ComicDB) FindChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error) {
	out := &khotruyenclub.Chapter{}
	err := d.engine.Table(tblChapter).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("chapter not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FindChapter ...
func (d *ComicDB) FindNextChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error) {
	out := &khotruyenclub.Chapter{}
	err := d.engine.Table(tblChapter).
		Where("comic_id = ?", in.ComicID).
		Where("id > ?", in.ID).
		Order("id asc").
		Select("id", "shortcut", "title").
		Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("chapter not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FindChapter ...
func (d *ComicDB) FindPreviousChapter(in *khotruyenclub.Chapter) (*khotruyenclub.Chapter, error) {
	out := &khotruyenclub.Chapter{}
	err := d.engine.Table(tblChapter).
		Where("comic_id = ?", in.ComicID).
		Where("id < ?", in.ID).
		Order("id desc").
		Select("id", "shortcut", "title").
		Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("chapter not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteChapter ...
func (d *ComicDB) DeleteChapter(selector *khotruyenclub.Chapter) error {
	result := d.engine.Table(tblChapter).Where(selector).Delete(&khotruyenclub.Chapter{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete chapter")
	}
	return nil
}

func (d *ComicDB) ListChapters(rq *khotruyenclub.ChapterRequest) ([]*khotruyenclub.Chapter, error) {
	var chapters []*khotruyenclub.Chapter
	ss := d.engine.Table(tblChapter)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	if rq.ComicID != 0 {
		ss.Where("comic_id = ?", rq.ComicID)
	}
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	if rq.Order != "" {
		ss.Order(rq.Order)
	}
	if len(rq.Cols) > 0 {
		ss.Select(rq.Cols)
	}
	err := ss.Order("id desc").Find(&chapters).Error
	if err != nil {
		return nil, err
	}
	return chapters, nil
}
func (d *ComicDB) CountChapters(rq *khotruyenclub.ChapterRequest) (int64, error) {
	ss := d.engine.Table(tblChapter)
	if rq.ID != 0 {
		ss.Where("id = ?", rq.ID)
	}
	if rq.ComicID != 0 {
		ss.Where("comic_id = ?", rq.ComicID)
	}
	var count int64
	err := ss.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// InsertComicCategory ...
func (d *ComicDB) InsertComicCategory(comicCategory ...*khotruyenclub.ComicCategory) error {
	result := d.engine.Table(tblComicCategory).Create(comicCategory)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert comic category")
	}
	return nil
}

// DeleteComicCategory ...
func (d *ComicDB) DeleteComicCategory(selector *khotruyenclub.ComicCategory) error {
	result := d.engine.Table(tblComicCategory).Where(selector).Delete(&khotruyenclub.ComicCategory{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete comic category")
	}
	return nil
}

// FindComicCategory ...
func (d *ComicDB) FindComicCategory(in *khotruyenclub.ComicCategory) (*khotruyenclub.ComicCategory, error) {
	out := &khotruyenclub.ComicCategory{}
	err := d.engine.Table(tblComicCategory).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("comic category not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FindComicCategory ...
func (d *ComicDB) ListComicCategory(in *khotruyenclub.ComicCategory) ([]*khotruyenclub.ComicCategory, error) {
	list := make([]*khotruyenclub.ComicCategory, 0)
	err := d.engine.Table(tblComicCategory).Where(in).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// InsertComicTag ...
func (d *ComicDB) InsertComicTag(comicTag ...*khotruyenclub.ComicTag) error {
	result := d.engine.Table(tblComicTag).Create(comicTag)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot insert comic tag")
	}
	return nil
}

// FindComicTag ...
func (d *ComicDB) FindComicTag(in *khotruyenclub.ComicTag) (*khotruyenclub.ComicTag, error) {
	out := &khotruyenclub.ComicTag{}
	err := d.engine.Table(tblComicTag).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("comic tag not found")
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteComicTag ...
func (d *ComicDB) DeleteComicTag(selector *khotruyenclub.ComicTag) error {
	result := d.engine.Table(tblComicTag).Where(selector).Delete(&khotruyenclub.ComicTag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cannot delete comic tag")
	}
	return nil
}

// FindComicCategory ...
func (d *ComicDB) ListComicTag(in *khotruyenclub.ComicTag) ([]*khotruyenclub.ComicTag, error) {
	list := make([]*khotruyenclub.ComicTag, 0)
	err := d.engine.Table(tblComicTag).Where(in).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
