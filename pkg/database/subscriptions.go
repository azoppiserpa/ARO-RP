package database

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/ARO-RP/pkg/api"
	"github.com/Azure/ARO-RP/pkg/database/cosmosdb"
)

type subscriptions struct {
	c    cosmosdb.SubscriptionDocumentClient
	uuid string
}

// Subscriptions is the database interface for SubscriptionDocuments
type Subscriptions interface {
	Create(*api.SubscriptionDocument) (*api.SubscriptionDocument, error)
	Get(string) (*api.SubscriptionDocument, error)
	Update(*api.SubscriptionDocument) (*api.SubscriptionDocument, error)
	Dequeue() (*api.SubscriptionDocument, error)
	Lease(string) (*api.SubscriptionDocument, error)
	EndLease(string, bool, bool) (*api.SubscriptionDocument, error)
}

// NewSubscriptions returns a new Subscriptions
func NewSubscriptions(uuid string, dbc cosmosdb.DatabaseClient, dbid, collid string) (Subscriptions, error) {
	collc := cosmosdb.NewCollectionClient(dbc, dbid)

	triggers := []*cosmosdb.Trigger{
		{
			ID:               "renewLease",
			TriggerOperation: cosmosdb.TriggerOperationAll,
			TriggerType:      cosmosdb.TriggerTypePre,
			Body: `function trigger() {
	var request = getContext().getRequest();
	var body = request.getBody();
	var date = new Date();
	body["leaseExpires"] = Math.floor(date.getTime() / 1000) + 60;
	request.setBody(body);
}`,
		},
		{
			ID:               "retryLater",
			TriggerOperation: cosmosdb.TriggerOperationAll,
			TriggerType:      cosmosdb.TriggerTypePre,
			Body: `function trigger() {
	var request = getContext().getRequest();
	var body = request.getBody();
	var date = new Date();
	body["leaseExpires"] = Math.floor(date.getTime() / 1000) + 600;
	request.setBody(body);
}`,
		},
	}

	triggerc := cosmosdb.NewTriggerClient(collc, collid)
	for _, trigger := range triggers {
		_, err := triggerc.Create(trigger)
		if err != nil && !cosmosdb.IsErrorStatusCode(err, http.StatusConflict) {
			return nil, err
		}
	}

	return &subscriptions{
		c:    cosmosdb.NewSubscriptionDocumentClient(collc, collid),
		uuid: uuid,
	}, nil
}

func (c *subscriptions) Create(doc *api.SubscriptionDocument) (*api.SubscriptionDocument, error) {
	if doc.ID != strings.ToLower(doc.ID) {
		return nil, fmt.Errorf("id %q is not lower case", doc.ID)
	}

	doc, err := c.c.Create(doc.ID, doc, nil)

	if err, ok := err.(*cosmosdb.Error); ok && err.StatusCode == http.StatusConflict {
		err.StatusCode = http.StatusPreconditionFailed
	}

	return doc, err
}

func (c *subscriptions) Get(id string) (*api.SubscriptionDocument, error) {
	if id != strings.ToLower(id) {
		return nil, fmt.Errorf("id %q is not lower case", id)
	}

	return c.c.Get(id, id)
}

func (c *subscriptions) patch(id string, f func(*api.SubscriptionDocument) error, options *cosmosdb.Options) (*api.SubscriptionDocument, error) {
	var doc *api.SubscriptionDocument

	err := cosmosdb.RetryOnPreconditionFailed(func() (err error) {
		doc, err = c.Get(id)
		if err != nil {
			return
		}

		err = f(doc)
		if err != nil {
			return
		}

		doc, err = c.update(doc, options)
		return
	})

	return doc, err
}

func (c *subscriptions) Update(doc *api.SubscriptionDocument) (*api.SubscriptionDocument, error) {
	return c.update(doc, nil)
}

func (c *subscriptions) update(doc *api.SubscriptionDocument, options *cosmosdb.Options) (*api.SubscriptionDocument, error) {
	if doc.ID != strings.ToLower(doc.ID) {
		return nil, fmt.Errorf("id %q is not lower case", doc.ID)
	}

	return c.c.Replace(doc.ID, doc, options)
}

func (c *subscriptions) Dequeue() (*api.SubscriptionDocument, error) {
	i := c.c.Query("", &cosmosdb.Query{
		Query: `SELECT * FROM Subscriptions doc WHERE (doc.deleting ?? false) AND (doc.leaseExpires ?? 0) < GetCurrentTimestamp() / 1000`,
	})

	for {
		docs, err := i.Next()
		if err != nil {
			return nil, err
		}
		if docs == nil {
			return nil, nil
		}

		for _, doc := range docs.SubscriptionDocuments {
			doc.LeaseOwner = c.uuid
			doc.Dequeues++
			doc, err = c.update(doc, &cosmosdb.Options{PreTriggers: []string{"renewLease"}})
			if cosmosdb.IsErrorStatusCode(err, http.StatusPreconditionFailed) { // someone else got there first
				continue
			}
			return doc, err
		}
	}
}

func (c *subscriptions) Lease(id string) (*api.SubscriptionDocument, error) {
	return c.patch(id, func(doc *api.SubscriptionDocument) error {
		if doc.LeaseOwner != c.uuid {
			return fmt.Errorf("lost lease")
		}
		return nil
	}, &cosmosdb.Options{PreTriggers: []string{"renewLease"}})
}

func (c *subscriptions) EndLease(id string, done, retryLater bool) (*api.SubscriptionDocument, error) {
	var options *cosmosdb.Options
	if retryLater {
		options = &cosmosdb.Options{PreTriggers: []string{"retryLater"}}
	}

	return c.patch(id, func(doc *api.SubscriptionDocument) error {
		if doc.LeaseOwner != c.uuid {
			return fmt.Errorf("lost lease")
		}

		if done {
			doc.Deleting = false
		}

		doc.LeaseOwner = ""
		doc.LeaseExpires = 0

		if done || retryLater {
			doc.Dequeues = 0
		}

		return nil
	}, options)
}
