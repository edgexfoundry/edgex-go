/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package ping

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	_ "github.com/lib/pq"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	CommandName string = "pingPgDb"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct

	// options
	host     string
	port     int
	username string
	dbname   string
	pwd      string
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	_ context.Context,
	_ *sync.WaitGroup,
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		client:        pkg.NewRequester(lc).Insecure(),
		configuration: configuration,
	}

	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.StringVar(&cmd.host, "host", cmd.configuration.StageGate.KongDB.Host, "the hostname of postgres database; "+
		cmd.configuration.StageGate.KongDB.Host+" will be use if omitted")

	flagSet.IntVar(&cmd.port, "port", cmd.configuration.StageGate.KongDB.Port, "the port number of postgres database; "+
		strconv.Itoa(configuration.StageGate.KongDB.Port)+" will be use if omitted")

	flagSet.StringVar(&cmd.username, "username", "postgres", "the username of postgres database; "+
		"postgres will be use if omitted")

	flagSet.StringVar(&cmd.dbname, "dbname", "", "the database instance name of postgres database; "+
		"this is required for pinging the readiness of the database")

	flagSet.StringVar(&cmd.pwd, "password", "", "the user's password of postgres database; "+
		"this is required for pinging the readiness of the database")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd, nil
}

// Execute implements Command and runs this command
// command pingPgDb pings the Postgres database with configured db info
func (c *cmd) Execute() (statusCode int, err error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	// test readiness of kong db
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.host, c.port, c.username, c.pwd, c.dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to open sql driver connector: %v", err)
	}

	defer func() {
		_ = db.Close()
	}()

	c.loggingClient.Debug("postgres sql driver connector opened")

	err = db.Ping()
	if err != nil {
		err = fmt.Errorf("failed to ping postgres database with provided db info: %v", err)
		return interfaces.StatusCodeExitWithError, err
	}

	c.loggingClient.Info("Postgres db is ready")

	// send to stdout so that the pinger can pick up the "ready" message
	fmt.Fprintln(os.Stdout, "ready")

	return interfaces.StatusCodeExitNormal, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}
