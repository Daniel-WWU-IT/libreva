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
	"testing"

	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"

	"github.com/Daniel-WWU-IT/libreva/internal/common"
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