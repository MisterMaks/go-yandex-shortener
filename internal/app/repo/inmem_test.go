package repo

import (
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestFilenamePattern string = "internal_app_repo_inmem_test_*.json"
)

func TestNewAppRepoInmem(t *testing.T) {
	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	appRepoInMem, err := NewAppRepoInmem(tmpFile.Name(), tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, appRepoInMem)
}

func TestAppRepoInmem_GetOrCreateURL(t *testing.T) {
	type fields struct {
		urls []*app.URL
	}
	type args struct {
		id     string
		rawURL string
		userID uint
	}
	type want struct {
		url     *app.URL
		wantErr bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "create new URL",
			fields: fields{
				urls: nil,
			},
			args: args{
				id:     "1",
				rawURL: "yandex.ru",
				userID: 1,
			},
			want: want{
				url: &app.URL{
					ID:     "1",
					URL:    "yandex.ru",
					UserID: 1,
				},
				wantErr: false,
			},
		},
		{
			name: "get existed URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru", UserID: 1}},
			},
			args: args{
				id:     "2",
				rawURL: "yandex.ru",
				userID: 2,
			},
			want: want{
				url: &app.URL{
					ID:     "1",
					URL:    "yandex.ru",
					UserID: 1,
				},
				wantErr: false,
			},
		},
	}

	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := newProducer(tmpFile.Name())
			if err != nil {
				t.Fatalf("CRITICAL\tUnexpected error. Error: %v\n", err)
			}
			ari := &AppRepoInmem{
				urls:     tt.fields.urls,
				mu:       sync.RWMutex{},
				producer: producer,
			}
			url, err := ari.GetOrCreateURL(tt.args.id, tt.args.rawURL, tt.args.userID)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.url, url)
			assert.Contains(t, ari.urls, url)
		})
	}
}

func TestAppRepoInmem_GetURL(t *testing.T) {
	type fields struct {
		urls []*app.URL
	}
	type args struct {
		id string
	}
	type want struct {
		url     *app.URL
		wantErr bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "get existed URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru"}},
			},
			args: args{
				id: "1",
			},
			want: want{
				url: &app.URL{
					ID:  "1",
					URL: "yandex.ru",
				},
				wantErr: false,
			},
		},
		{
			name: "get non-existent URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru"}},
			},
			args: args{
				id: "2",
			},
			want: want{
				url:     nil,
				wantErr: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ari := &AppRepoInmem{
				urls: tt.fields.urls,
				mu:   sync.RWMutex{},
			}
			url, err := ari.GetURL(tt.args.id)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.url, url)
		})
	}
}

func TestAppRepoInmem_CheckIDExistence(t *testing.T) {
	type fields struct {
		urls []*app.URL
	}
	type args struct {
		id string
	}
	type want struct {
		checked bool
		wantErr bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "check existed URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru"}},
			},
			args: args{
				id: "1",
			},
			want: want{
				checked: true,
				wantErr: false,
			},
		},
		{
			name: "check non-existed URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru"}},
			},
			args: args{
				id: "2",
			},
			want: want{
				checked: false,
				wantErr: false,
			},
		},
	}

	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := newProducer(tmpFile.Name())
			if err != nil {
				t.Fatalf("CRITICAL\tUnexpected error. Error: %v\n", err)
			}
			ari := &AppRepoInmem{
				urls:     tt.fields.urls,
				mu:       sync.RWMutex{},
				producer: producer,
			}
			checked, err := ari.CheckIDExistence(tt.args.id)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.checked, checked)
		})
	}
}

func generateTestURLID(id, userID uint) string {
	return strconv.FormatUint(uint64(id), 10) + "_" + strconv.FormatUint(uint64(userID), 10)
}

func generateTestURLs(countUsers, countUserURLs uint) []*app.URL {
	defaultURL := "test_url"
	urls := make([]*app.URL, 0, countUserURLs*countUsers)

	for i := uint(0); i < countUsers; i++ {
		userID := i
		url := defaultURL + "_" + strconv.FormatUint(uint64(i), 10)

		for j := uint(0); j < countUserURLs; j++ {
			id := generateTestURLID(j, userID)
			urls = append(urls, &app.URL{
				ID:        id,
				URL:       url,
				UserID:    userID,
				IsDeleted: false,
			})
		}
	}

	return urls
}

func BenchmarkAppRepoInmem_GetOrCreateURL(b *testing.B) {
	urls := generateTestURLs(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		appRepoInmem, err := NewAppRepoInmem("", "")
		require.NoError(b, err)

		for _, url := range urls {
			b.StartTimer()
			_, err = appRepoInmem.GetOrCreateURL(url.ID, url.URL, url.UserID)
			b.StopTimer()
			require.NoError(b, err)
		}

		for _, url := range urls[1 : len(urls)-2] {
			b.StartTimer()
			_, err = appRepoInmem.GetOrCreateURL(url.ID, url.URL, url.UserID)
			b.StopTimer()
			require.NoError(b, err)
		}

		err = appRepoInmem.Close()
		require.NoError(b, err)
	}
}

func BenchmarkAppRepoInmem_GetOrCreateURLs(b *testing.B) {
	urls := generateTestURLs(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		appRepoInmem, err := NewAppRepoInmem("", "")
		require.NoError(b, err)

		b.StartTimer()
		_, err = appRepoInmem.GetOrCreateURLs(urls)
		_, err2 := appRepoInmem.GetOrCreateURLs(urls[1 : len(urls)-2])
		b.StopTimer()

		require.NoError(b, err)
		require.NoError(b, err2)

		err = appRepoInmem.Close()
		require.NoError(b, err)
	}
}

func BenchmarkAppRepoInmem_CheckIDExistence(b *testing.B) {
	urls := generateTestURLs(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		appRepoInmem, err := NewAppRepoInmem("", "")
		require.NoError(b, err)

		_, err = appRepoInmem.GetOrCreateURLs(urls)
		require.NoError(b, err)

		b.StartTimer()
		_, err = appRepoInmem.CheckIDExistence(urls[3].ID)
		_, err2 := appRepoInmem.CheckIDExistence("aaa")
		b.StopTimer()

		require.NoError(b, err)
		require.NoError(b, err2)

		err = appRepoInmem.Close()
		require.NoError(b, err)
	}
}

func BenchmarkAppRepoInmem_GetUserURLs(b *testing.B) {
	urls := generateTestURLs(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		appRepoInmem, err := NewAppRepoInmem("", "")
		require.NoError(b, err)

		_, err = appRepoInmem.GetOrCreateURLs(urls)
		require.NoError(b, err)

		b.StartTimer()
		_, err = appRepoInmem.GetUserURLs(urls[len(urls)/2].UserID)
		_, err2 := appRepoInmem.GetUserURLs(uint(len(urls)) * 2)
		b.StopTimer()

		require.NoError(b, err)
		require.NoError(b, err2)

		err = appRepoInmem.Close()
		require.NoError(b, err)
	}
}

func BenchmarkAppRepoInmem_DeleteUserURLs(b *testing.B) {
	urls := generateTestURLs(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		appRepoInmem, err := NewAppRepoInmem("", "")
		require.NoError(b, err)

		_, err = appRepoInmem.GetOrCreateURLs(urls)
		require.NoError(b, err)

		b.StartTimer()
		err = appRepoInmem.DeleteUserURLs(urls[3 : len(urls)-5])
		err2 := appRepoInmem.DeleteUserURLs([]*app.URL{{ID: "aaa", UserID: uint(len(urls)) * 2}})
		b.StopTimer()

		require.NoError(b, err)
		require.NoError(b, err2)

		err = appRepoInmem.Close()
		require.NoError(b, err)
	}
}
