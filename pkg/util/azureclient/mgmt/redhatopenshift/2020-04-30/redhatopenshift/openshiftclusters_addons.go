package redhatopenshift

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"

	mgmtredhatopenshift20200430 "github.com/Azure/ARO-RP/pkg/client/services/redhatopenshift/mgmt/2020-04-30/redhatopenshift"
)

// OpenShiftClustersClientAddons contains addons for OpenShiftClustersClient
type OpenShiftClustersClientAddons interface {
	CreateOrUpdateAndWait(ctx context.Context, resourceGroupName string, resourceName string, parameters mgmtredhatopenshift20200430.OpenShiftCluster) error
	DeleteAndWait(ctx context.Context, resourceGroupName string, resourceName string) error
	List(ctx context.Context) (clusters []mgmtredhatopenshift20200430.OpenShiftCluster, err error)
	ListByResourceGroup(ctx context.Context, resourceGroupName string) (clusters []mgmtredhatopenshift20200430.OpenShiftCluster, err error)
}

func (c *openShiftClustersClient) CreateOrUpdateAndWait(ctx context.Context, resourceGroupName string, resourceName string, parameters mgmtredhatopenshift20200430.OpenShiftCluster) error {
	future, err := c.CreateOrUpdate(ctx, resourceGroupName, resourceName, parameters)
	if err != nil {
		return err
	}

	return future.WaitForCompletionRef(ctx, c.Client)
}

func (c *openShiftClustersClient) DeleteAndWait(ctx context.Context, resourceGroupName string, resourceName string) error {
	future, err := c.Delete(ctx, resourceGroupName, resourceName)
	if err != nil {
		return err
	}

	return future.WaitForCompletionRef(ctx, c.Client)
}

func (c *openShiftClustersClient) List(ctx context.Context) (clusters []mgmtredhatopenshift20200430.OpenShiftCluster, err error) {
	page, err := c.OpenShiftClustersClient.List(ctx)
	if err != nil {
		return nil, err
	}

	for page.NotDone() {
		clusters = append(clusters, page.Values()...)

		err = page.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
	}

	return clusters, nil
}

func (c *openShiftClustersClient) ListByResourceGroup(ctx context.Context, resourceGroupName string) (clusters []mgmtredhatopenshift20200430.OpenShiftCluster, err error) {
	page, err := c.OpenShiftClustersClient.ListByResourceGroup(ctx, resourceGroupName)
	if err != nil {
		return nil, err
	}

	for page.NotDone() {
		clusters = append(clusters, page.Values()...)

		err = page.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
	}

	return clusters, nil
}
