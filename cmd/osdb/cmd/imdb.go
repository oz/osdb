package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

var client *osdb.Client

func init() {
	imdbCmd.AddCommand(imdbShowCmd)
	RootCmd.AddCommand(imdbCmd)
}

var imdbCmd = &cobra.Command{
	Use:   "imdb [query]",
	Short: "Search IMDB",
	Long:  `Search IMDB for a movie, through OSDB's API.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		c, err := InitClient(os.Getenv("OSDB_LANG"))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		client = c
	},
	Run: func(cmd *cobra.Command, args []string) {
		q := strings.Join(args, " ")
		fmt.Printf("Searching %s on IMDB...\n\n", q)
		movies, err := client.IMDBSearch(q)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		if movies.Empty() {
			fmt.Println("No results.")
		}
		for _, m := range movies {
			fmt.Printf("%s %s http://www.imdb.com/title/tt%s/\n", m.ID, m.Title, m.ID)
		}
	},
}

var imdbShowCmd = &cobra.Command{
	Use:   "show [imdb_ids...]",
	Short: "Show movie details",
	Long:  `Display movie facts for an IMDB movie.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Missing movie ID")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, id := range args {
			m, err := client.GetIMDBMovieDetails(id)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			showMovieDetails(m)
		}
	},
}

func showMovieDetails(m *osdb.Movie) {
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
	w.Flush()
	fmt.Println()
}
