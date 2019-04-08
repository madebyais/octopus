package consul

import (
	"errors"
	"fmt"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/madebyais/octopus/kubectl"
	"github.com/madebyais/octopus/util"
)

// ICommand is interface for command
type ICommand interface {
	GetServices() ([]string, error)
	GetServiceDetail(serviceName string) ([]map[string]interface{}, error)
	CheckRegisteredPod(serviceName, namespace, podPrefix string) ([]map[string]string, error)
	DeregisterService(datacenter, node, serviceName string) error
}

// Command ...
type Command struct {
	ConsulClient *consulapi.Client
	Host         string
}

// New returns interface of command
func New(addr string) ICommand {
	if addr == "" {
		addr = "localhost:8500"
	}

	client, err := consulapi.NewClient(&consulapi.Config{
		Address: addr,
	})

	if err != nil {
		panic(err)
	}

	return &Command{
		ConsulClient: client,
		Host:         addr,
	}
}

// GetServices returns all services registered to consul
func (cmd *Command) GetServices() ([]string, error) {
	data, _, err := cmd.ConsulClient.Catalog().Services(&consulapi.QueryOptions{})
	if err != nil {
		return nil, err
	}

	var services []string
	for key, _ := range data {
		services = append(services, key)
	}

	return services, nil
}

// GetServiceDetail returns service details get by service name
func (cmd *Command) GetServiceDetail(serviceName string) ([]map[string]interface{}, error) {
	data, _, err := cmd.ConsulClient.Catalog().Service(serviceName, "", nil)
	if err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	for _, item := range data {
		tempItem := make(map[string]interface{})
		tempItem["id"] = item.ID
		tempItem["node"] = item.Node
		tempItem["address"] = item.Address
		tempItem["serviceId"] = item.ServiceID
		tempItem["serviceAddress"] = item.ServiceAddress
		tempItem["servicePort"] = strconv.Itoa(item.ServicePort)

		items = append(items, tempItem)
	}

	return items, nil
}

// CheckRegisteredPod return list of registered pod and mark it as red if it doesn't exist in pod
func (cmd *Command) CheckRegisteredPod(serviceName, namespace, prefix string) ([]map[string]string, error) {
	var items []map[string]string

	services, err := cmd.GetServiceDetail(serviceName)
	if err != nil {
		return items, err
	}

	kubectl := kubectl.New()
	pods, err := kubectl.GetByPrefix(namespace, prefix)
	if err != nil {
		return items, err
	}

	tmpServiceData := make(map[string]map[string]string)
	for _, service := range services {
		key := fmt.Sprintf("%s_%s_%s", service["node"].(string), service["serviceId"].(string), service["serviceAddress"].(string))
		if _, ok := tmpServiceData[key]; !ok {
			tmpServiceData[key] = make(map[string]string)
			tmpServiceData[key]["node"] = service["node"].(string)
			tmpServiceData[key]["serviceId"] = service["serviceId"].(string)
			tmpServiceData[key]["serviceAddress"] = service["serviceAddress"].(string)
		}
	}

	tmpPodData := make(map[string]map[string]string)
	for _, pod := range pods {
		key := fmt.Sprintf("%s_%s_%s", pod["node"], pod["serviceId"], pod["serviceAddress"])
		if _, ok := tmpPodData[key]; !ok {
			tmpPodData[key] = make(map[string]string)
			tmpPodData[key]["node"] = pod["node"]
			tmpPodData[key]["serviceId"] = pod["serviceId"]
			tmpPodData[key]["serviceAddress"] = pod["serviceAddress"]
		}

		key = fmt.Sprintf("%s_%s_%s", pod["node"], pod["serviceId"]+"-"+prefix, pod["serviceAddress"])
		if _, ok := tmpPodData[key]; !ok {
			tmpPodData[key] = make(map[string]string)
			tmpPodData[key]["node"] = pod["node"]
			tmpPodData[key]["serviceId"] = pod["serviceId"]
			tmpPodData[key]["serviceAddress"] = pod["serviceAddress"]
		}
	}

	for key, serviceValue := range tmpServiceData {
		tempItem := make(map[string]string)
		if podValue, ok := tmpPodData[key]; ok {
			tempItem["node"] = fmt.Sprintf("\033[1;32m%s\033[0m", serviceValue["node"])
			tempItem["serviceId"] = fmt.Sprintf("\033[1;32m%s\033[0m", serviceValue["serviceId"])
			tempItem["serviceAddress"] = fmt.Sprintf("\033[1;32m%s\033[0m", serviceValue["serviceAddress"])
			tempItem["podServiceId"] = fmt.Sprintf("\033[1;32m%s\033[0m", podValue["serviceId"])
			tempItem["podServiceAddress"] = fmt.Sprintf("\033[1;32m%s\033[0m", podValue["serviceAddress"])
		} else {
			tempItem["node"] = fmt.Sprintf("\033[1;31m%s\033[0m", serviceValue["node"])
			tempItem["serviceId"] = fmt.Sprintf("\033[1;31m%s\033[0m", serviceValue["serviceId"])
			tempItem["serviceAddress"] = fmt.Sprintf("\033[1;31m%s\033[0m", serviceValue["serviceAddress"])
			tempItem["podServiceId"] = ""
			tempItem["podServiceAddress"] = ""
		}

		items = append(items, tempItem)
	}

	for key, podValue := range tmpPodData {
		if _, ok := tmpServiceData[key]; !ok {
			tempItem := make(map[string]string)
			tempItem["node"] = fmt.Sprintf("\033[1;31m%s\033[0m", podValue["node"])
			tempItem["serviceId"] = ""
			tempItem["serviceAddress"] = ""
			tempItem["podServiceId"] = fmt.Sprintf("\033[1;31m%s\033[0m", podValue["serviceId"])
			tempItem["podServiceAddress"] = fmt.Sprintf("\033[1;31m%s\033[0m", podValue["serviceAddress"])

			items = append(items, tempItem)
		}
	}

	return items, nil
}

// DeregisterService is used to deregister a service from consul
func (cmd *Command) DeregisterService(datacenter, node, serviceName string) error {
	restAPI := util.NewRestAPI()
	restAPI.SetHeader("Content-Type", "application/json")

	url := fmt.Sprintf("http://%s/v1/catalog/deregister", cmd.Host)
	body := fmt.Sprintf(`{"Datacenter": "%s", "Node": "%s", "ServiceID": "%s"}`, datacenter, node, serviceName)
	resp, err := restAPI.Put(url, body)
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New("failed to deregister service from consul, got status_code=" + strconv.Itoa(resp.StatusCode()))
	}

	return nil
}
