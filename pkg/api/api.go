package api

import (
	"GoNews/pkg/storage"
	"encoding/json"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
)

// Программный интерфейс сервера GoNews
type API struct {
	db     storage.Interface
	router *mux.Router
}

// Конструктор объекта API
func New(db storage.Interface) *API {
	api := API{
		db: db,
	}
	api.router = mux.NewRouter()
	api.router.Use(errorMiddleware)
	api.router.Use(corsMiddleware)
	api.endpoints()
	return &api
}

// Получение маршрутизатора запросов.
// Требуется для передачи маршрутизатора веб-серверу.
func (api *API) Router() *mux.Router {
	return api.router
}

func errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if _, ok := err.(error); ok { // Проверка на ошибку
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Регистрация обработчиков API.
func (api *API) endpoints() {

	api.router.HandleFunc("/", api.templateHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/news/{rubric}/{countNews}", api.newsHandler).Methods(http.MethodGet, http.MethodOptions)

	api.router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./ui"))))
}

// Базовый маршрут.
func (api *API) templateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("ui/html/base.html", "ui/html/routes.html"))

	// Отправляем HTML страницу с данными
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// Получение всех новостей.
func (api *API) newsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	rubric := vars["rubric"]
	countNewsStr := vars["countNews"]
	countNews, err := strconv.Atoi(countNewsStr)

	if err != nil {
		http.Error(w, "Invalid count parameter", http.StatusBadRequest)
		return
	}

	news, err := api.db.News(rubric, uint(countNews))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(news)
}
