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

	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
)

// EnumFilesAction offers functions to enumerate files and directories of a container.
type EnumFilesAction struct {
	action
}

// ListAll retrieves all files and directories contained in the provided path.
func (action *EnumFilesAction) ListAll(path string, includeSubdirectories bool) ([]*storage.ResourceInfo, error) {
	ref := &storage.Reference{
		Spec: &storage.Reference_Path{Path: path},
	}
	req := &storage.ListContainerRequest{Ref: ref}

	if res, err := action.session.Client().ListContainer(action.session.Context(), req); err == nil {
		if err := reva.CheckRPCStatus(res.Status); err != nil {
			return []*storage.ResourceInfo{}, err
		}

		fileList := make([]*storage.ResourceInfo, 0, len(res.Infos))
		for _, fi := range res.Infos {
			// Ignore resources that are neither files nor directories
			if fi.Type <= storage.ResourceType_RESOURCE_TYPE_INVALID || fi.Type >= storage.ResourceType_RESOURCE_TYPE_INTERNAL {
				continue
			}

			fileList = append(fileList, fi)

			if includeSubdirectories {
				// Recurse into subdirectories
				if fi.Type == storage.ResourceType_RESOURCE_TYPE_CONTAINER {
					if subFileList, err := action.ListAll(fi.Path, includeSubdirectories); err == nil {
						for _, fiSub := range subFileList {
							fileList = append(fileList, fiSub)
						}
					} else {
						return []*storage.ResourceInfo{}, err
					}
				}
			}
		}

		return fileList, nil
	} else {
		return []*storage.ResourceInfo{}, fmt.Errorf("unable to list files in '%v': %v", path, err)
	}
}

// NewEnumFilesAction creates a new enum files action.
func NewEnumFilesAction(session *reva.Session) (*EnumFilesAction, error) {
	action := &EnumFilesAction{}
	if err := action.initAction(session); err != nil {
		return nil, fmt.Errorf("unable to create the EnumFilesAction: %v", err)
	}
	return action, nil
}
