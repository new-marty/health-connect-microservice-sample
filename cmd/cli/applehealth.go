package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var appleHealthCmd = &cobra.Command{
	Use:   "apple-health",
	Short: "Apple Health data (weight, vitals, ingest)",
}

var ahWeightCmd = &cobra.Command{
	Use:   "weight",
	Short: "Weight readings",
}

var ahWeightListCmd = &cobra.Command{
	Use:   "list",
	Short: "List weight readings",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/apple-health/weight", map[string]string{"from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ahWeightCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Log a weight reading",
	RunE: func(cmd *cobra.Command, args []string) error {
		weightKg, _ := cmd.Flags().GetFloat64("weight-kg")
		bodyFatPct, _ := cmd.Flags().GetFloat64("body-fat-pct")
		date, _ := cmd.Flags().GetString("date")
		source, _ := cmd.Flags().GetString("source")

		payload := map[string]interface{}{"weight_kg": weightKg}
		if bodyFatPct > 0 {
			payload["body_fat_pct"] = bodyFatPct
		}
		if date != "" {
			payload["date"] = date
		}
		if source != "" {
			payload["source"] = source
		}

		data, err := doPost("/api/v1/apple-health/weight", payload)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ahVitalsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List vitals",
	RunE: func(cmd *cobra.Command, args []string) error {
		metric, _ := cmd.Flags().GetString("metric")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/apple-health/vitals", map[string]string{"metric": metric, "from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var ahVitalsCmd = &cobra.Command{
	Use:   "vitals",
	Short: "Vitals (heart rate, steps, SpO2, etc.)",
}

var ahIngestCmd = &cobra.Command{
	Use:   "ingest <json-file>",
	Short: "Ingest Health Auto Export payload",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		var payload interface{}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		data, err := doPost("/api/v1/apple-health/ingest", payload)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	ahWeightListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	ahWeightListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	ahWeightCreateCmd.Flags().Float64("weight-kg", 0, "Weight in kg (required)")
	ahWeightCreateCmd.Flags().Float64("body-fat-pct", 0, "Body fat percentage")
	ahWeightCreateCmd.Flags().String("date", "", "Date (YYYY-MM-DD, defaults to today)")
	ahWeightCreateCmd.Flags().String("source", "", "Source (defaults to manual)")
	_ = ahWeightCreateCmd.MarkFlagRequired("weight-kg")

	ahVitalsListCmd.Flags().String("metric", "", "Metric name filter")
	ahVitalsListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	ahVitalsListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	ahWeightCmd.AddCommand(ahWeightListCmd, ahWeightCreateCmd)
	ahVitalsCmd.AddCommand(ahVitalsListCmd)
	appleHealthCmd.AddCommand(ahWeightCmd, ahVitalsCmd, ahIngestCmd)
}
