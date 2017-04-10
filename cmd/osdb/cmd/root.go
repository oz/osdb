package cmd

import (
	"os"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

// RootCmd is the main OSDB program command.
var RootCmd = &cobra.Command{
	Use:   "osdb",
	Short: "osdb is a command-line client for OpenSubtitles",
	Long:  "Search and download subtitles from the command-line.",
	Run: func(c *cobra.Command, args []string) {
	},
}

// InitClient returns a Client connected to OSDB API using env. vars OSDB_LOGIN,
// OSDB_PASSWORD.
func InitClient(lang string) (client *osdb.Client, err error) {
	if client, err = osdb.NewClient(); err != nil {
		return
	}
	if err = client.LogIn(
		os.Getenv("OSDB_LOGIN"),
		os.Getenv("OSDB_PASSWORD"),
		lang); err != nil {
		return
	}
	return
}
