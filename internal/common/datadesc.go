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

package common

import (
	"os"
	"time"
)

// DataDescriptor implements os.FileInfo to provide file information for non-file data objects.
type DataDescriptor struct {
	name string
	size int64
}

func (ddesc *DataDescriptor) Name() string {
	return ddesc.name
}

func (ddesc *DataDescriptor) Size() int64 {
	return ddesc.size
}

func (ddesc *DataDescriptor) Mode() os.FileMode {
	return os.ModePerm
}

func (ddesc *DataDescriptor) ModTime() time.Time {
	return time.Now()
}

func (ddesc *DataDescriptor) IsDir() bool {
	return false
}

func (ddesc *DataDescriptor) Sys() interface{} {
	return nil
}

func CreateDataDescriptor(name string, size int64) *DataDescriptor {
	return &DataDescriptor{
		name: name,
		size: size,
	}
}
