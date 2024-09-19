package postgres

import (
	"GoNews/pkg/storage"
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Хранилище данных.
type Store struct {
	db *pgxpool.Pool
}

func (s *Store) GetInform() string {
	return "PostgreSQL"
}

// Конструктор объекта хранилища.
func New(constr string) (*Store, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Store{
		db: db,
	}

	fmt.Println("Loaded bd: ", s.GetInform())

	return &s, nil
}

func (s *Store) Close() {
	s.db.Close()
}

// News возвращает последние новости из БД.
func (s *Store) News(rubric string, countNews uint) ([]storage.News, error) {
	if countNews == 0 {
		countNews = 10
	}

	rows, err := s.db.Query(context.Background(), `
	SELECT id, title, content, public_time, image_link, rubric, link, link_title FROM news
	WHERE rubric LIKE $1
	ORDER BY public_time DESC
	LIMIT $2
	`,
		rubric,
		countNews,
	)
	if err != nil {
		return nil, err
	}
	var news []storage.News
	for rows.Next() {
		var p storage.News
		err = rows.Scan(
			&p.Id,
			&p.Title,
			&p.Content,
			&p.PublicTime,
			&p.ImageLink,
			&p.Rubric,
			&p.Link,
			&p.LinkTitle,
		)
		if err != nil {
			return nil, err
		}
		news = append(news, p)
	}
	return news, rows.Err()
}

// Добавляем новость в БД.
func (s *Store) AddNew(news []storage.News) error {
	for _, newsRec := range news {
		_, err := s.db.Exec(context.Background(), `
		INSERT INTO news(title, content, public_time, image_link, rubric, link, link_title)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			newsRec.Title,
			newsRec.Content,
			newsRec.PublicTime,
			newsRec.ImageLink,
			newsRec.Rubric,
			newsRec.Link,
			newsRec.LinkTitle,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
