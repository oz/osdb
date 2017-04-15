package cmd

import (
	"fmt"
	"path"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

func init() {
	getCmd.Flags().StringVarP(&paramLang, "lang", "l", GetEnvLang(), "Subtitle language")
	RootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [file]",
	Short: "Get subtitles for a file",
	Long:  `Download subtitles for a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, l := range paramLangs {
			client, err := InitClient(l)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}

			for _, file := range args {
				if err := getSubs(client, file, l); err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					return
				}
			}
		}
	},
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
