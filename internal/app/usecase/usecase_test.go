package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/MisterMaks/go-yandex-shortener/internal/app/usecase/mocks"
	"github.com/golang/mock/gomock"
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

func TestNewAppUsecase(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	m.EXPECT().DeleteUserURLs(gomock.Any()).Return(nil).AnyTimes()

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
					AppRepo:                       m,
					BaseURL:                       "http://example.com/",
					CountRegenerationsForLengthID: 1,
					LengthID:                      1,
					MaxLengthID:                   1,
					deleteURLsTicker:              time.NewTicker(5 * time.Second),
					deleteURLsChan:                make(chan *app.URL, 1024),
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
			appUsecase, err := NewAppUsecase(
				m,
				tt.args.resultAddrPrefix,
				tt.args.countRegenerationsForLengthID,
				tt.args.lengthID,
				tt.args.maxLengthID,
				nil,
				1024,
				5*time.Second,
			)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.EqualExportedValues(t, tt.want.appUsecase, appUsecase)
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
		userID uint
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
				userID: 1,
			},
			want: want{
				url: &app.URL{ID: TestURLID, URL: TestURL, UserID: 1},
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
				userID: 1,
			},
			want: want{
				url: nil,
				err: ErrMaxLengthIDLessLengthID,
			},
		},
		{
			name: "test 3",
			fields: fields{
				countRegenerationsForLengthID: 1,
				lengthID:                      0,
				maxLengthID:                   1,
			},
			args: args{
				rawURL: TestURL,
				userID: 1,
			},
			want: want{
				url: nil,
				err: ErrZeroLengthID,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	m.EXPECT().GetOrCreateURL(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_, rawURL string, userID uint) (*app.URL, error) {
		url := &app.URL{ID: TestURLID, URL: rawURL, UserID: userID}
		return url, nil
	}).AnyTimes()
	m.EXPECT().CheckIDExistence(TestURLID).Return(true, nil).AnyTimes()
	m.EXPECT().CheckIDExistence(gomock.Any()).Return(false, nil).AnyTimes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       m,
				CountRegenerationsForLengthID: tt.fields.countRegenerationsForLengthID,
				LengthID:                      tt.fields.lengthID,
				MaxLengthID:                   tt.fields.maxLengthID,
			}
			url, _, err := au.GetOrCreateURL(tt.args.rawURL, tt.args.userID)
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

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	m.EXPECT().GetURL(TestURLID).Return(&app.URL{
		ID:  TestURLID,
		URL: TestURL,
	}, nil).AnyTimes()
	m.EXPECT().GetURL(gomock.Any()).Return(nil, ErrTestIDNotFound)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       m,
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

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppRepoInterface(ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &AppUsecase{
				AppRepo:                       m,
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
