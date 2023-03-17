//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package tls

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName           = "tls"
	NginxUid              = 101 // Hardcoded in nginx vendor container
	NginxGid              = 101 // Hardcoded in nginx vendor container
	DefaultNginxTlsFolder = "/etc/ssl/nginx"
	DefaultNginxCertFile  = "nginx.crt"
	DefaultNginxKeyFile   = "nginx.key"
)

type cmd struct {
	loggingClient   logger.LoggingClient
	client          internal.HttpCaller
	fileOpener      fileioperformer.FileIoPerformer
	certificatePath string
	privateKeyPath  string
	targetFolder    string
	certFilename    string
	keyFilename     string
}

func NewCommand(
	lc logger.LoggingClient,
	args []string) (*cmd, error) {

	cmd := cmd{
		loggingClient: lc,
		client:        pkg.NewRequester(lc).Insecure(),
		fileOpener:    fileioperformer.NewDefaultFileIoPerformer(),
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.StringVar(&cmd.certificatePath, "inCert", "", "Path to PEM-encoded leaf certificate")
	flagSet.StringVar(&cmd.privateKeyPath, "inKey", "", "Path to PEM-encoded private key")
	flagSet.StringVar(&cmd.targetFolder, "targetFolder", DefaultNginxTlsFolder, "Path to TLS key file")
	flagSet.StringVar(&cmd.certFilename, "certFilename", DefaultNginxCertFile, "Filename of certificate file (on target)")
	flagSet.StringVar(&cmd.keyFilename, "keyFilename", DefaultNginxKeyFile, "Filename of private key file (on target")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, err
	}
	if cmd.certificatePath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --inCert is required", os.Args[0])
	}
	if cmd.privateKeyPath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --inKey is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (statusCode int, err error) {

	destCertPath := filepath.Join(c.targetFolder, c.certFilename)
	if err := c.copyFile(destCertPath, c.certificatePath, 0644, NginxUid, NginxGid); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	destKeyPath := filepath.Join(c.targetFolder, c.keyFilename)
	if err := c.copyFile(destKeyPath, c.privateKeyPath, 0600, NginxUid, NginxGid); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return interfaces.StatusCodeExitNormal, nil

}

// copyFile copies a single file with given permissions
func (c *cmd) copyFile(destPath string, srcPath string, perm os.FileMode, uid int, gid int) error {

	src, err := c.fileOpener.OpenFileReader(srcPath, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	if srcCloser, _ := src.(io.Closer); srcCloser != nil {
		defer func() { _ = srcCloser.Close() }()
	}

	dest, err := c.fileOpener.OpenFileWriter(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	destCloser, _ := dest.(io.Closer)
	if destCloser != nil {
		defer func() { _ = destCloser.Close() }()
	}

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	// Explicit close of target file
	if err = destCloser.Close(); err != nil {
		return err
	}

	return nil
}
