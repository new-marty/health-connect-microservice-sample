package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "health",
	Short: "Health Connect CLI",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiBase, "api", "http://localhost:8080", "API base URL")

	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(ouraCmd)
	rootCmd.AddCommand(stravaCmd)
	rootCmd.AddCommand(hevyCmd)
	rootCmd.AddCommand(inbodyCmd)
	rootCmd.AddCommand(appleHealthCmd)
	rootCmd.AddCommand(mealsCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(analysisCmd)
	rootCmd.AddCommand(syncCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
