// Copyright 2018-2020 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this filePath except in compliance with the License.
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
	"io/ioutil"
	"strconv"

	"github.com/studio-b12/gowebdav"

	"github.com/Daniel-WWU-IT/libreva/internal/common"
)

const (
	WebDAVTokenName = "webdav-token"
	WebDAVPathName  = "webdav-file-path"
)

// WebDAVClient is a simple client for down- and uploading files via WebDAV.
type WebDAVClient struct {
	client *gowebdav.Client

	token    string
	filePath string
}

func (webdav *WebDAVClient) initClient(endpoint string, values map[string]string) error {
	// Extract all necessary information from the Opaque object
	if token, ok := values[WebDAVTokenName]; ok {
		webdav.token = token
	} else {
		return fmt.Errorf("missing WebDAV token")
	}

	if file, ok := values[WebDAVPathName]; ok {
		webdav.filePath = file
	} else {
		return fmt.Errorf("missing WebDAV file path")
	}

	// Create the WebDAV client
	webdav.client = gowebdav.NewClient(endpoint, "", "")
	webdav.client.SetHeader(common.AccessTokenName, webdav.token)

	return nil
}

// Read reads all data from the endpoint.
func (webdav *WebDAVClient) Read() ([]byte, error) {
	if reader, err := webdav.client.ReadStream(webdav.filePath); err == nil {
		defer reader.Close()

		if data, err := ioutil.ReadAll(reader); err == nil {
			return data, nil
		} else {
			return nil, fmt.Errorf("unable to read the data: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unable to create reader: %v", err)
	}
}

// Write writes data from a stream to the endpoint.
func (webdav *WebDAVClient) Write(data io.Reader, size int64) error {
	if size > 0 {
		webdav.client.SetHeader("Upload-Length", strconv.FormatInt(size, 10))

		if err := webdav.client.WriteStream(webdav.filePath, data, 0700); err != nil {
			return fmt.Errorf("unable to write the data: %v", err)
		}
	}

	return nil
}

// IsSupported checks whether the endpoint supports WebDAV.
func (webdav *WebDAVClient) IsSupported() bool {
	return webdav.client != nil
}

// NewWebDAVClient creates a new WebDAV client.
func NewWebDAVClient(endpoint string, values map[string]string) (*WebDAVClient, error) {
	client := &WebDAVClient{}
	if err := client.initClient(endpoint, values); err != nil {
		return nil, fmt.Errorf("unable to create the WebDAV client: %v", err)
	}
	return client, nil
}
