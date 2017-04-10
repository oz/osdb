package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

// Version is the program's version. Yep.
const Version = "0.3"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of osdb",
	Long:  `All software has versions. This is osdb's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OSDB v%s\n", Version)
	},
}
