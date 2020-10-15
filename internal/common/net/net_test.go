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

package net_test

import (
	"fmt"
	"strings"
	"testing"

	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/internal/common"
	"github.com/Daniel-WWU-IT/libreva/internal/common/crypto"
	"github.com/Daniel-WWU-IT/libreva/internal/common/net"
	testintl "github.com/Daniel-WWU-IT/libreva/internal/testing"
)

func TestCheckRPCStatus(t *testing.T) {
	status := rpc.Status{
		Code: rpc.Code_CODE_OK,
	}
	if err := net.CheckRPCStatus("ok-check", &status); err != nil {
		t.Errorf(testintl.FormatTestError("CheckRPCStatus", err, "ok-check", status))
	}

	status.Code = rpc.Code_CODE_PERMISSION_DENIED
	if err := net.CheckRPCStatus("fail-check", &status); err == nil {
		t.Errorf(testintl.FormatTestError("CheckRPCStatus", fmt.Errorf("accepted an invalid RPC status"), "fail-check", status))
	}
}

func TestTUSClient(t *testing.T) {
	tests := []struct {
		endpoint      string
		shouldSucceed bool
	}{
		{"https://tusd.tusdemo.net/files/", true},
		{"https://google.de", false},
	}

	for _, test := range tests {
		t.Run(test.endpoint, func(t *testing.T) {
			if client, err := net.NewTUSClient(test.endpoint, "", ""); err == nil {
				data := strings.NewReader("This is a simple TUS test")
				dataDesc := common.CreateDataDescriptor("tus-test.txt", data.Size())
				checksumTypeName := crypto.GetChecksumTypeName(provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_MD5)

				if err := client.Write(data, dataDesc.Name(), &dataDesc, checksumTypeName, ""); err != nil && test.shouldSucceed {
					t.Errorf(testintl.FormatTestError("TUSClient.Write", err, data, dataDesc.Name(), &dataDesc, checksumTypeName, ""))
				} else if err == nil && !test.shouldSucceed {
					t.Errorf(testintl.FormatTestError("TUSClient.Write", fmt.Errorf("writing to a non-TUS host succeeded"), data, dataDesc.Name(), &dataDesc, checksumTypeName, ""))
				}
			} else {
				t.Errorf(testintl.FormatTestError("NewTUSClient", err, test.endpoint, "", ""))
			}
		})
	}
}

func TestWebDAVClient(t *testing.T) {
	tests := []struct {
		endpoint      string
		shouldSucceed bool
	}{
		{"https://zivowncloud2.uni-muenster.de/owncloud/remote.php/dav/files/testUser/", true},
		{"https://google.de", false},
	}

	for _, test := range tests {
		t.Run(test.endpoint, func(t *testing.T) {
			if client, err := net.NewWebDAVClient(test.endpoint, "testUser", "test12345"); err == nil {
				const fileName = "webdav-test.txt"

				data := strings.NewReader("This is a simple WebDAV test")
				if err := client.Write(fileName, data, data.Size()); err == nil && test.shouldSucceed {
					if _, err := client.Read(fileName); err != nil {
						t.Errorf(testintl.FormatTestError("WebDAVClient.Read", err))
					}

					if err := client.Remove(fileName); err != nil {
						t.Errorf(testintl.FormatTestError("WebDAVClient.Remove", err))
					}
				} else if err != nil && test.shouldSucceed {
					t.Errorf(testintl.FormatTestError("WebDAVClient.Write", err, fileName, data, data.Size()))
				} else if err == nil && !test.shouldSucceed {
					t.Errorf(testintl.FormatTestError("WebDAVClient.Write", fmt.Errorf("writing to a non-WebDAV host succeeded"), fileName, data, data.Size()))
				}
			} else {
				t.Errorf(testintl.FormatTestError("NewWebDavClient", err, test.endpoint, "testUser", "test12345"))
			}
		})
	}
}
