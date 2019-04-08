package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/madebyais/octopus/consul"
	"github.com/madebyais/octopus/util"
	"github.com/spf13/cobra"
)

var host string
var datacenter string
var node string
var serviceName string
var podNamePrefix string
var podNamespace string

func init() {
	consulCmd := &cobra.Command{
		Use:   "consul",
		Short: "Run octopus CONSUL-related commands",
	}

	rootCmd.AddCommand(consulCmd)

	consulCmd.PersistentFlags().StringVarP(&host, "host", "", "localhost:8500", "set consul host addr")

	consulServicesCmd.AddCommand(consulServicesGetAllServiceCmd)
	consulServicesCmd.AddCommand(consulServicesGetServiceDetailCmd)
	consulCmd.AddCommand(consulServicesCmd)

	consulCheckRegisteredPod.Flags().StringVarP(&serviceName, "service", "s", "", "set service name for consul")
	consulCheckRegisteredPod.Flags().StringVarP(&podNamespace, "namespace", "n", "default", "set namespace")
	consulCheckRegisteredPod.Flags().StringVarP(&podNamePrefix, "pod-prefix", "p", "", "set pod name prefix")
	consulCmd.AddCommand(consulCheckRegisteredPod)

	consulDeregister.Flags().StringVarP(&datacenter, "datacenter", "d", "dc1", "set consul datacenter")
	consulDeregister.Flags().StringVarP(&node, "node", "n", "", "set consul node")
	consulDeregister.Flags().StringVarP(&serviceName, "service", "s", "", "set service name registered in consul")
	consulCmd.AddCommand(consulDeregister)
}

var consulServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Get list of services from consul",
}

var consulServicesGetAllServiceCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all service registered",
	Run: func(cmd *cobra.Command, args []string) {
		consulCmd := consul.New(host)
		data, err := consulCmd.GetServices()
		if err != nil {
			fmt.Printf("[ CONSUL ] [ ERROR ] %s", err.Error())
			return
		}

		headers := []string{"SERVICES"}
		var rows [][]string

		sort.Strings(data)
		for _, val := range data {
			rows = append(rows, []string{val})
		}

		util.GenerateTable(headers, rows)
	},
}

var consulServicesGetServiceDetailCmd = &cobra.Command{
	Use:   "get",
	Short: "Get service detail by service name",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}

		serviceName = args[0]

		consulCmd := consul.New(host)
		data, err := consulCmd.GetServiceDetail(serviceName)
		if err != nil {
			fmt.Printf("[ CONSUL ] [ ERROR ] %s", err.Error())
			return
		}

		headers := []string{"NODE", "ADDR", "SERVICE ID", "SERVICE ADDR", "SERVICE PORT"}
		var rows [][]string
		for _, item := range data {
			rows = append(rows, []string{
				item["node"].(string),
				item["address"].(string),
				strings.Replace(item["serviceId"].(string), serviceName, fmt.Sprintf("\033[1;32m%s\033[0m", serviceName), -1),
				item["serviceAddress"].(string),
				item["servicePort"].(string),
			})
		}

		util.GenerateTable(headers, rows)
	},
}

var consulCheckRegisteredPod = &cobra.Command{
	Use:   "check",
	Short: "Check registered pod",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		consulCmd := consul.New(host)
		data, err := consulCmd.CheckRegisteredPod(serviceName, podNamespace, podNamePrefix)
		if err != nil {
			fmt.Printf("[ CONSUL ] [ ERROR ] %s", err.Error())
			return
		}

		headers := []string{"NODE", "SERVICE ID", "SERVICE ADDR", "POD SERVICE ADDR", "POD SERVICE ID"}
		var rows [][]string
		for _, item := range data {
			rows = append(rows, []string{
				item["node"],
				item["serviceId"],
				item["serviceAddress"],
				item["podServiceAddress"],
				item["podServiceId"],
			})
		}

		util.GenerateTable(headers, rows)
	},
}

var consulDeregister = &cobra.Command{
	Use:   "deregister",
	Short: "Deregister existing services",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if node == "" || serviceName == "" {
			return errors.New("please specify node or service name ")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		consulCmd := consul.New(host)
		err := consulCmd.DeregisterService(datacenter, node, serviceName)
		if err != nil {
			log("[ ERROR ] " + err.Error())
			return
		}

		log("Service has been deregistered successfully, datacenter=" + datacenter + " node=" + node + " service=" + serviceName)
	},
}

func log(text string) {
	fmt.Printf("\n[ CONSUL ] %s", text)
	fmt.Println(``)
}
