package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
	filetype "gopkg.in/h2non/filetype.v1"
)

var NoSub = errors.New("No subtitles found!")

func init() {
	getCmd.Flags().StringVarP(&paramLang, "lang", "l", GetEnvLang(), "Subtitle language")
	RootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [file/directory]",
	Short: "Get subtitles for a file or for all files in a directory.",
	Long:  `Download subtitles for a file or for all files in a directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, l := range paramLangs {
			client, err := InitClient(l)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
			}
			if len(args) == 1 {
				x, err := os.Stat(args[0])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					break
				} else if x.IsDir() {
					args = getFilesFromPath(args[0])
				}
			}
			for _, file := range args {
				if err := getSubs(client, file, l); err != nil {
					if err != NoSub {
						fmt.Printf("Error: %s\n", err)
						return
					}
					fmt.Println(err)
					continue
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
	return NoSub
}

func getFilesFromPath(dir string) []string {
	files := []string{}
	entries, _ := ioutil.ReadDir(dir)
	for _, e := range entries {
		file := path.Join(dir, e.Name())
		buf, _ := ioutil.ReadFile(file)
		if filetype.IsVideo(buf) {
			files = append(files, file)
		}
	}
	return files
}
