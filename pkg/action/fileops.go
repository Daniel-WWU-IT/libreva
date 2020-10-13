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
	p "path"
	"strings"

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	storage "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/pkg/common/net"
	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

// FileOperationsAction offers basic file operations.
type FileOperationsAction struct {
	action
}

// Stat queries the file information for the specified remote file.
func (action *FileOperationsAction) Stat(path string) (*storage.ResourceInfo, error) {
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
		return nil, fmt.Errorf("unable to query information for '%v': %v", path, err)
	}
}

// FileExists checks whether the specified exists.
func (action *FileOperationsAction) FileExists(path string) bool {
	// Stat the file and see if that succeeds; if so, check if the resource is indeed a file
	if info, err := action.Stat(path); err == nil {
		return info.Type == provider.ResourceType_RESOURCE_TYPE_FILE
	} else {
		return false
	}
}

// DirExists checks whether the specified directory exists.
func (action *FileOperationsAction) DirExists(path string) bool {
	// Stat the file and see if that succeeds; if so, check if the resource is indeed a directory
	if info, err := action.Stat(path); err == nil {
		return info.Type == provider.ResourceType_RESOURCE_TYPE_CONTAINER
	} else {
		return false
	}
}

// ResourceExists checks whether the specified resource exists (w/o checking for its actual type).
func (action *FileOperationsAction) ResourceExists(path string) bool {
	// Stat the file and see if that succeeds
	_, err := action.Stat(path)
	return err == nil
}

// MakePath creates all directories specified by the path.
func (action *FileOperationsAction) MakePath(path string) error {
	path = strings.TrimPrefix(path, "/")

	var curPath string
	for _, token := range strings.Split(path, "/") {
		curPath = p.Join(curPath, "/"+token)

		fileInfo, err := action.Stat(curPath)
		if err != nil { // Stating failed, so the path probably doesn't exist yet
			ref := &provider.Reference{
				Spec: &provider.Reference_Path{Path: curPath},
			}
			req := &provider.CreateContainerRequest{Ref: ref}

			if res, err := action.session.Client().CreateContainer(action.session.Context(), req); err == nil {
				if err := net.CheckRPCStatus(res.Status); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unable to create container '%v': %v", curPath, err)
			}
		} else { // The path exists, so make sure that is actually a directory
			if fileInfo.Type != provider.ResourceType_RESOURCE_TYPE_CONTAINER {
				return fmt.Errorf("'%v' is not a directory", curPath)
			}
		}
	}

	return nil
}

// Move moves the specified path to a new location. If moving a file, the caller must ensure that the target directory exists.
func (action *FileOperationsAction) Move(source string, target string) error {
	if !action.ResourceExists(source) {
		return fmt.Errorf("the source '%v' doesn't exist", source)
	}
	if action.ResourceExists(target) {
		return fmt.Errorf("the target '%v' already exists", source)
	}

	sourceRef := &provider.Reference{
		Spec: &provider.Reference_Path{Path: source},
	}
	targetRef := &provider.Reference{
		Spec: &provider.Reference_Path{Path: target},
	}
	req := &provider.MoveRequest{Source: sourceRef, Destination: targetRef}

	if res, err := action.session.Client().Move(action.session.Context(), req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("unable to move '%v' to '%v': %v", source, target, err)
	}
}

// MoveTo moves the specified source to the target directory, creating it if necessary.
func (action *FileOperationsAction) MoveTo(source string, path string) error {
	if err := action.MakePath(path); err != nil {
		return fmt.Errorf("unable to create the target directory '%v': %v", path, err)
	}

	path = p.Join(path, p.Base(source)) // Keep the original resource base name
	return action.Move(source, path)
}

// Remove removes the specified path.
func (action *FileOperationsAction) Remove(path string) error {
	ref := &provider.Reference{
		Spec: &provider.Reference_Path{Path: path},
	}
	req := &provider.DeleteRequest{Ref: ref}

	if res, err := action.session.Client().Delete(action.session.Context(), req); err == nil {
		if err := net.CheckRPCStatus(res.Status); err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("unable to delete '%v': %v", path, err)
	}
}

// NewFileOperationsAction creates a new file operations action.
func NewFileOperationsAction(session *reva.Session) (*FileOperationsAction, error) {
	action := &FileOperationsAction{}
	if err := action.initAction(session); err != nil {
		return nil, fmt.Errorf("unable to create the FileOperationsAction: %v", err)
	}
	return action, nil
}
