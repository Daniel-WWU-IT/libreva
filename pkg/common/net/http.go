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
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
)

// ReadHTTPEndpoint reads data from a Reva HTTP endpoint.
func ReadHTTPEndpoint(endpoint string, sessionToken string, transportToken string) ([]byte, error) {
	if httpReq, err := http.NewRequest("GET", endpoint, nil); err == nil {
		httpReq.Header.Set(common.AccessTokenName, sessionToken)
		httpReq.Header.Set(common.TransportTokenName, transportToken)

		httpClient := http.Client{
			Timeout: time.Duration(24 * int64(time.Hour)),
		}

		if httpRes, err := httpClient.Do(httpReq); err == nil {
			if httpRes.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("retrieving data from '%v' failed: %v", endpoint, httpRes.Status)
			}
			defer httpRes.Body.Close()

			if data, err := ioutil.ReadAll(httpRes.Body); err == nil {
				return data, nil
			} else {
				return nil, fmt.Errorf("reading data from '%v' failed: %v", endpoint, err)
			}
		} else {
			return nil, fmt.Errorf("unable to perform the HTTP request for '%v': %v", endpoint, err)
		}
	} else {
		return nil, fmt.Errorf("unable to generate the HTTP request for '%v': %v", endpoint, err)
	}
}
