package cmd

import (
	"os"
	"strings"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

const (
	// DefaultLang is the default language when searching subtitles.
	DefaultLang = "ENG"
)

var (
	paramLang = DefaultLang

	// Slice of languages selected from env. command-line, or DefaultLang
	paramLangs []string
)

// RootCmd is the main OSDB program command.
var RootCmd = &cobra.Command{
	Use:   "osdb",
	Short: "osdb is a command-line client for OpenSubtitles",
	Long:  "Search and download subtitles from the command-line.",
	PersistentPreRun: func(c *cobra.Command, args []string) {
		paramLangs = strings.Split(paramLang, ",")
	},
	Run: func(c *cobra.Command, args []string) {
	},
}

// InitClient returns a Client connected to OSDB API using env. vars OSDB_LOGIN,
// OSDB_PASSWORD.
func InitClient(lang string) (client *osdb.Client, err error) {
	if client, err = osdb.NewClient(); err != nil {
		return
	}
	if err = client.LogIn(os.Getenv("OSDB_LOGIN"), os.Getenv("OSDB_PASSWORD"), lang); err != nil {
		return
	}
	return
}

// GetEnvLang checks OSDB_LANG env. var, to set the API client language.
func GetEnvLang() string {
	if val, ok := os.LookupEnv("OSDB_LANG"); ok {
		return val
	}
	return DefaultLang
}
