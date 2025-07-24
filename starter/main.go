package main

import (
	"context"
	"go.temporal.io/sdk/client"
	"log"
	"temporal-sa/vaultauth"

	"github.com/temporalio/samples-go/helloworld"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:  "127.0.0.1:9000",
		Namespace: "brendan-myers-aws.a2dd6",
		HeadersProvider: &vaultauth.VaultHeadersProvider{
			Config: vaultauth.VaultConfig{
				VaultAddr: "http://127.0.0.1:8200",
				Username:  "brendan",
				Password:  "secret",
				OidcRole:  "temporal-worker",
			},
			WorkloadId: "brendan-keys",
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
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
