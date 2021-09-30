package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mongodb-forks/digest"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/atlas/mongodbatlas"
)

type Options struct {
	ProjectName string `long:"projectName" description:"Atlas project name" required:"true"`
	List        bool   `long:"list" description:"list alerts"`
	Import      bool   `long:"import" description:"export alerts"`
	Export      bool   `long:"export" description:"export alerts"`
	DeleteAll   bool   `long:"deleteAll" description:"delete all alerts in project"`
}

func printAlerts(alertConfigs []mongodbatlas.AlertConfiguration) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Event Type", "Metric Name"})

	for _, alertConfig := range alertConfigs {
		//fmt.Println("ProjectId: ", alertConfig)

		metricName := ""
		if alertConfig.MetricThreshold != nil {
			metricName = alertConfig.MetricThreshold.MetricName
		}
		var data = []string{alertConfig.ID, alertConfig.EventTypeName, metricName}
		table.Append(data)
	}
	table.Render()
}

func exportAlerts(alertConfigs []mongodbatlas.AlertConfiguration, fileName string) {
	file, _ := json.MarshalIndent(alertConfigs, "", " ")
	_ = ioutil.WriteFile(fileName, file, 0644)
}

func importAlerts(fileName string, client *mongodbatlas.Client) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var alertConfigs []mongodbatlas.AlertConfiguration
	json.Unmarshal(byteValue, &alertConfigs)

	for _, alertConfig := range alertConfigs {
		_, _, err := client.AlertConfigurations.Create(context.Background(), alertConfig.GroupID, &alertConfig)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

}

func deleteAlerts(alertConfigs []mongodbatlas.AlertConfiguration, client *mongodbatlas.Client) {
	for _, alertConfig := range alertConfigs {
		response, err := client.AlertConfigurations.Delete(context.Background(), alertConfig.GroupID, alertConfig.ID)
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Println("delete: ", response)
	}
}

func main() {

	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
	t := digest.NewTransport(os.Getenv("ATLAS_PUBLIC_KEY"), os.Getenv("ATLAS_PRIVATE_KEY"))
	tc, err := t.Client()
	if err != nil {
		log.Fatalf(err.Error())
	}

	client := mongodbatlas.NewClient(tc)

	project, _, err := client.Projects.GetOneProjectByName(context.Background(), opts.ProjectName)

	if err != nil {
		log.Fatalf("Projects.GetOneProjectByName returned error: %v", err)
	}

	fmt.Println("ProjectId: ", project.ID)

	alertConfigs, _, err := client.AlertConfigurations.List(context.Background(), project.ID, nil)
	if err != nil {
		log.Fatalf("AlertConfigurations.List returned error: %v", err)
	}

	if opts.List {
		printAlerts(alertConfigs)
	}

	if opts.DeleteAll {
		deleteAlerts(alertConfigs, client)
	}

	if opts.Import {
		fileName := strings.ReplaceAll(opts.ProjectName, " ", "") + "_alerts.json"
		importAlerts(fileName, client)
	}

	if opts.Export {
		fileName := strings.ReplaceAll(opts.ProjectName, " ", "") + "_alerts.json"
		exportAlerts(alertConfigs, fileName)
	}

}
