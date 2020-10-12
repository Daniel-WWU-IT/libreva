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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common"
	"github.com/Daniel-WWU-IT/libreva/pkg/common/crypto"
)

func ReadHTTPEndpoint(ctx context.Context, endpoint string, transportToken string) ([]byte, error) {
	if httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil); err == nil {
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

func WriteHTTPEndpoint(ctx context.Context, endpoint string, data io.Reader, size int64, checksumType provider.ResourceChecksumType, transportToken string, enableTUS bool) error {
	checksum, err := computeDataChecksum(checksumType, data)
	if err != nil {
		return fmt.Errorf("unable to compute the data checksum: %v", err)
	}

	if enableTUS {
		return writeHTTPEndpointTUS(ctx, endpoint, data, size, checksumType, checksum, transportToken)
	} else {
		return writeHTTPEndpointPUT(ctx, endpoint, data, size, checksumType, checksum, transportToken)
	}
}

func computeDataChecksum(checksumType provider.ResourceChecksumType, data io.Reader) (string, error) {
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

func writeHTTPEndpointTUS(ctx context.Context, endpoint string, data io.Reader, size int64, checksumType provider.ResourceChecksumType, checksum string, transportToken string) error {
	return nil
}

func writeHTTPEndpointPUT(ctx context.Context, endpoint string, data io.Reader, size int64, checksumType provider.ResourceChecksumType, checksum string, transportToken string) error {
	if httpReq, err := http.NewRequestWithContext(ctx, "PUT", endpoint, data); err == nil {
		httpReq.Header.Set(common.TransportTokenName, transportToken)

		query := httpReq.URL.Query()
		query.Add("xs", checksum)
		query.Add("xs_type", storageprovider.GRPC2PKGXS(xsType).String())
		httpReq.URL.RawQuery = query.Encode()

		/*
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
		*/

	} else {
		return fmt.Errorf("unable to generate the HTTP request for '%v': %v", endpoint, err)
	}

	/*
		httpReq, err := rhttp.NewRequest(ctx, "PUT", dataServerURL, reader)
					if err != nil {
						bar.Finish()
						return err
					}

					httpReq.Header.Set(datagateway.TokenTransportHeader, res.Token)
					q := httpReq.URL.Query()
					q.Add("xs", xs)
					q.Add("xs_type", storageprovider.GRPC2PKGXS(xsType).String())
					httpReq.URL.RawQuery = q.Encode()

					httpClient := rhttp.GetHTTPClient(
						rhttp.Context(ctx),
						// TODO make insecure configurable
						rhttp.Insecure(true),
						// TODO make timeout configurable
						rhttp.Timeout(time.Duration(24*int64(time.Hour))),
					)

					httpRes, err := httpClient.Do(httpReq)
					if err != nil {
						bar.Finish()
						return err
					}
					defer httpRes.Body.Close()
					if httpRes.StatusCode != http.StatusOK {
						bar.Finish()
						return err
					}
	*/
	return nil
}
