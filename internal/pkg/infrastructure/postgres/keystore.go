//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/google/uuid"
)

// AddKey adds a new key to the database
func (c *Client) AddKey(name, content string) errors.EdgeX {
	exists, err := c.KeyExists(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	} else if exists {
		return errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("key '%s' already exists", name), nil)
	}

	_, pgxErr := c.ConnPool.Exec(
		context.Background(), sqlInsert(keyStoreTableName, idCol, nameCol, contentCol),
		uuid.New().String(), name, content)
	if pgxErr != nil {
		return pgClient.WrapDBError("failed to insert row to key_store table", pgxErr)
	}
	return nil
}

// UpdateKey updates the key by name
func (c *Client) UpdateKey(name, content string) errors.EdgeX {
	_, pgxErr := c.ConnPool.Exec(
		context.Background(), sqlUpdateColsByCondCol(keyStoreTableName, nameCol, contentCol, modifiedCol), content, time.Now().UTC(), name)
	if pgxErr != nil {
		return pgClient.WrapDBError("failed to update row to key_store table", pgxErr)
	}
	return nil
}

// ReadKeyContent reads key content from the database
func (c *Client) ReadKeyContent(name string) (string, errors.EdgeX) {
	var fileContent string
	row := c.ConnPool.QueryRow(context.Background(),
		fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", contentCol, keyStoreTableName, nameCol), name)
	if err := row.Scan(&fileContent); err != nil {
		return fileContent, pgClient.WrapDBError("failed to query key content", err)
	}
	return fileContent, nil
}

// KeyExists check whether the key file exits
func (c *Client) KeyExists(filename string) (bool, errors.EdgeX) {
	var exists bool
	err := c.ConnPool.QueryRow(context.Background(), sqlCheckExistsByCol(keyStoreTableName, nameCol), filename).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to check key by name '%s' from %s table", nameCol, keyStoreTableName), err)
	}
	return exists, nil
}
