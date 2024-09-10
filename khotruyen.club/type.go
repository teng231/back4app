package khotruyenclub

type Comic struct {
	ID          int64  `json:"id,omitempty"  gorm:"primaryKey;not null;autoIncrement"`
	Title       string `json:"title,omitempty"`
	Shortcut    string `json:"shortcut,omitempty"`
	Description string `json:"description,omitempty"`
	// Author        string  `json:"author,omitempty"`
	Artist        string  `json:"artist,omitempty"`
	Status        string  `json:"status,omitempty"`
	Type          string  `json:"type,omitempty"`
	CoverImageURL string  `json:"cover_image_url,omitempty"`
	Rating        float64 `json:"rating,omitempty"`
	CreatedAt     int64   `json:"created_at,omitempty"`
	// UpdatedAt     int64   `json:"updated_at,omitempty"`
	SourceURL string `json:"source_url,omitempty"`

	LastChapter        int64 `json:"last_chapter,omitempty"`
	LastChapterUpdated int64 `json:"last_chapter_updated,omitempty" gorm:"index:idx_last_update"`

	CategoryItems string      `json:"category_items,omitempty"`
	Categories    []*Category `json:"categories,omitempty" gorm:"-"`
	TagItems      string      `json:"tag_items,omitempty"`
	Tags          []*Tag      `json:"tags,omitempty" gorm:"-"`
	Chapters      []*Chapter  `json:"chapters,omitempty" gorm:"-"`
	TotalChapter  int32       `json:"total_chapter,omitempty"`
}

type ComicRequest struct {
	ID  int64   `json:"id,omitempty"`
	IDs []int64 `json:"ids,omitempty"`

	CategoryID int64    `json:"category_id,omitempty"`
	TagID      string   `json:"tag_id,omitempty"`
	SourceURL  string   `json:"source_url,omitempty"`
	Status     string   `json:"status,omitempty"` // 2 active | 3 inactive or del
	Limit      int      `json:"limit,omitempty"`
	Page       int      `json:"page,omitempty"`
	Cols       []string `json:"cols,omitempty"`
	Order      string   `json:"order,omitempty"`
}

type Category struct {
	ID        int64  `json:"id,omitempty"  gorm:"primaryKey;not null;autoIncrement"`
	Title     string `json:"title,omitempty"`
	Shortcut  string `json:"shortcut,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	// Comics    []Comic `json:"comics,omitempty" gorm:"many2many:comic_categories;"`
}

type CategoryRequest struct {
	ID     int64    `json:"id,omitempty"`
	Title  string   `json:"title,omitempty"`
	Titles []string `json:"titles,omitempty"`

	Status int      `json:"status,omitempty"` // 2 active | 3 inactive or del
	Limit  int      `json:"limit,omitempty"`
	Page   int      `json:"page,omitempty"`
	Cols   []string `json:"cols,omitempty"`
	Order  string   `json:"order,omitempty"`
}

type Tag struct {
	ID        int64  `json:"id,omitempty"  gorm:"primaryKey;not null;autoIncrement"`
	Title     string `json:"title,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	// Comics    []Comic `json:"comics,omitempty" gorm:"many2many:comic_tags;"`
}
type TagRequest struct {
	ID     int64    `json:"id,omitempty"`
	Title  string   `json:"title,omitempty"`
	Titles []string `json:"titles,omitempty"`
	Status int      `json:"status,omitempty"` // 2 active | 3 inactive or del
	Limit  int      `json:"limit,omitempty"`
	Page   int      `json:"page,omitempty"`
	Cols   []string `json:"cols,omitempty"`
	Order  string   `json:"order,omitempty"`
}
type Chapter struct {
	ID              int64    `json:"id,omitempty"  gorm:"primaryKey;not null;autoIncrement"`
	ComicID         int64    `json:"comic_id,omitempty" gorm:"index:idx_member"`
	Shortcut        string   `json:"shortcut,omitempty"`
	Title           string   `json:"title,omitempty"`
	ChapterNumber   int64    `json:"chapter_number,omitempty"`
	Content         string   `json:"content,omitempty"`
	ImageURLs       string   `json:"image_urls,omitempty"`
	ImageURLsArr    []string `json:"image_urls_arr,omitempty" gorm:"-"`
	CreatedAt       int64    `json:"created_at,omitempty"`
	UpdatedAt       int64    `json:"updated_at,omitempty"`
	SourceURL       string   `json:"source_url,omitempty"`
	PreviousChapter *Chapter `json:"previous_chapter,omitempty" gorm:"-"`
	NextChapter     *Chapter `json:"next_chapter,omitempty" gorm:"-"`
}

type ChapterRequest struct {
	ID        int64    `json:"id,omitempty"`
	ComicID   int64    `json:"comic_id,omitempty"`
	SourceURL string   `json:"source_url,omitempty"`
	Shortcut  string   `json:"shortcut,omitempty"`
	Status    int      `json:"status,omitempty"` // 2 active | 3 inactive or del
	Limit     int      `json:"limit,omitempty"`
	Page      int      `json:"page,omitempty"`
	Cols      []string `json:"cols,omitempty"`
	Order     string   `json:"order,omitempty"`
}

type ComicCategory struct {
	ComicID    int64   `json:"comic_id,omitempty" gorm:"primaryKey;not null"`
	CategoryID int64   `json:"category_id,omitempty" gorm:"primaryKey;not null"`
	ComicIDs   []int64 `gorm:"-"`
}

type ComicTag struct {
	ComicID int64 `json:"comic_id,omitempty" gorm:"primaryKey;not null"`
	TagID   int64 `json:"tag_id,omitempty" gorm:"primaryKey;not null"`

	ComicIDs []int64 `gorm:"-"`
}
