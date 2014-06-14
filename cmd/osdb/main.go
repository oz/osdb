package main

import (
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"github.com/docopt/docopt.go"
	"github.com/oz/osdb"
)

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

func Get(file string, lang string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fmt.Printf("- Getting subtitles for file: %s\n", path.Base(file))
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
	return fmt.Errorf("No subtitles found!")
}

// Search movies on IMDB
func ImdbSearch(q string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	fmt.Printf("Searching %s on IMDB...\n\n", q)
	movies, err := client.SearchOnImdb(q)
	if err != nil {
		return err
	}
	if movies.Empty() {
		fmt.Println("No results.")
	}
	for _, m := range movies {
		fmt.Printf("%s http://www.imdb.com/title/tt%s/\n", m.Title, m.Id)
	}
	return nil
}

// Show IMDB movie details
func ImdbShow(id string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	m, err := client.GetImdbMovieDetails(id)
	if err != nil {
		return err
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintf(w, "IMDB Id:\t%s\n", m.Id)
	fmt.Fprintf(w, "Title:\t%s\n", m.Title)
	fmt.Fprintf(w, "Year:\t%s\n", m.Year)
	fmt.Fprintf(w, "Duration:\t%s\n", m.Duration)
	fmt.Fprintf(w, "Cover:\t%s\n", m.Cover)
	fmt.Fprintf(w, "TagLine:\t%s\n", m.TagLine)
	fmt.Fprintf(w, "Plot:\t%s\n", m.Plot)
	fmt.Fprintf(w, "Goofs:\t%s\n", m.Goofs)
	fmt.Fprintf(w, "Trivia:\t%s\n", m.Trivia)
	w.Flush()
	return nil
}

func main() {
	usage := `OSDB, an OpenSubtitles client.

Usage:
	osdb get [--language=<lang>] <file>
	osdb imdb <query>
	osdb imdb show <movie id>
	osdb -h | --help
	osdb --version

Options:
	--language=<lang>	Subtitles' language [default: ENG].
`
	arguments, err := docopt.Parse(usage, nil, true, "OSDB 0.1a", false)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	lang := "ENG"
	if arguments["--language"] != nil {
		lang = arguments["--language"].(string)
	}

	if arguments["get"] == true {
		if err = Get(arguments["<file>"].(string), lang); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}

	if arguments["imdb"] == true {
		if arguments["show"] == true {
			if err = ImdbShow(arguments["<movie id>"].(string)); err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		} else {
			if err = ImdbSearch(arguments["<query>"].(string)); err != nil {
				fmt.Printf("Error: %s\n", err)
			}
		}
	}
}
