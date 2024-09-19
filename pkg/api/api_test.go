package api

import (
	"GoNews/pkg/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Моковый хранилище для тестирования
type MockStore struct {
	news []storage.News
}

func (ms *MockStore) News(rubric string, countNews uint) ([]storage.News, error) {
	return ms.news, nil
}

func (ms *MockStore) AddNew(news []storage.News) error {
	ms.news = append(ms.news, news...)
	return nil
}

func (ts *MockStore) GetInform() string {
	return "TestStore"
}

func (ms *MockStore) Close() {}

func TestAPI_newsHandler(t *testing.T) {
	// Подготовка тестовых данных
	testNews := []storage.News{
		{
			Id:         1,
			Title:      "Title001",
			Content:    "Content001",
			PublicTime: time.Now().Unix(),
			ImageLink:  "database/image/imageSport.png",
			Rubric:     "Sport",
			Link:       "https://example.com/news1",
			LinkTitle:  "LinkTitle001",
		},
		{
			Id:         2,
			Title:      "Title002",
			Content:    "Content002",
			PublicTime: time.Now().Unix() - 1000,
			ImageLink:  "database/image/imageSport.png",
			Rubric:     "Sport",
			Link:       "https://example.com/news2",
			LinkTitle:  "LinkTitle002",
		},
	}

	// Создание мокового хранилища
	mockStore := &MockStore{
		news: testNews,
	}

	// Создание экземпляра API
	api := New(mockStore)

	t.Run("GetNews", func(t *testing.T) {
		// Создание тестового запроса
		req := httptest.NewRequest(http.MethodGet, "/news/Sport/2", nil)
		// Создание записи для ответа
		rr := httptest.NewRecorder()
		// Обработка запроса
		api.router.ServeHTTP(rr, req)

		// Проверка ответа
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		// Декодирование JSON-ответа
		var response []storage.News
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		// Проверка полученных данных
		assert.Equal(t, 2, len(response))
		assert.Equal(t, "Title001", response[0].Title)
		assert.Equal(t, "Title002", response[1].Title)
	})

	t.Run("InvalidCountParameter", func(t *testing.T) {
		// Создание тестового запроса с неверным параметром count
		req := httptest.NewRequest(http.MethodGet, "/news/Sport/invalid", nil)
		rr := httptest.NewRecorder()
		api.router.ServeHTTP(rr, req)

		// Проверка ответа
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAPI_templateHandler(t *testing.T) {

	// Создание экземпляра API
	api := New(&MockStore{})

	// Создание временного файла для тестирования
	baseHTML := "ui/html/base.html"
	routesHTML := "ui/html/routes.html"

	// Создаем содержимое файла
	err := os.MkdirAll("ui/html", 0755)
	if err != nil {
		t.Fatalf("failed to create ui/html dir: %v", err)
	}
	defer os.RemoveAll("ui/html")

	// Определяем шаблон "base"
	baseTemplate := `
		{{define "base"}}
		<!DOCTYPE html>
		<html>
		<head>
			<title>News</title>
		</head>
		<body>
			<main>
				{{block "content" .}} {{end}}
			</main>
		</body>
		</html>
		{{end}}
		`
	err = os.WriteFile(baseHTML, []byte(baseTemplate), 0644)
	if err != nil {
		t.Fatalf("failed to write base.html: %v", err)
	}
	// Определяем шаблон "routes"
	routesTemplate := `
		{{define "content"}}
		<div class="wrapper">
			<p>Routes</p>
		</div>
		{{end}}
		`
	err = os.WriteFile(routesHTML, []byte(routesTemplate), 0644)
	if err != nil {
		t.Fatalf("failed to write routes.html: %v", err)
	}

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Обработка запроса
	api.templateHandler(rr, req)

	// Проверка ответа
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "<!DOCTYPE html>")
	assert.Contains(t, rr.Body.String(), "<title>News</title>")
	assert.Contains(t, rr.Body.String(), "Routes")
}

func TestAPI_corsMiddleware(t *testing.T) {
	// Создание мокового обработчика
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello from mock handler")
	})

	// Создание middleware
	handler := corsMiddleware(mockHandler)

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Обработка запроса
	handler.ServeHTTP(rr, req)

	// Проверка заголовков ответа
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestAPI_errorMiddleware(t *testing.T) {
	// Создание мокового обработчика, который генерирует ошибку
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("mock error"))
	})

	// Создание middleware
	handler := errorMiddleware(mockHandler)

	// Создание тестового запроса
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Обработка запроса
	handler.ServeHTTP(rr, req)

	// Проверка ответа
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "Internal Server Error\n", rr.Body.String())
}

func TestAPI_Router(t *testing.T) {
	api := New(&MockStore{})
	assert.NotNil(t, api.Router())
}

func TestAPI_New(t *testing.T) {
	mockStore := &MockStore{}
	api := New(mockStore)
	assert.NotNil(t, api)
	assert.Equal(t, mockStore, api.db)
}
