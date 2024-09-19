package main

import (
	"GoNews/pkg/storage"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моковый хранилище для тестирования
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) News(rubric string, count uint) ([]storage.News, error) {
	args := m.Called(rubric, count)
	return args.Get(0).([]storage.News), args.Error(1)
}

func (m *MockStorage) AddNew(news []storage.News) error {
	args := m.Called(news)
	return args.Error(0)
}

func (m *MockStorage) GetInform() string {
	return "TestStore"
}

func (m *MockStorage) Close() {}

func Test_getNewsFromAllRSS(t *testing.T) {
	newsChannel := make(chan []storage.News)
	errorChannel := make(chan error)

	// Конфигурация RSS
	configRSS := ConfigRSS{
		RSS: map[string]struct {
			Link  []string `json:"link"`
			Image string   `json:"image"`
		}{
			"Sport": {
				Link:  []string{"http://news.mail.ru/rss/sport/91/"},
				Image: "database/image/imageSport.png",
			},
		},
		Duration: 1, // Для упрощения тестов
	}

	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // cancel when we are finished consuming integers

	go getNewsFromAllRSS(ctx, configRSS, newsChannel, errorChannel)

	select {
	case news := <-newsChannel:
		assert.NotNil(t, news)
	case err := <-errorChannel:
		assert.NotNil(t, err)
	case <-time.After(time.Second * 3):
		t.Error("Timeout waiting for news")
	}

	close(errorChannel)
	close(newsChannel)

}

func Test_writeNewsToDB(t *testing.T) {
	// Создайте mock-канал и mock-хранилище для имитации зависимостей
	newsChannel := make(chan []storage.News)
	errorChannel := make(chan error)
	mockStorage := new(MockStorage)
	mockStorage.On("AddNew", mock.Anything).Return(nil)

	// Создание контекста с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // cancel when we are finished consuming integers

	// Проверка вызова метода AddNew в mock-хранилище
	newsBatch := []storage.News{
		{
			Title:      "Title001",
			Content:    "Content001",
			PublicTime: time.Now().Unix(),
			ImageLink:  "database/image/imageSport.png",
			Rubric:     "Sport",
			Link:       "https://go.dev/play/" + time.Now().Format("2006-01-02 15:04:05"),
			LinkTitle:  "LinkTitle001",
		},
	}

	go writeNewsToDB(ctx, mockStorage, newsChannel, errorChannel)

	// Проверка вызова метода AddNew в mock-хранилище
	newsChannel <- newsBatch
	time.Sleep(time.Second * 1) // Даем время для обработки данных
	mockStorage.AssertCalled(t, "AddNew", newsBatch)

	close(errorChannel)
	close(newsChannel)
}

func Test_handleErrors(t *testing.T) {
	t.Run("handle errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errs := make(chan error, 3)
		errs <- errors.New("error 1")
		errs <- errors.New("error 2")
		errs <- errors.New("error 3")
		close(errs)

		go handleErrors(ctx, errs)

		time.Sleep(2 * time.Second)

		if ctx.Err() != nil {
			t.Errorf("expected context to not be done, got: %v", ctx.Err())
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		errs := make(chan error, 3)
		errs <- errors.New("error 1")
		errs <- errors.New("error 2")
		errs <- errors.New("error 3")
		close(errs)

		go handleErrors(ctx, errs)

		time.Sleep(3 * time.Second)

		if ctx.Err() == nil {
			t.Errorf("expected context to be done")
		}
	})
}

func TestConfigRSSUnmarshal(t *testing.T) {
	// Тест для проверки десериализации конфигурации RSS
	data, err := ioutil.ReadFile("./config.json")
	assert.NoError(t, err)

	var configRSS ConfigRSS
	err = json.Unmarshal(data, &configRSS)
	assert.NoError(t, err)

	assert.NotEmpty(t, configRSS.RSS)
	assert.Equal(t, 1, configRSS.Duration)
}
