package osdb

import (
	"bufio"
	"fmt"

	_ "github.com/orchestrate-io/dvr"
)

func ExampleClient_BestMovieByHashes() {
	c, err := NewClient()
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}

	err = c.LogIn("", "", "")
	if err != nil {
		fmt.Printf("can't login: %s\n", err)
		return
	}

	hashes := []uint64{0x09a2c497663259cb, 0x46e33be00464c12e}
	movies, err := c.BestMoviesByHashes(hashes)
	if err != nil {
		fmt.Printf("can't search: %s\n", err)
		return
	}

	for i, hash := range hashes {
		if movies[i] != nil {
			fmt.Printf("%016x: %s (%s) - id %s\n", hash,
				movies[i].Title, movies[i].Year, movies[i].Id)
		} else {
			fmt.Printf("%016x: unknown\n", hash)
		}
	}

	// Output:
	// 09a2c497663259cb: Nochnoy dozor (2004) - id 0403358
	// 46e33be00464c12e: "Game of Thrones" Two Swords (2014) - id 2816136
}

func ExampleClient_ImdbIdSearch() {
	c, err := NewClient()
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}

	err = c.LogIn("", "", "")
	if err != nil {
		fmt.Printf("can't login: %s\n", err)
		return
	}

	ids := []string{"0403358", "2816136"}
	langs := []string{"eng", "rus"}

	subs, err := c.ImdbIdSearch(ids, langs)
	if err != nil {
		fmt.Printf("can't search: %s\n", err)
		return
	}

	for _, sub := range subs {
		if sub.SubHash == "fb1a2837e1e6a4cceeb237154fac5f21" ||
			sub.SubHash == "51988e72deb96e1d1abd79ac8daf4b3b" {

			// check if that couple of top rated subtitles have been returned
			fmt.Printf("%s: %s\n", sub.IDMovieImdb, sub.SubFileName)
		}
	}

	// please note that sometimes Opensubtitles will remove leading zeroes!

	// Output:
	// 403358: Night.Watch.2004.720p.BluRay.x264-SiNNERS.srt
	// 2816136: Game of Thrones Season 4- Fire and Ice Foreshadowing.srt
}

func ExampleClient_DownloadSubtitles() {
	c, err := NewClient()
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}

	err = c.LogIn("", "", "")
	if err != nil {
		fmt.Printf("can't login: %s\n", err)
		return
	}

	ids := []int{1951968569, 1954123031}

	subFiles, err := c.DownloadSubtitles(ids)
	if err != nil {
		fmt.Printf("can't download subtitles: %s\n", err)
		return
	}

	for i, sf := range subFiles {
		reader, err := sf.Reader()
		if err != nil {
			fmt.Printf("can't open reader: %s\n", err)
			return
		}
		scanner := bufio.NewScanner(reader)
		for j := 0; j < 98; j++ {
			if !scanner.Scan() {
				fmt.Printf("too few lines in subtitle file\n")
				return
			}
		}
		if scanner.Scan() {
			fmt.Printf("99th line of subtitle %d: %s\n", i, scanner.Text())
		}
	}

	// Output:
	// 99th line of subtitle 0: ...and he knew that unless
	// 99th line of subtitle 1: and top it in many ways.
}

func ExampleClient_DownloadSubtitles_foreign() {
	c, err := NewClient()
	if err != nil {
		fmt.Printf("can't create client: %s\n", err)
		return
	}

	err = c.LogIn("", "", "")
	if err != nil {
		fmt.Printf("can't login: %s\n", err)
		return
	}

	ids := []int{1955009224}

	subFiles, err := c.DownloadSubtitles(ids)
	if err != nil {
		fmt.Printf("can't download subtitles: %s\n", err)
		return
	}

	for i, sf := range subFiles {
		reader, err := sf.Reader()
		if err != nil {
			fmt.Printf("can't open reader: %s\n", err)
			return
		}
		scanner := bufio.NewScanner(reader)
		for j := 0; j < 99; j++ {
			if !scanner.Scan() {
				fmt.Printf("too few lines in subtitle file\n")
				return
			}
		}
		if scanner.Scan() {
			fmt.Printf("99th line of subtitle %d: %s\n", i, scanner.Text())
		}
	}

	// Output:
	// 100th line of subtitle 0: Хорошо. Скоро увидимся.
}
