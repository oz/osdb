package osdb

import "testing"

func TestBestWithNoResults(t *testing.T) {
	subs := Subtitles{}
	res := subs.Best()
	if res != nil {
		t.Fatalf("Expected nil, got: ", res)
	}
}

func TestBestWithResults(t *testing.T) {
	subs := Subtitles{
		Subtitle{MovieName: "Go"},
		Subtitle{MovieName: "Never Let Me Go"},
	}
	res := subs.Best()
	if res == nil {
		t.Fatalf("Expected Go Subtitle, got:", res)
	}
	if res.MovieName != "Go" {
		t.Fatalf("Expected Go, got:", res)
	}
}
