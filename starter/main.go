package main

import (
	"context"
	vaultauth "github.com/temporal-sa/temporal-proxy-jwt-worker"
	"go.temporal.io/sdk/client"
	"log"

	"github.com/temporalio/samples-go/helloworld"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:  "127.0.0.1:9000",        // address of Temporal Cloud proxy
		Namespace: "<namespace>.<account>", // Temporal Cloud namespace must be provided
		HeadersProvider: &vaultauth.VaultHeadersProvider{
			Config: vaultauth.VaultConfig{
				VaultAddr: "http://127.0.0.1:8200",
				Username:  "username",
				Password:  "password",
				OidcRole:  "temporal-worker",
			},
			WorkloadId: "my-workload", // maps to proxy config
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "vault-auth",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
