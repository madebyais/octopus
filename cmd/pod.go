package cmd

import (
	"fmt"
	"strings"

	"github.com/madebyais/octopus/kubectl"
	"github.com/madebyais/octopus/util"
	"github.com/spf13/cobra"
)

var namespace string
var prefix string

func init() {
	podCmd := &cobra.Command{
		Use:   "pod",
		Short: "Run octopus k8s POD-related commands",
	}

	rootCmd.AddCommand(podCmd)

	podGetAllCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "set namespace")
	podCmd.AddCommand(podGetAllCmd)

	podGetByPrefix.Flags().StringVarP(&namespace, "namespace", "n", "default", "set namespace")
	podGetByPrefix.Flags().StringVarP(&prefix, "prefix", "p", "", "set pod prefix name (e.g. withdrawal-service)")
	podCmd.AddCommand(podGetByPrefix)
}

var podGetAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all existing pods",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		pod := kubectl.New()
		data, err := pod.GetAll(namespace)
		if err != nil {
			fmt.Printf("[ POD ] [ ERROR ] %s", err.Error())
			return
		}

		headers := []string{"NAMESPACE", "NODE", "SERVICE ID", "SERVICE ADDR"}
		var rows [][]string

		for _, item := range data {
			rows = append(rows, []string{
				item["namespace"],
				item["node"],
				item["serviceId"],
				item["serviceAddress"],
			})
		}

		util.GenerateTable(headers, rows)
	},
}

var podGetByPrefix = &cobra.Command{
	Use:   "get",
	Short: "Get existing pod by prefix",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		pod := kubectl.New()
		data, err := pod.GetByPrefix(namespace, prefix)
		if err != nil {
			fmt.Printf("[ POD ] [ ERROR ] %s", err.Error())
			return
		}

		headers := []string{"NAMESPACE", "NODE", "SERVICE ID", "SERVICE ADDR"}
		var rows [][]string

		for _, item := range data {
			rows = append(rows, []string{
				item["namespace"],
				item["node"],
				strings.Replace(item["serviceId"], prefix, fmt.Sprintf("\033[1;32m%s\033[0m", prefix), -1),
				item["serviceAddress"],
			})
		}

		util.GenerateTable(headers, rows)
	},
}
