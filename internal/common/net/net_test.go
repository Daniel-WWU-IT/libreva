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

package net

import (
	"fmt"
	"strings"
	"testing"

	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/internal/common"
	"github.com/Daniel-WWU-IT/libreva/internal/common/crypto"
)

func TestCheckRPCStatus(t *testing.T) {
	status := rpc.Status{
		Code: rpc.Code_CODE_OK,
	}
	if err := CheckRPCStatus("ok-check", &status); err != nil {
		t.Errorf(common.FormatTestError("CheckRPCStatus", err, "ok-check", status))
	}

	status.Code = rpc.Code_CODE_PERMISSION_DENIED
	if err := CheckRPCStatus("fail-check", &status); err == nil {
		t.Errorf(common.FormatTestError("CheckRPCStatus", fmt.Errorf("accepted an invalid RPC status"), "fail-check", status))
	}
}

func TestTUSClient(t *testing.T) {
	if client, err := NewTUSClient("https://tusd.tusdemo.net/files/", "", ""); err == nil {
		data := strings.NewReader("This is a simple TUS test to the tus.io public tusd server")
		dataDesc := common.CreateDataDescriptor("tus-test.txt", data.Size())
		checksumTypeName := crypto.GetChecksumTypeName(provider.ResourceChecksumType_RESOURCE_CHECKSUM_TYPE_MD5)

		if err := client.Write(data, "tus-test.txt", &dataDesc, checksumTypeName, ""); err != nil {
			t.Errorf(common.FormatTestError("TUSClient.Write", err, data, "tus-test.txt", &dataDesc, checksumTypeName, ""))
		}
	} else {
		t.Errorf(common.FormatTestError("NewTUSClient", err, "http://tusd.tusdemo.net/files/", "", ""))
	}
}
