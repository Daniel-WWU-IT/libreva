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

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

// FileInfoAction queries information about a remote file through Reva.
type FileInfoAction struct {
	action
}

// Stat queries the file information for the specified remote file.
func (action *FileInfoAction) Stat(path string) (*storage.ResourceInfo, error) {
	ref := &provider.Reference{
		Spec: &provider.Reference_Path{Path: path},
	}
	req := &provider.StatRequest{Ref: ref}

	if res, err := action.session.Client().Stat(action.session.Context(), req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return nil, err
		}

		return res.Info, nil
	} else {
		return nil, fmt.Errorf("unable to query information of '%v': %v", path, err)
	}
}

func (action *FileInfoAction) FileExists(path string) bool {
	// Stat the file and see if that succeeds; if so, check if the resource is indeed a file
	if info, err := action.Stat(path); err == nil {
		return info.Type == provider.ResourceType_RESOURCE_TYPE_FILE
	} else {
		return false
	}
}

func (action *FileInfoAction) DirExists(path string) bool {
	// Stat the file and see if that succeeds; if so, check if the resource is indeed a directory
	if info, err := action.Stat(path); err == nil {
		return info.Type == provider.ResourceType_RESOURCE_TYPE_CONTAINER
	} else {
		return false
	}
}

// NewFileInfoAction creates a new file info action.
func NewFileInfoAction(session *reva.Session) (*FileInfoAction, error) {
	action := &FileInfoAction{}
	if err := action.initAction(session); err != nil {
		return nil, fmt.Errorf("unable to create the FileInfoAction: %v", err)
	}
	return action, nil
}
