package cmd

import (
	"fmt"
	"os"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(putCmd)
}

var putCmd = &cobra.Command{
	Use:   "put [movie_file] [sub_file]",
	Short: "Upload subtitles for a file",
	Long:  `Submit new subtitles for a file.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("Invalid parameters.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		client, err := InitClient(os.Getenv("OSDB_LANG"))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		if err := putSubs(client, args[0], args[1]); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	},
}

func putSubs(client *osdb.Client, movieFile string, subFile string) (err error) {
	fmt.Println("- Checking file against OSDB...")
	subs, err := osdb.NewSubtitles(movieFile, []string{subFile})
	if err != nil {
		return
	}

	alreadyInDb, err := client.HasSubtitles(subs)
	if err != nil {
		return
	}
	if alreadyInDb == true {
		return fmt.Errorf("these subtitles already exist")
	}

	fmt.Println("- Uploading...")
	subURL, err := client.UploadSubtitles(subs)
	if err != nil {
		return
	}
	fmt.Printf("- Subtitles uploaded to %s\n", subURL)
	return
}
