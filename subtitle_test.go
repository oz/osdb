package osdb

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestBestWithNoResults(t *testing.T) {
	subs := Subtitles{}
	res := subs.Best()
	if res != nil {
		t.Fatalf("Expected nil, got: %v", res)
	}
}

func TestBestWithResults(t *testing.T) {
	subs := Subtitles{
		Subtitle{MovieName: "Go", SubDownloadsCnt: "1"},
		Subtitle{MovieName: "Never Let Me Go", SubDownloadsCnt: "2"},
		Subtitle{MovieName: "Don't Let Me Go", SubDownloadsCnt: "3"},
	}
	res := subs.Best()
	if res == nil {
		t.Fatalf("Expected Subtitle, got: %v", res)
	}
	if res.MovieName != "Don't Let Me Go" {
		t.Fatalf("Expected Don't Let Me Go, got: %v", res)
	}
}

func TestNewSubtitle(t *testing.T) {
	data := make([]byte, ChunkSize*2)

	// Generate dummy movie file
	copy(data, []byte("raw movie bytes go here"))
	movieHash := "8a4f474701fbf13e"
	err := ioutil.WriteFile("./test-movie.avi", data, 0644)
	if err != nil {
		t.Fatalf("Can't create test-movie.avi")
	}
	defer os.Remove("./test-movie.avi")

	// Generate dummy subtitles file
	copy(data, []byte("subtitle file data goes here"))
	subHash := "95ad809f6ee45b779abf6d5337efeb44"
	err = ioutil.WriteFile("./test-movie.srt", data, 0644)
	if err != nil {
		t.Fatalf("Can't create test-movie.srt")
	}
	defer os.Remove("./test-movie.srt")

	s, err := NewSubtitle("./test-movie.avi", "./test-movie.srt")
	if err != nil {
		t.Fatalf("Expected Subtitle, got error: %v", err)
	}

	if s.SubFileName != "test-movie.srt" {
		t.Fatalf("Expected SubFileName test-movie.srt, got %s", s.SubFileName)
	}
	if s.MovieFileName != "test-movie.avi" {
		t.Fatalf("Expected MovieFileName test-movie.avi, got %s", s.MovieFileName)
	}
	if s.SubHash != subHash {
		t.Fatalf("Expected subtitle Hash %s, got %s", subHash, s.SubHash)
	}
	if s.MovieHash != movieHash {
		t.Fatalf("Expected movie Hash %s, got %s", movieHash, s.MovieHash)
	}
}
