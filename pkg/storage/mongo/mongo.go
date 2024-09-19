package mongo

import (
	"GoNews/pkg/storage"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName   = "test" // имя учебной БД
	collectionNews = "news" // имя коллекции в учебной БД
)

// Хранилище данных.
type Store struct {
	db *mongo.Client
}

func (s *Store) GetInform() string {
	return "MongoDB"
}

// Конструктор объекта хранилища.
func New(constr string) (*Store, error) {
	// подключение к СУБД MongoDB
	mongoOpts := options.Client().ApplyURI(constr)
	db, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		return nil, err
	}
	// проверка связи с БД
	err = db.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	s := Store{
		db: db,
	}

	return &s, nil
}

func (s *Store) Close() {
	s.db.Disconnect(context.Background())
}

// News возвращает последние новости из БД.

func (s *Store) News(rubric string, countNews uint) ([]storage.News, error) {

	if countNews == 0 {
		countNews = 10
	}

	collection := s.db.Database(databaseName).Collection(collectionNews)

	filter := bson.M{}
	if rubric != "" {
		filter = bson.M{"rubric": rubric}
	}

	cursor, err := collection.Find(context.Background(), filter, options.Find().SetSort(bson.M{"public_time": -1}).SetLimit(int64(countNews)))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var news []storage.News
	for cursor.Next(context.Background()) {
		var result storage.News
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		news = append(news, result)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	//fmt.Printf("%#v\n", news[0])

	return news, nil
}

// Добавляем новость в БД.
func (s *Store) AddNew(news []storage.News) error {

	collection := s.db.Database(databaseName).Collection(collectionNews)
	for _, newsRec := range news {

		// Проверка, существует ли запись с таким же link
		existingNews := storage.News{}
		err := collection.FindOne(context.Background(), bson.M{"link": newsRec.Link}).Decode(&existingNews)
		if err == nil {
			return fmt.Errorf("dublicate link %v", newsRec.Link)
		} else if err != mongo.ErrNoDocuments {
			return err
		} else {
			_, err := collection.InsertOne(context.Background(), bson.M{
				"title":       newsRec.Title,
				"content":     newsRec.Content,
				"public_time": newsRec.PublicTime,
				"image_link":  newsRec.ImageLink,
				"rubric":      newsRec.Rubric,
				"link":        newsRec.Link,
				"link_title":  newsRec.LinkTitle,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
