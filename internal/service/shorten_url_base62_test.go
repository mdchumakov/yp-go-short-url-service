package service

import (
	"math/big"
	"testing"
)

func Test_shortenURLBase62(t *testing.T) {
	type args struct {
		longURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with a sample URL",
			args: args{longURL: "https://example.com/some/long/url"},
			want: "4ZyG5E7z",
		},
		{
			name: "Testings with yandex practicum url",
			args: args{longURL: "https://practicum.yandex.ru/"},
			want: "1BYWBNb1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenURLBase62(tt.args.longURL)
			if got != tt.want {
				t.Errorf("shortenURLBase62() = %v, want %v", got, tt.want)
			}
			if gotLen := len(got); gotLen != shortURLSize {
				t.Errorf("shortenURLBase62() length = %v, want %v", got, shortURLSize)
			}
		})
	}
}

func Test_toBase62(t *testing.T) {
	type args struct {
		num *big.Int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with zero",
			args: args{num: big.NewInt(0)},
			want: "0",
		},
		{
			name: "Test with a small number",
			args: args{num: big.NewInt(123)},
			want: "1z",
		},
		{
			name: "Test with a large number",
			args: args{num: big.NewInt(123456789)},
			want: "8M0kX",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toBase62(tt.args.num); got != tt.want {
				t.Errorf("toBase62() = %v, want %v", got, tt.want)
			}
		})
	}
}
