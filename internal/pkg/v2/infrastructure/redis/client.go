//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	redisClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/google/uuid"
)

var currClient *Client // a singleton so Readings can be de-referenced
var once sync.Once

type Client struct {
	*redisClient.Client
}

func NewClient(config db.Configuration, logger logger.LoggingClient) (*Client, error) {
	var err error
	dc := &Client{}
	dc.Client, err = redisClient.NewClient(config, logger)
	if err != nil {
		return nil, err
	}

	return dc, err
}

// CloseSession closes the connections to Redis
func (c *Client) CloseSession() {
	c.Pool.Close()

	currClient = nil
	once = sync.Once{}
}

// Add a new event
func (c *Client) AddEvent(e model.Event) (addedEvent model.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if e.Id != "" {
		_, err = uuid.Parse(e.Id)
		if err != nil {
			return model.Event{}, db.ErrInvalidObjectId
		}
	}

	return addEvent(conn, e)
}
