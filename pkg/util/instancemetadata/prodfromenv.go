package instancemetadata

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest/azure"
)

type prodfromenv struct {
	instanceMetadata

	Getenv    func(key string) string
	LookupEnv func(key string) (string, bool)
}

func newProdFromEnv(ctx context.Context) (InstanceMetadata, error) {
	p := &prodfromenv{
		Getenv:    os.Getenv,
		LookupEnv: os.LookupEnv,
	}

	err := p.populateInstanceMetadata()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *prodfromenv) populateInstanceMetadata() error {

	for _, key := range []string{
		"AZURE_ENVIRONMENT",
		"AZURE_SUBSCRIPTION_ID",
		"AZURE_TENANT_ID",
		"LOCATION",
		"RESOURCEGROUP",
	} {
		if _, found := p.LookupEnv(key); !found {
			return fmt.Errorf("environment variable %q unset", key)
		}
	}

	// optional env variables
	// * HOSTNAME_OVERRIDE: defaults to os.Hostname()

	envStr := p.Getenv("AZURE_ENVIRONMENT")
	environment, err := azure.EnvironmentFromName(envStr)
	if err != nil {
		return err
	}
	p.environment = &environment

	p.subscriptionID = p.Getenv("AZURE_SUBSCRIPTION_ID")
	p.tenantID = p.Getenv("AZURE_TENANT_ID")
	p.location = p.Getenv("LOCATION")
	p.resourceGroup = p.Getenv("RESOURCEGROUP")
	p.hostname = p.Getenv("HOSTNAME_OVERRIDE") // empty string returned if not set

	if p.hostname == "" {
		hostname, err := os.Hostname()
		if err == nil {
			p.hostname = hostname
		}
	}

	return nil
}
