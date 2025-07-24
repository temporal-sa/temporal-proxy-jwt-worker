package main

import (
	"github.com/temporalio/samples-go/helloworld"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
	"temporal-sa/vaultauth"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
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

	w := worker.New(c, "vault-auth", worker.Options{})

	w.RegisterWorkflow(helloworld.Workflow)
	w.RegisterActivity(helloworld.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
