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
	"time"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
)

type HTTPRequest struct {
	endpoint string
	data     io.Reader

	client  *http.Client
	request *http.Request
}

func (request *HTTPRequest) initRequest(session *Session, endpoint string, method string, transportToken string, data io.Reader) error {
	request.endpoint = endpoint
	request.data = data

	// Initialize the HTTP client
	request.client = &http.Client{
		Timeout: time.Duration(24 * int64(time.Hour)),
	}

	// Initialize the HTTP request
	if httpReq, err := http.NewRequestWithContext(session.Context(), method, endpoint, data); err == nil {
		request.request = httpReq

		// Set mandatory header values
		request.request.Header.Set(common.AccessTokenName, session.Token())
		request.request.Header.Set(common.TransportTokenName, transportToken)

		return nil
	} else {
		return err
	}
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

// AddParameters adds the specified parameters to the resulting query.
func (request *HTTPRequest) AddParameters(params map[string]string) {
	query := request.request.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	request.request.URL.RawQuery = query.Encode()
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
func (request *HTTPRequest) Write() error {
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
