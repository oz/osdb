package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

const (
	// DefaultLang is the default language when searching subtitles.
	DefaultLang = "ENG"
)

var (
	lang = DefaultLang
)

func init() {
	getCmd.Flags().StringVarP(&lang, "lang", "l", getEnvLang(), "Subtitle language")
	RootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [file]",
	Short: "Get subtitles for a file",
	Long:  `Download subtitles for a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		langs := strings.Split(lang, ",")
		for _, l := range langs {
			client, err := InitClient(l)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}

			for _, file := range args {
				if err := getSubs(client, file, l); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					break
				}
			}
		}
	},
}

func getEnvLang() string {
	if val, ok := os.LookupEnv("OSDB_LANG"); ok {
		return val
	}
	return DefaultLang
}

func getSubs(client *osdb.Client, file string, lang string) error {
	fmt.Printf("- Getting %s subtitles for file: %s\n", lang, path.Base(file))
	subs, err := client.FileSearch(file, []string{lang})
	if err != nil {
		return err
	}

	if best := subs.Best(); best != nil {
		dest := file[0:len(file)-len(path.Ext(file))] + ".srt"
		fmt.Printf("- Downloading to: %s\n", dest)
		// XXX check if dest exists instead of overwriting?
		return client.DownloadTo(best, dest)
	}

	fmt.Println("- No subtitles found!")
	return nil
}
