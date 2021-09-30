package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mongodb-forks/digest"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/atlas/mongodbatlas"
)

type Options struct {
	ProjectName string `long:"projectName" description:"Atlas project name" required:"true"`
	List        bool   `long:"list" description:"list alerts"`
	Add         string `long:"projectName" description:"Atlas project name" required:"true"`
}

func print(list []mongodbatlas.ProjectIPAccessList) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CIDR Block", "Comment"})

	for _, item := range list {
		var data = []string{item.CIDRBlock, item.Comment}
		table.Append(data)
	}
	table.Render()
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

	lists, _, err := client.ProjectIPAccessList.List(context.Background(), project.ID, nil)
	if err != nil {
		log.Fatalf("AlertConfigurations.List returned error: %v", err)
	}

	if opts.List {
		print(lists.Results)
	}

}
