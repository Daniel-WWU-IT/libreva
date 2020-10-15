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

	types "github.com/cs3org/go-cs3apis/cs3/types/v1beta1"
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

	filePath string
}

func (webdav *WebDAVClient) initClient(endpoint string, filePath string, userName string, password string, accessToken string) error {
	if filePath == "" {
		return fmt.Errorf("no file path specified")
	}
	webdav.filePath = filePath

	// Create the WebDAV client
	webdav.client = gowebdav.NewClient(endpoint, userName, password)

	if accessToken != "" {
		webdav.client.SetHeader(common.AccessTokenName, accessToken)
	}

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

func newWebDAVClient(endpoint string, filePath string, userName string, password string, accessToken string) (*WebDAVClient, error) {
	client := &WebDAVClient{}
	if err := client.initClient(endpoint, filePath, userName, password, accessToken); err != nil {
		return nil, fmt.Errorf("unable to create the WebDAV client: %v", err)
	}
	return client, nil
}

// NewWebDAVClient creates a new WebDAV client using an access token.
func NewWebDAVClient(endpoint string, filePath string, accessToken string) (*WebDAVClient, error) {
	return newWebDAVClient(endpoint, filePath, "", "", accessToken)
}

// NewWebDAVClientWithOpaque creates a new WebDAV client using the information stored in the opaque.
func NewWebDAVClientWithOpaque(endpoint string, opaque *types.Opaque) (*WebDAVClient, error) {
	if values, err := common.GetValuesFromOpaque(opaque, []string{WebDAVTokenName, WebDAVPathName}, true); err == nil {
		return NewWebDAVClient(endpoint, values[WebDAVPathName], values[WebDAVTokenName])
	} else {
		return nil, fmt.Errorf("invalid opaque object: %v", err)
	}
}

// NewWebDAVClientWithCredentials creates a new WebDAV client with user credentials.
func NewWebDAVClientWithCredentials(endpoint string, filePath string, userName string, password string) (*WebDAVClient, error) {
	return newWebDAVClient(endpoint, filePath, userName, password, "")
}
