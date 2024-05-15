package usecase

import (
	"errors"
	"testing"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/stretchr/testify/assert"
)

const (
	TestAddr  string = "localhost:8080"
	TestURLID string = "1"
	TestURL   string = "example.com"
)

var (
	ErrTestIDNotFound = errors.New("ID not found")
)

func Test_generateID(t *testing.T) {
	type want struct {
		length int
		err    error
	}

	tests := []struct {
		name   string
		length uint
		want   want
	}{
		{
			name:   "test 1",
			length: 5,
			want: want{
				length: 5,
				err:    nil,
			},
		},
		{
			name:   "test 2",
			length: 10,
			want: want{
				length: 10,
				err:    nil,
			},
		},
		{
			name:   "invalid length ID",
			length: 0,
			want: want{
				length: 0,
				err:    ErrZeroLengthID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := generateID(tt.length)
			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.length, len(id))
		})
	}
}

type testAppRepo struct{}

func (tar *testAppRepo) GetOrCreateURL(id, rawURL string) (*app.URL, error) {
	url, err := app.NewURL(TestURLID, rawURL)
	return url, err
}

func (tar *testAppRepo) GetURL(id string) (*app.URL, error) {
	switch id {
	case TestURLID:
		return &app.URL{
			ID:  TestURLID,
			URL: TestURL,
		}, nil
	}
	return nil, ErrTestIDNotFound
}

func (tar *testAppRepo) CheckIDExistence(id string) (bool, error) {
	return id == TestURLID, nil
}

func TestNewAppUsecase(t *testing.T) {
	type args struct {
		resultAddrPrefix              string
		countRegenerationsForLengthID uint
		lengthID                      uint
		maxLengthID                   uint
	}
	type want struct {
		appUsecase *AppUsecase
		wantErr    bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				resultAddrPrefix:              "http://example.com/",
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			want: want{
				appUsecase: &AppUsecase{
					AppRepo:                       &testAppRepo{},
					BaseURL:                       "http://example.com/",
					CountRegenerationsForLengthID: 1,
					LengthID:                      1,
					MaxLengthID:                   1,
				},
				wantErr: false,
			},
		},
		{
			name: "invalid length ID",
			args: args{
				resultAddrPrefix:              "http://example.com/",
				countRegenerationsForLengthID: 1,
				lengthID:                      0,
				maxLengthID:                   1,
			},
			want: want{
				appUsecase: nil,
				wantErr:    true,
			},
		},
		{
			name: "invalid max length ID",
			args: args{
				resultAddrPrefix:              "http://example.com/",
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   0,
			},
			want: want{
				appUsecase: nil,
				wantErr:    true,
			},
		},
		{
			name: "invalid max length ID with length ID",
			args: args{
				resultAddrPrefix:              "http://example.com/",
				countRegenerationsForLengthID: 1,
				lengthID:                      3,
				maxLengthID:                   2,
			},
			want: want{
				appUsecase: nil,
				wantErr:    true,
			},
		},
		{
			name: "invalid prefix of the resulting address",
			args: args{
				resultAddrPrefix:              "invalid prefix of the resulting address",
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			want: want{
				appUsecase: nil,
				wantErr:    true,
			},
		},
		{
			name: "invalid prefix of the resulting address 2",
			args: args{
				resultAddrPrefix:              "http://example.com",
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			want: want{
				appUsecase: nil,
				wantErr:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tar := &testAppRepo{}
			appUsecase, err := NewAppUsecase(tar, tt.args.resultAddrPrefix, tt.args.countRegenerationsForLengthID, tt.args.lengthID, tt.args.maxLengthID)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.appUsecase, appUsecase)
		})
	}
}

func TestAppUsecase_GetOrCreateURL(t *testing.T) {
	type fields struct {
		countRegenerationsForLengthID uint
		lengthID                      uint
		maxLengthID                   uint
	}
	type args struct {
		rawURL string
	}
	type want struct {
		url *app.URL
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "test 1",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			args: args{
				rawURL: TestURL,
			},
			want: want{
				url: &app.URL{ID: TestURLID, URL: TestURL},
				err: nil,
			},
		},
		{
			name: "test 2",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      2,
				maxLengthID:                   1,
			},
			args: args{
				rawURL: TestURL,
			},
			want: want{
				url: nil,
				err: ErrMaxLengthIDLessLengthID,
			},
		},
		{
			name: "test 2",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      0,
				maxLengthID:                   1,
			},
			args: args{
				rawURL: TestURL,
			},
			want: want{
				url: nil,
				err: ErrZeroLengthID,
			},
		},
	}

	tar := &testAppRepo{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       tar,
				CountRegenerationsForLengthID: tt.fields.countRegenerationsForLengthID,
				LengthID:                      tt.fields.lengthID,
				MaxLengthID:                   tt.fields.maxLengthID,
			}
			url, err := au.GetOrCreateURL(tt.args.rawURL)
			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.url, url)
		})
	}
}

func TestAppUsecase_GetURL(t *testing.T) {
	type fields struct {
		countRegenerationsForLengthID uint
		lengthID                      uint
		maxLengthID                   uint
	}
	type args struct {
		id string
	}
	type want struct {
		url *app.URL
		err error
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "existing ID",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			args: args{
				id: TestURLID,
			},
			want: want{
				url: &app.URL{
					ID:  TestURLID,
					URL: TestURL,
				},
				err: nil,
			},
		},
		{
			name: "non-existent ID",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			args: args{
				id: "non-existent ID",
			},
			want: want{
				url: nil,
				err: ErrTestIDNotFound,
			},
		},
	}

	tar := &testAppRepo{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       tar,
				CountRegenerationsForLengthID: tt.fields.countRegenerationsForLengthID,
				LengthID:                      tt.fields.lengthID,
				MaxLengthID:                   tt.fields.maxLengthID,
			}
			url, err := au.GetURL(tt.args.id)
			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.url, url)
		})
	}
}

func TestAppUsecase_GenerateShortURL(t *testing.T) {
	type fields struct {
		countRegenerationsForLengthID uint
		lengthID                      uint
		maxLengthID                   uint
	}
	type args struct {
		addr string
		id   string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test 1",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			args: args{
				addr: TestAddr,
				id:   TestURLID,
			},
			want: "http://example.com/1",
		},
		{
			name: "test 2",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      1,
				maxLengthID:                   1,
			},
			args: args{
				addr: "example.com",
				id:   "2",
			},
			want: "http://example.com/2",
		},
	}

	tar := &testAppRepo{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       tar,
				BaseURL:                       "http://example.com/",
				CountRegenerationsForLengthID: tt.fields.countRegenerationsForLengthID,
				LengthID:                      tt.fields.lengthID,
				MaxLengthID:                   tt.fields.maxLengthID,
			}
			shortURL := au.GenerateShortURL(tt.args.id)
			assert.Equal(t, tt.want, shortURL)
		})
	}
}
