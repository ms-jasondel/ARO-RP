package main

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	deployer "github.com/Azure/ARO-RP/pkg/deploy"
)

func deploy(ctx context.Context, log *logrus.Entry) error {
	for _, key := range []string{
		"AZURE_SUBSCRIPTION_ID",
		"LOCATION",
		"RESOURCEGROUP",
	} {
		if _, found := os.LookupEnv(key); !found {
			return fmt.Errorf("environment variable %q unset", key)
		}
	}

	deployVersion := gitCommit
	if os.Getenv("RP_VERSION") != "" {
		deployVersion = os.Getenv("RP_VERSION")
	}

	if deployVersion == "unknown" || strings.Contains(deployVersion, "dirty") {
		return fmt.Errorf("invalid deploy version %q", deployVersion)
	}

	config, err := deployer.GetConfig(flag.Arg(1), flag.Arg(2))
	if err != nil {
		return err
	}

	deployer, err := deployer.New(ctx, log, config, deployVersion)
	if err != nil {
		return err
	}

	rpServicePrincipalID, err := deployer.PreDeploy(ctx)
	if err != nil {
		return err
	}

	// if pre-deploy is set, we terminate early. This is so we could populate
	// vault and other configurations, required for main deployment. This is usually
	// day 1 deployment setting only
	if os.Getenv("RP_PREDEPLOY_ONLY") != "" {
		return nil
	}

	err = deployer.Deploy(ctx, rpServicePrincipalID)
	if err != nil {
		return err
	}

	return deployer.Upgrade(ctx)
}
