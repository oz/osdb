package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/docopt/docopt-go"
	"github.com/oz/osdb"
)

// Program version
const VERSION = "0.2"

// Get an anonymous client connected to OSDB.
func getClient() (client *osdb.Client, err error) {
	if client, err = osdb.NewClient(); err != nil {
		return
	}
	if err = client.LogIn(
		os.Getenv("OSDB_LOGIN"),
		os.Getenv("OSDB_PASSWORD"),
		os.Getenv("OSDB_LANG")); err != nil {
		return
	}
	return
}

func getSubs(file string, lang string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fmt.Printf("- Getting %s subtitles for file: %s\n", lang, path.Base(file))
	subs, err := client.FileSearch(file, []string{lang})
	if err != nil {
		return err
	}
	if best := subs.Best(); best != nil {
		dest := file[0:len(file)-len(path.Ext(file))] + ".srt"
		fmt.Printf("- Downloading to: %s\n", dest)
		// FIXME check if dest exists instead of overwriting. O:)
		return client.DownloadTo(best, dest)
	}
	fmt.Println("- No subtitles found!")
	return nil
}

func putSubs(movieFile string, subFile string) error {
	fmt.Println("- Checking file against OSDB...")

	client, err := getClient()
	if err != nil {
		return err
	}
	alreadyInDb, err := client.HasSubtitlesForFiles(movieFile, subFile)
	if err != nil {
		return err
	}
	if alreadyInDb == true {
		fmt.Println("These subtitles already exist.")
	} else {
		fmt.Println("Uploading new subtitles... once the feature's implemented.")
	}
	return nil
}

func fileToSubtitle(file string) (s osdb.Subtitle, err error) {
	err = fmt.Errorf("Not implemented.")
	return
}

// Search movies on IMDB
func imdbSearch(q string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fmt.Printf("Searching %s on IMDB...\n\n", q)
	movies, err := client.IMDBSearch(q)
	if err != nil {
		return err
	}
	if movies.Empty() {
		fmt.Println("No results.")
	}
	for _, m := range movies {
		fmt.Printf("%s %s http://www.imdb.com/title/tt%s/\n", m.ID, m.Title, m.ID)
	}
	return nil
}

// Show IMDB movie details
func imdbShow(id string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	m, err := client.GetIMDBMovieDetails(id)
	if err != nil {
		return err
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "IMDB Id:\t%s\n", m.ID)
	fmt.Fprintf(w, "Title:\t%s\n", m.Title)
	fmt.Fprintf(w, "Year:\t%s\n", m.Year)
	fmt.Fprintf(w, "Duration:\t%s\n", m.Duration)
	fmt.Fprintf(w, "Cover:\t%s\n", m.Cover)
	fmt.Fprintf(w, "TagLine:\t%s\n", m.TagLine)
	fmt.Fprintf(w, "Plot:\t%s\n", m.Plot)
	fmt.Fprintf(w, "Goofs:\t%s\n", m.Goofs)
	fmt.Fprintf(w, "Trivia:\t%s\n", m.Trivia)
	return w.Flush()
}

func main() {
	usage := `
OSDB, an OpenSubtitles client.

Usage:
  osdb get [--lang=<lang>] <file>
  osdb (put|upload) <movie_file> <sub_file>
  osdb hash <file>
  osdb imdb show <movie id>
  osdb imdb <query>...
  osdb -h | --help
  osdb --version

Options:
	--lang=<lang>	Subtitles' languages, comma separated [default: ENG].
`
	arguments, err := docopt.Parse(usage, nil, true, "OSDB "+VERSION, false)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	// Figure out which language we want.
	l := os.Getenv("OSDB_LANG")
	if arguments["--lang"] != nil {
		l = arguments["--lang"].(string)
	}
	langs := strings.Split(l, ",")
	if len(langs) == 0 {
		langs = []string{"ENG"}
	}

	// Download subtitles
	if arguments["get"] == true {
		for _, l := range langs {
			if err = getSubs(arguments["<file>"].(string), l); err != nil {
				fmt.Printf("Error: %s\n", err)
			} else {
				break
			}
		}
	}

	// Upload subtitles
	if arguments["upload"] == true || arguments["put"] == true {
		if err = putSubs(arguments["<movie_file>"].(string), arguments["<sub_file>"].(string)); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	// Search IMDB
	if arguments["imdb"] == true {
		if arguments["show"] == true {
			if err = imdbShow(arguments["<movie id>"].(string)); err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		} else {
			query := strings.Join(arguments["<query>"].([]string), " ")
			if err = imdbSearch(query); err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		}
	}

	if arguments["hash"] == true {
		h, err := osdb.Hash(arguments["<file>"].(string))
		if err != nil {
			fmt.Printf("Error: %s", err)
		} else {
			fmt.Printf("%x", h)
		}
	}
}
