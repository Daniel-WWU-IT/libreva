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

package action

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	types "github.com/cs3org/go-cs3apis/cs3/types/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common/crypto"
	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

// UploadAction is used to upload files through Reva.
type UploadAction struct {
	action
}

// UploadFile uploads the provided file data to the target file; in case of an error, nil is returned.
func (action *UploadAction) UploadFile(target string, data io.Reader, fileInfo os.FileInfo, enableTUS bool) (*storage.ResourceInfo, error) {
	if target == "" {
		return nil, fmt.Errorf("no target specified")
	}

	// TODO: Check if target exists (stat), make target dir (mkdir), etc.

	// Issue a file upload request to Reva; this will provide the endpoint to write the file data to
	if upload, err := action.initiateUpload(target, fileInfo.Size()); err == nil {
		// Try to upload the file via WebDAV first
		if client, err := net.NewWebDAVClient(upload.UploadEndpoint, upload.Opaque); err == nil {
			if err := client.Write(data, fileInfo.Size()); err != nil {
				return nil, fmt.Errorf("error while writing to '%v' via WebDAV: %v", upload.UploadEndpoint, err)
			}
		} else {
			// WebDAV is not supported, so directly write to the HTTP endpoint
			checksumType := action.selectChecksumType(upload.AvailableChecksums)
			checksumTypeName := crypto.GetChecksumTypeName(checksumType)
			checksum, err := crypto.ComputeChecksum(checksumType, data)
			if err != nil {
				return nil, fmt.Errorf("unable to compute data checksum: %v", err)
			}

			// Check if the data object can be seeked; if so, reset it to its beginning
			if seeker, ok := data.(io.Seeker); ok {
				seeker.Seek(0, 0)
			}

			if enableTUS {
				if err := action.uploadFileTUS(upload, target, data, fileInfo, checksum, checksumTypeName); err != nil {
					return nil, fmt.Errorf("error while writing to '%v' via TUS: %v", upload.UploadEndpoint, err)
				}
			} else {
				if err := action.uploadFilePUT(upload, data, checksum, checksumTypeName); err != nil {
					return nil, fmt.Errorf("error while writing to '%v' via HTTP: %v", upload.UploadEndpoint, err)
				}
			}
		}
	} else {
		return nil, err
	}

	// Query information about the just-uploaded file
	if fileInfoAct, err := NewFileInfoAction(action.session); err == nil {
		if info, err := fileInfoAct.Stat(target); err == nil {
			return info, nil
		} else {
			return nil, fmt.Errorf("the uploaded data was not written to the target file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("unable to create file info action: %v", err)
	}
}

func (action *UploadAction) initiateUpload(target string, size int64) (*gateway.InitiateFileUploadResponse, error) {
	// Initiating an upload request gets us the upload endpoint for the specified target
	req := &provider.InitiateFileUploadRequest{
		Ref: &provider.Reference{
			Spec: &provider.Reference_Path{
				Path: target,
			},
		},
		Opaque: &types.Opaque{
			Map: map[string]*types.OpaqueEntry{
				"Upload-Length": {
					Decoder: "plain",
					Value:   []byte(strconv.FormatInt(size, 10)),
				},
			},
		},
	}

	if res, err := action.session.Client().InitiateFileUpload(action.session.Context(), req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return nil, err
		}

		return res, nil
	} else {
		return nil, fmt.Errorf("unable to initiate upload to '%v': %v", target, err)
	}
}

func (action *UploadAction) selectChecksumType(checksumTypes []*provider.ResourceChecksumPriority) provider.ResourceChecksumType {
	var selChecksumType provider.ResourceChecksumType
	var maxPrio uint32 = math.MaxUint32
	for _, xs := range checksumTypes {
		if xs.Priority < maxPrio {
			maxPrio = xs.Priority
			selChecksumType = xs.Type
		}
	}
	return selChecksumType
}

func (action *UploadAction) uploadFilePUT(upload *gateway.InitiateFileUploadResponse, data io.Reader, checksum string, checksumType string) error {
	if request, err := action.session.NewWriteRequest(upload.UploadEndpoint, upload.Token, data); err == nil {
		request.AddParameters(map[string]string{
			"xs":      checksum,
			"xs_type": checksumType,
		})

		return request.Write()
	} else {
		return fmt.Errorf("unable to create HTTP request for '%v': %v", upload.UploadEndpoint, err)
	}
}

func (action *UploadAction) uploadFileTUS(upload *gateway.InitiateFileUploadResponse, target string, data io.Reader, fileInfo os.FileInfo, checksum string, checksumType string) error {
	if tusClient, err := net.NewTUSClient(upload.UploadEndpoint, action.session.Token(), upload.Token); err == nil {
		return tusClient.Write(data, target, fileInfo, checksumType, checksum)
	} else {
		return fmt.Errorf("unable to create TUS client: %v", err)
	}
}

// NewUploadAction creates a new upload action.
func NewUploadAction(session *reva.Session) (*UploadAction, error) {
	action := &UploadAction{}
	if err := action.initAction(session); err != nil {
		return nil, fmt.Errorf("unable to create the UploadAction: %v", err)
	}
	return action, nil
}
