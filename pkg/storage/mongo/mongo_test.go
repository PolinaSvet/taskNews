package mongo

import (
	"GoNews/pkg/storage"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// проверяем работу пакета, используем тестовую БД:
// - TestStore_NewsImit
// - TestStore_AddNewImit

// тестовые данные
var (
	testNews1 = storage.News{
		Id:         1,
		Title:      "Title001",
		Content:    "Content001",
		PublicTime: time.Now().Unix(),
		ImageLink:  "database/image/imageSport.png",
		Rubric:     "Sport",
		Link:       "https://go.dev/play/" + time.Now().Format("2006-01-02 15:04:05"),
		LinkTitle:  "LinkTitle001",
	}
	strConnection = "NEWSDBMONGO_TEST" //используем тестовую БД
	strInform     = "MongoDB"
)

// проверяем соединение
func TestNew(t *testing.T) {
	t.Run("New mongo: Good (valid connection)", func(t *testing.T) {

		connstr := os.Getenv(strConnection)
		assert.NotNil(t, connstr)

		store, err := New(connstr)
		assert.NoError(t, err)
		assert.NotNil(t, store)
		assert.Equal(t, store.GetInform(), strInform)
		store.Close()
	})

	t.Run("New mongo: Error (invalid connection)", func(t *testing.T) {
		store, err := New("invalid_conn_string")
		assert.Error(t, err)
		assert.Nil(t, store)
	})
}

// получаем данные из бд
func TestStore_NewsAndAddNew(t *testing.T) {

	connstr := os.Getenv(strConnection)
	assert.NotNil(t, connstr)

	store, err := New(connstr)
	assert.NoError(t, err)
	defer store.Close()

	t.Run("AddNew mongo: Good (add one line with error)", func(t *testing.T) {
		err := store.AddNew([]storage.News{testNews1})
		assert.NoError(t, err)
	})

	t.Run("News mongo: Good (get records from bd)", func(t *testing.T) {
		news, err := store.News("Sport", 10)
		assert.NoError(t, err)
		// Проверка, что получен хотя бы один результат
		assert.NotEmpty(t, news)
	})
}
