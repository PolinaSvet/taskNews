package main

import (
	"GoNews/pkg/api"
	"GoNews/pkg/rss"
	"GoNews/pkg/storage"
	"GoNews/pkg/storage/mongo"
	"GoNews/pkg/storage/postgres"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"flag"
	"fmt"
	"log"
	"net/http"
)

// tasknews> go test ./... -coverprofile=coverage.out
// tasknews> go test ./... -v -coverprofile=coverage.out

type ConfigRSS struct {
	RSS map[string]struct {
		Link  []string `json:"link"`
		Image string   `json:"image"`
	} `json:"rss"`
	Duration int `json:"duration"`
}

// Сервер GoNews
type server struct {
	db  storage.Interface
	api *api.API
}

func main() {

	// Обрабатываем флаги при запуске программы
	// go run server.go -typebd pg
	var typebd string

	flag.StringVar(&typebd, "typebd", "pg", "DataBase: pg-PostgreSQL, mongo-MongoDB")
	flag.Parse()

	fmt.Println("flags: type bd->", typebd)

	// Создаём объект сервера.
	var srv server
	// Создаём объекты баз данных.
	// Инициализируем хранилище сервера конкретной БД.
	switch typebd {
	case "pg":
		// Реляционная БД PostgreSQL.
		connstr := os.Getenv("NEWSDBPG")
		if connstr == "" {
			log.Fatal(errors.New("no connection to pg bd"))
		}
		db_pg, err := postgres.New(connstr)
		if err != nil {
			log.Fatal(err)
		}
		srv.db = db_pg

	case "mongo":
		// Не реляционная БД MongoDB.
		connstr := os.Getenv("NEWSDBMONGO")
		if connstr == "" {
			log.Fatal(errors.New("no connection to mongo bd"))
		}
		db_mongo, err := mongo.New(connstr)
		if err != nil {
			log.Fatal(err)
		}
		srv.db = db_mongo

	}

	defer srv.db.Close()

	srv.api = api.New(srv.db)

	// чтение и раскодирование файла конфигурации
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	var configRSS ConfigRSS
	err = json.Unmarshal(data, &configRSS)
	if err != nil {
		log.Fatal(err)
	}

	newsChannel := make(chan []storage.News)
	errorChannel := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers

	// парсим rss, каждую ссылку в отдельном потоке
	go getNewsFromAllRSS(ctx, configRSS, newsChannel, errorChannel)
	// записываем информацию по каждой ссылке в бд
	go writeNewsToDB(ctx, srv.db, newsChannel, errorChannel)
	// выводим ошибки
	go handleErrors(ctx, errorChannel)

	fmt.Println("Запуск веб-сервера на http://127.0.0.1:8080 ...")
	http.ListenAndServe(":8080", srv.api.Router())
}

func getNewsFromAllRSS(ctx context.Context, configRSS ConfigRSS, news chan<- []storage.News, errs chan<- error) {
	for rubric, value := range configRSS.RSS {
		for _, link := range value.Link {
			go func(url, rubric, image string) {
				for {
					select {
					case <-ctx.Done(): // context checking
						return // returning not to leak the goroutine
					default:
						newsResp, err := rss.GetNewsFromRss(url, rubric, image)
						if err != nil {
							errs <- err
						} else {
							news <- newsResp
						}

						time.Sleep(time.Minute * time.Duration(configRSS.Duration))
					}
				}
			}(link, rubric, value.Image)
		}
	}
}

func writeNewsToDB(ctx context.Context, db storage.Interface, news <-chan []storage.News, errs chan<- error) {
	for newsBatch := range news {
		select {
		case <-ctx.Done():
			return
		default:
			err := db.AddNew(newsBatch)
			if err != nil {
				errs <- err
				continue
			}
		}
	}
}

func handleErrors(ctx context.Context, errs <-chan error) {
	for err := range errs {
		select {
		case <-ctx.Done():
			return
		default:
			log.Println("error:", err)
			//log.Println("handleErrors:", runtime.NumGoroutine())
		}
	}
}
