package repo

import (
	"os"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/stretchr/testify/require"
)

const (
	COUNT_URLS = 100000
)

func BenchmarkConsumerReadURLs(b *testing.B) {
	url := &app.URL{
		ID:        "1",
		URL:       "test_url",
		UserID:    1,
		IsDeleted: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer() // останавливаем таймер

		tmpFile, err := os.CreateTemp("", TestFilenamePattern)
		require.NoError(b, err)

		producer, err := newProducer(tmpFile.Name())
		require.NoError(b, err)

		for i := 0; i < COUNT_URLS; i++ {
			err = producer.writeURL(url)
			require.NoError(b, err)
		}

		consumer, err := newConsumer(tmpFile.Name())
		require.NoError(b, err)

		b.StartTimer() // возобновляем таймер
		consumer.readURLs()
		b.StopTimer() // останавливаем таймер

		consumer.close()
		os.Remove(tmpFile.Name())
	}
}
