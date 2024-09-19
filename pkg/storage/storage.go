package storage

// Публикация, получаемая из RSS.
type News struct {
	Id         int
	Title      string `bson:"title"`
	Content    string `bson:"content"`
	PublicTime int64  `bson:"public_time"`
	ImageLink  string `bson:"image_link"`
	Rubric     string `bson:"rubric"`
	Link       string `bson:"link"`
	LinkTitle  string `bson:"link_title"`
}

// Interface задаёт контракт на работу с БД.
type Interface interface {
	GetInform() string
	Close()

	News(rubric string, countNews uint) ([]News, error) // News возвращает последние новости из БД.
	AddNew(news []News) error                           // Добавляем новость в БД.
}
