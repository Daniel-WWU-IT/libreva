/*
 * MIT License
 *
 * Copyright (c) 2020 Daniel Mueller
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package action

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	types "github.com/cs3org/go-cs3apis/cs3/types/v1beta1"
	"github.com/cs3org/reva/pkg/rhttp"
	"github.com/studio-b12/gowebdav"

	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

// DownloadAction is used to download files through Reva.
type DownloadAction struct {
	action
}

// DownloadFile retrieves the provided file data; in case of an error, nil is returned.
func (action *DownloadAction) DownloadFile(fileInfo *storage.ResourceInfo) ([]byte, error) {
	if fileInfo.Type != storage.ResourceType_RESOURCE_TYPE_FILE {
		return []byte{}, fmt.Errorf("resource is not a file")
	}

	// Issue a file download request to Reva; this will provide the endpoint to read the file data from
	if download, err := action.initiateDownload(fileInfo); err == nil {
		// Try to get a WebDAV reader first
		reader, supported, err := action.getWebDAVReader(download.DownloadEndpoint, download.Opaque)
		if err != nil {
			if supported { // WebDAV is supported but failed
				return nil, fmt.Errorf("downloading via WebDAV failed: %v", err)
			}

			reader, err = action.getHTTPReader(download.DownloadEndpoint, download.Token)
			if err != nil {
				return nil, fmt.Errorf("no data reader could be created to read from '%v': %v", download.DownloadEndpoint, err)
			}
		}

		if data, err := ioutil.ReadAll(reader); err == nil {
			return data, nil
		} else {
			return nil, fmt.Errorf("error while reading from '%v': %v", download.DownloadEndpoint, err)
		}
	} else {
		return nil, err
	}
}

func (action *DownloadAction) initiateDownload(fileInfo *storage.ResourceInfo) (*gateway.InitiateFileDownloadResponse, error) {
	req := &provider.InitiateFileDownloadRequest{
		Ref: &provider.Reference{
			Spec: &provider.Reference_Path{
				Path: fileInfo.Path,
			},
		},
	}

	if res, err := action.session.Client().InitiateFileDownload(action.session.Context(), req); err == nil {
		if err := reva.CheckRPCStatus(res.Status); err != nil {
			return nil, err
		}

		return res, nil
	} else {
		return nil, fmt.Errorf("unable to initiate download on '%v': %v", fileInfo.Path, err)
	}
}

func (action *DownloadAction) getWebDAVReader(endpoint string, opaque *types.Opaque) (io.Reader, bool, error) {
	if opaque == nil {
		return nil, false, fmt.Errorf("missing Opaque object")
	}

	checkOpaqueDecoder := func(o *types.OpaqueEntry) error {
		if strings.EqualFold(o.Decoder, "plain") {
			return nil
		} else {
			return fmt.Errorf("unsupported Opaque decoder '%v'", o.Decoder)
		}
	}

	if tokenOpaque, ok := opaque.Map["webdav-token"]; ok {
		if err := checkOpaqueDecoder(tokenOpaque); err != nil {
			return nil, false, err
		}

		if fileOpaque, ok := opaque.Map["webdav-file-path"]; ok {
			if err := checkOpaqueDecoder(fileOpaque); err != nil {
				return nil, false, err
			}

			webdav := gowebdav.NewClient(endpoint, "", "")
			webdav.SetHeader(reva.AccessTokenName, string(tokenOpaque.Value))

			if reader, err := webdav.ReadStream(string(fileOpaque.Value)); err == nil {
				return reader, true, nil
			} else {
				return nil, true, fmt.Errorf("unable to read from WebDAV endpoint '%v': %v", endpoint, err)
			}
		} else {
			return nil, false, fmt.Errorf("WebDAV file path missing")
		}
	} else {
		return nil, false, fmt.Errorf("WebDAV token missing")
	}
}

func (action *DownloadAction) getHTTPReader(endpoint string, transportToken string) (io.Reader, error) {
	if httpReq, err := rhttp.NewRequest(action.session.Context(), "GET", endpoint, nil); err == nil {
		httpReq.Header.Set(reva.TransportTokenName, transportToken)

		httpClient := rhttp.GetHTTPClient(
			rhttp.Context(action.session.Context()),
			rhttp.Insecure(true),
			rhttp.Timeout(time.Duration(24*int64(time.Hour))),
		)

		if httpRes, err := httpClient.Do(httpReq); err == nil {
			defer httpRes.Body.Close()

			if httpRes.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("retrieving data from '%v' failed: %v", endpoint, httpRes.Status)
			}

			return httpRes.Body, nil
		} else {
			return nil, fmt.Errorf("unable to perform the HTTP request for '%v': %v", endpoint, err)
		}
	} else {
		return nil, fmt.Errorf("unable to generate the HTTP request for '%v': %v", endpoint, err)
	}
}

// NewDownloadAction creates a new download action.
func NewDownloadAction(session *reva.Session) (*DownloadAction, error) {
	action := &DownloadAction{}
	if err := action.initAction(session); err != nil {
		return nil, fmt.Errorf("unable to create the DownloadAction: %v", err)
	}
	return action, nil
}
