package main

import (
	vaultauth "github.com/temporal-sa/temporal-proxy-jwt-worker"
	"github.com/temporalio/samples-go/helloworld"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
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

	w := worker.New(c, "vault-auth", worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
