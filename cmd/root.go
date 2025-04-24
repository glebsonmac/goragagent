package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dataFile string
	rootCmd  = &cobra.Command{
		Use:   "goragagent",
		Short: "A tax information query system",
		Long: `GoragAgent is a CLI tool that helps you query tax information
using natural language processing and AI to provide accurate answers.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dataFile, "data", "data/data1.csv", "path to the CSV data file")
}
