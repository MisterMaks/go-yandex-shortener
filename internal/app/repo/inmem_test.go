package repo

import (
	"os"
	"sync"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/stretchr/testify/assert"
)

const (
	TestFilename string = "../../../test/internal_app_repo_inmem_test.txt"
)

func TestNewAppRepoInmem(t *testing.T) {
	appRepoInMem, err := NewAppRepoInmem(TestFilename)
	assert.NoError(t, err)
	assert.NotNil(t, appRepoInMem)
	os.Remove(TestFilename)
}

func TestAppRepoInmem_GetOrCreateURL(t *testing.T) {
	type fields struct {
		urls []*app.URL
	}
	type args struct {
		id     string
		rawURL string
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
			name: "get existed URL",
			fields: fields{
				urls: []*app.URL{{ID: "1", URL: "yandex.ru"}},
			},
			args: args{
				id:     "2",
				rawURL: "yandex.ru",
			},
			want: want{
				url: &app.URL{
					ID:  "1",
					URL: "yandex.ru",
				},
				wantErr: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := newProducer(TestFilename)
			if err != nil {
				t.Fatalf("CRITICAL\tUnexpected error. Error: %v\n", err)
			}
			ari := &AppRepoInmem{
				urls:     tt.fields.urls,
				mu:       sync.RWMutex{},
				producer: producer,
			}
			url, err := ari.GetOrCreateURL(tt.args.id, tt.args.rawURL)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.url, url)
			assert.Contains(t, ari.urls, url)
		})
	}
	os.Remove(TestFilename)
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
	os.Remove(TestFilename)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := newProducer(TestFilename)
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
	os.Remove(TestFilename)
}
