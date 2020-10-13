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

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

// DownloadAction is used to download files through Reva.
type DownloadAction struct {
	action
}

// DownloadFileByPath retrieves data of the provided file path; in case of an error, nil is returned.
func (action *DownloadAction) DownloadFileByPath(path string) ([]byte, error) {
	// Get the ResourceInfo object of the specified path
	if fileInfoAct, err := NewFileInfoAction(action.session); err == nil {
		if info, err := fileInfoAct.Stat(path); err == nil {
			return action.DownloadFile(info)
		} else {
			return nil, fmt.Errorf("the path '%v' was not found: %v", path, err)
		}
	} else {
		return nil, fmt.Errorf("unable to create file info action: %v", err)
	}
}

// DownloadFile retrieves data of the provided file; in case of an error, nil is returned.
func (action *DownloadAction) DownloadFile(fileInfo *storage.ResourceInfo) ([]byte, error) {
	if fileInfo.Type != storage.ResourceType_RESOURCE_TYPE_FILE {
		return nil, fmt.Errorf("resource is not a file")
	}

	// Issue a file download request to Reva; this will provide the endpoint to read the file data from
	if download, err := action.initiateDownload(fileInfo); err == nil {
		// Try to get the file via WebDAV first
		if client, err := net.NewWebDAVClient(download.DownloadEndpoint, download.Opaque); err == nil {
			if data, err := client.Read(); err == nil {
				return data, nil
			} else {
				return nil, fmt.Errorf("error while reading from '%v' via WebDAV: %v", download.DownloadEndpoint, err)
			}
		} else {
			// WebDAV is not supported, so directly read the HTTP endpoint
			if request, err := action.session.NewReadRequest(download.DownloadEndpoint, download.Token); err == nil {
				if data, err := request.Read(); err == nil {
					return data, nil
				} else {
					return nil, fmt.Errorf("error while reading from '%v' via HTTP: %v", download.DownloadEndpoint, err)
				}
			} else {
				return nil, fmt.Errorf("unable to create an HTTP request for '%v': %v", download.DownloadEndpoint, err)
			}
		}
	} else {
		return nil, err
	}
}

func (action *DownloadAction) initiateDownload(fileInfo *storage.ResourceInfo) (*gateway.InitiateFileDownloadResponse, error) {
	// Initiating a download request gets us the download endpoint for the specified resource
	req := &provider.InitiateFileDownloadRequest{
		Ref: &provider.Reference{
			Spec: &provider.Reference_Path{
				Path: fileInfo.Path,
			},
		},
	}

	if res, err := action.session.Client().InitiateFileDownload(action.session.Context(), req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return nil, err
		}

		return res, nil
	} else {
		return nil, fmt.Errorf("unable to initiate download on '%v': %v", fileInfo.Path, err)
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
