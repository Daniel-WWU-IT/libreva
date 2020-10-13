// Copyright 2018-2020 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package net

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eventials/go-tus"
	"github.com/eventials/go-tus/memorystore"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
)

// TUSClient is a simple client for uploading files via TUS.
type TUSClient struct {
	config *tus.Config
	client *tus.Client
}

func (client *TUSClient) initClient(endpoint string, accessToken string, transportToken string) error {
	// Create the TUS configuration
	client.config = tus.DefaultConfig()
	client.config.Resume = true

	if memStore, err := memorystore.NewMemoryStore(); err == nil {
		client.config.Store = memStore
	} else {
		return fmt.Errorf("unable to create a TUS memory store: %v", err)
	}

	client.config.Header.Add(common.AccessTokenName, accessToken)

	if transportToken != "" {
		client.config.Header.Add(common.TransportTokenName, transportToken)
	}

	// Create the TUS client
	if tusClient, err := tus.NewClient(endpoint, client.config); err == nil {
		client.client = tusClient
	} else {
		return fmt.Errorf("error creating the TUS client: %v", err)
	}

	return nil
}

// Write writes data from a stream to the endpoint.
func (client *TUSClient) Write(data io.Reader, target string, fileInfo os.FileInfo, checksumType string, checksum string) error {
	metadata := map[string]string{
		"filename": filepath.Base(target),
		"dir":      filepath.Dir(target),
		"checksum": fmt.Sprintf("%s %s", checksumType, checksum),
	}
	fingerprint := fmt.Sprintf("%s-%d-%s-%s", filepath.Base(target), fileInfo.Size(), fileInfo.ModTime(), checksum)

	upload := tus.NewUpload(data, fileInfo.Size(), metadata, fingerprint)
	client.config.Store.Set(upload.Fingerprint, client.client.Url)
	uploader := tus.NewUploader(client.client, client.client.Url, upload, 0)

	if err := uploader.Upload(); err != nil {
		return fmt.Errorf("unable to perform the TUS upload for '%v': %v", client.client.Url, err)
	}

	return nil
}

// NewTUSClient creates a new TUS client.
func NewTUSClient(endpoint string, accessToken string, transportToken string) (*TUSClient, error) {
	client := &TUSClient{}
	if err := client.initClient(endpoint, accessToken, transportToken); err != nil {
		return nil, fmt.Errorf("unable to create the TUS client: %v", err)
	}
	return client, nil
}
