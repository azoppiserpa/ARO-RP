package main

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"github.com/Azure/ARO-RP/pkg/backend"
	"github.com/Azure/ARO-RP/pkg/database"
	"github.com/Azure/ARO-RP/pkg/env"
	"github.com/Azure/ARO-RP/pkg/frontend"
)

func rp(ctx context.Context, log *logrus.Entry) error {
	uuid := uuid.NewV4().String()
	log.Printf("uuid %s", uuid)

	env, err := env.NewEnv(ctx, log)
	if err != nil {
		return err
	}

	db, err := database.NewDatabase(env, uuid)
	if err != nil {
		return err
	}

	sigterm := make(chan os.Signal, 1)
	stop := make(chan struct{})
	done := make(chan struct{})
	signal.Notify(sigterm, syscall.SIGTERM)

	b, err := backend.NewBackend(ctx, log.WithField("component", "backend"), env, db)
	if err != nil {
		return err
	}

	f, err := frontend.NewFrontend(ctx, log.WithField("component", "frontend"), env, db)
	if err != nil {
		return err
	}

	log.Print("listening")

	go b.Run(stop)
	go f.Run(stop, done)

	<-sigterm
	log.Print("received SIGTERM")
	close(stop)
	<-done

	return nil
}
