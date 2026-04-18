package main

import (
	"github.com/spf13/cobra"
)

var mealsCmd = &cobra.Command{
	Use:   "meals",
	Short: "Meal logging",
}

var mealsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List meals",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		data, err := doGet("/api/v1/meals", map[string]string{"date": date, "from": from, "to": to})
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mealsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Log a meal",
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		meal, _ := cmd.Flags().GetString("meal")
		desc, _ := cmd.Flags().GetString("description")
		cal, _ := cmd.Flags().GetInt("calories")
		protein, _ := cmd.Flags().GetFloat64("protein-g")
		fat, _ := cmd.Flags().GetFloat64("fat-g")
		carbs, _ := cmd.Flags().GetFloat64("carbs-g")

		payload := map[string]interface{}{
			"date":        date,
			"description": desc,
		}
		if meal != "" {
			payload["meal"] = meal
		}
		if cal > 0 {
			payload["calories"] = cal
		}
		if protein > 0 {
			payload["protein_g"] = protein
		}
		if fat > 0 {
			payload["fat_g"] = fat
		}
		if carbs > 0 {
			payload["carbs_g"] = carbs
		}

		data, err := doPost("/api/v1/meals", payload)
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

var mealsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a meal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := doDelete("/api/v1/meals/" + args[0])
		if err != nil {
			return err
		}
		printJSON(data)
		return nil
	},
}

func init() {
	mealsListCmd.Flags().String("date", "", "Specific date (YYYY-MM-DD)")
	mealsListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	mealsListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	mealsCreateCmd.Flags().String("date", "", "Date (YYYY-MM-DD, required)")
	mealsCreateCmd.Flags().String("meal", "", "Meal type (breakfast, lunch, dinner, snack)")
	mealsCreateCmd.Flags().String("description", "", "Food description (required)")
	mealsCreateCmd.Flags().Int("calories", 0, "Calories")
	mealsCreateCmd.Flags().Float64("protein-g", 0, "Protein in grams")
	mealsCreateCmd.Flags().Float64("fat-g", 0, "Fat in grams")
	mealsCreateCmd.Flags().Float64("carbs-g", 0, "Carbs in grams")
	_ = mealsCreateCmd.MarkFlagRequired("date")
	_ = mealsCreateCmd.MarkFlagRequired("description")

	mealsCmd.AddCommand(mealsListCmd, mealsCreateCmd, mealsDeleteCmd)
}
