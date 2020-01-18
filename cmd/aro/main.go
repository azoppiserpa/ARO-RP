package main

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"os"
	"strings"

	utillog "github.com/Azure/ARO-RP/pkg/util/log"
)

var (
	gitCommit = "unknown"
)

func main() {
	ctx := context.Background()
	log := utillog.GetLogger()

	log.Printf("starting, git commit %s", gitCommit)

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s {rp,mirror}", os.Args[0])
	}

	var err error
	switch strings.ToLower(os.Args[1]) {
	case "mirror":
		err = mirror(ctx, log)
	case "monitor":
		err = monitor(ctx, log)
	case "rp":
		err = rp(ctx, log)
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		log.Fatal(err)
	}
}
