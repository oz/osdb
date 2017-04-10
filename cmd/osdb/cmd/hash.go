package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/oz/osdb"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(hashCmd)
}

var hashCmd = &cobra.Command{
	Use:   "hash [file]",
	Short: "Shows OSDB hash for file.",
	Long:  `Read file and compute its OSDB hash.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid parameters.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, file := range args {
			h, err := osdb.Hash(file)
			if err != nil {
				fmt.Printf("Error: %s", err)
			} else {
				basePath := path.Base(file)
				fmt.Printf("%s: %x\n", basePath, h)
			}
		}
	},
}
