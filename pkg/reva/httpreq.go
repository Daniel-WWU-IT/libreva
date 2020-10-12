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

package reva

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
	"github.com/Daniel-WWU-IT/libreva/pkg/common/crypto"
	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
)

type HTTPRequest struct {
	endpoint       string
	data           io.Reader
	accessToken    string
	transportToken string

	client  *http.Client
	request *http.Request
}

func (request *HTTPRequest) initRequest(session *Session, endpoint string, method string, transportToken string, data io.Reader) error {
	request.endpoint = endpoint
	request.data = data
	request.accessToken = session.Token()
	request.transportToken = transportToken

	// Initialize the HTTP client
	request.client = &http.Client{
		Timeout: time.Duration(24 * int64(time.Hour)),
	}

	// Initialize the HTTP request
	if httpReq, err := http.NewRequestWithContext(session.Context(), method, endpoint, data); err == nil {
		request.request = httpReq

		// Set mandatory header values
		request.request.Header.Set(common.AccessTokenName, request.accessToken)
		request.request.Header.Set(common.TransportTokenName, request.transportToken)

		return nil
	} else {
		return err
	}
}

func (request *HTTPRequest) addParameters(params map[string]string) {
	query := request.request.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	request.request.URL.RawQuery = query.Encode()
}

func (request *HTTPRequest) do() (*http.Response, error) {
	if httpRes, err := request.client.Do(request.request); err == nil {
		if httpRes.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("performing the HTTP request failed: %v", httpRes.Status)
		}
		return httpRes, nil
	} else {
		return nil, err
	}
}

// Read reads the data from the HTTP endpoint.
func (request *HTTPRequest) Read() ([]byte, error) {
	if httpRes, err := request.do(); err == nil {
		defer httpRes.Body.Close()

		if data, err := ioutil.ReadAll(httpRes.Body); err == nil {
			return data, nil
		} else {
			return nil, fmt.Errorf("reading data from '%v' failed: %v", request.endpoint, err)
		}
	} else {
		return nil, fmt.Errorf("unable to perform the HTTP request for '%v': %v", request.endpoint, err)
	}
}

// Write writes data to the HTTP endpoint.
func (request *HTTPRequest) Write(target string, fileInfo os.FileInfo, checksumType provider.ResourceChecksumType, enableTUS bool) error {
	checksum, err := request.computeDataChecksum(checksumType, request.data)
	if err != nil {
		return fmt.Errorf("unable to compute the data checksum: %v", err)
	}
	checksumTypeName := crypto.GetChecksumTypeName(checksumType)

	// Check if the data object can be seeked; if so, reset it to its beginning
	if seeker, ok := request.data.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	if enableTUS {
		return request.writeTUS(target, fileInfo, checksumTypeName, checksum)
	} else {
		return request.writePUT(checksumTypeName, checksum)
	}
}

func (request *HTTPRequest) computeDataChecksum(checksumType provider.ResourceChecksumType, data io.Reader) (string, error) {
	switch checksumType {
	case provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_ADLER32:
		return crypto.ComputeAdler32Checksum(data)
	case provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_MD5:
		return crypto.ComputeMD5Checksum(data)
	case provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_SHA1:
		return crypto.ComputeSHA1Checksum(data)
	case provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_UNSET:
		return "", nil
	default:
		return "", fmt.Errorf("invalid checksum type: %s", checksumType)
	}
}

func (request *HTTPRequest) writeTUS(target string, fileInfo os.FileInfo, checksumType string, checksum string) error {
	if tusClient, err := net.NewTUSClient(request.endpoint, request.accessToken, request.transportToken); err == nil {
		if err := tusClient.Write(request.data, target, fileInfo, checksumType, checksum); err != nil {
			return fmt.Errorf("writing data to '%v' via TUS failed: %v", request.endpoint, err)
		}

		return nil
	} else {
		return fmt.Errorf("unable to create the TUS client: %v", err)
	}
}

func (request *HTTPRequest) writePUT(checksumType string, checksum string) error {
	request.addParameters(map[string]string{
		"xs":      checksum,
		"xs_type": checksumType,
	})

	if httpRes, err := request.do(); err == nil {
		defer httpRes.Body.Close()
		return nil
	} else {
		return fmt.Errorf("unable to perform the HTTP request for '%v': %v", request.endpoint, err)
	}
}

func newHTTPRequest(session *Session, endpoint string, method string, transportToken string, data io.Reader) (*HTTPRequest, error) {
	request := &HTTPRequest{}
	if err := request.initRequest(session, endpoint, method, transportToken, data); err != nil {
		return nil, fmt.Errorf("unable to initialize the HTTP request: %v", err)
	}
	return request, nil
}
