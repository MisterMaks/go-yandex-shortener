package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewURL(t *testing.T) {
	type args struct {
		id     string
		rawURL string
	}
	type want struct {
		url     *URL
		wantErr bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid url",
			args: args{
				id:     "1",
				rawURL: "example.com",
			},
			want: want{
				url:     &URL{ID: "1", URL: "example.com"},
				wantErr: false,
			},
		},
		{
			name: "invalid url",
			args: args{
				id:     "1",
				rawURL: "invalid url.com",
			},
			want: want{
				url:     nil,
				wantErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := NewURL(tt.args.id, tt.args.rawURL)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.url, url)
		})
	}
}
