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

package common

import (
	"os"
	"testing"
	"time"
)

func TestDataDescriptor(t *testing.T) {
	const name = "DATA_DESC"
	const size = 42

	dataDesc := CreateDataDescriptor(name, size)
	now := time.Now().Round(time.Millisecond)
	if v := dataDesc.Name(); v != name {
		t.Errorf(FormatTestResult("DataDescriptor.Name", name, v))
	}
	if v := dataDesc.Size(); v != size {
		t.Errorf(FormatTestResult("DataDescriptor.Size", size, v))
	}
	if v := dataDesc.Mode(); v != os.ModePerm {
		t.Errorf(FormatTestResult("DataDescriptor.Mode", os.ModePerm, v))
	}
	if v := dataDesc.IsDir(); v != false {
		t.Errorf(FormatTestResult("DataDescriptor.IsDir", false, v))
	}
	if v := dataDesc.ModTime(); !v.Round(time.Millisecond).Equal(now) {
		// Since there's always a slight chance that the rounded times won't match, just log this mismatch
		t.Logf(FormatTestResult("DataDescriptor.ModTime", now, v))
	}
	if v := dataDesc.Sys(); v != nil {
		t.Errorf(FormatTestResult("DataDescriptor.Sys", nil, v))
	}
}

func TestFindString(t *testing.T) {
	tests := []struct {
		input  []string
		needle string
		wants  int
	}{
		{[]string{}, "so empty", -1},
		{[]string{"12345", "hello", "goodbye"}, "hello", 1},
		{[]string{"Rudimentär", "Ich bin du", "Wüste", "SANDIGER GRUND"}, "Wüste", 2},
		{[]string{"Rudimentär", "Ich bin du", "Wüste", "SANDIGER GRUND", "Sandiger Grund"}, "Sandiger Grund", 4},
		{[]string{"Nachahmer", "Roger", "k thx bye"}, "thx", -1},
		{[]string{"Six Feet Under", "Rock&Roll", "k thx bye"}, "Six Feet Under", 0},
		{[]string{"Six Feet Under", "Rock&Roll", "k thx bye"}, "Six Feet UNDER", -1},
	}

	for _, test := range tests {
		found := FindString(test.input, test.needle)
		if found != test.wants {
			t.Errorf(FormatTestResult("FindString", test.wants, found, test.input, test.needle))
		}
	}
}

func TestFindStringNoCase(t *testing.T) {
	tests := []struct {
		input  []string
		needle string
		wants  int
	}{
		{[]string{}, "so empty", -1},
		{[]string{"12345", "hello", "goodbye"}, "hello", 1},
		{[]string{"Rudimentär", "Ich bin du", "Wüste", "SANDIGER GRUND"}, "Wüste", 2},
		{[]string{"Rudimentär", "Ich bin du", "Wüste", "SANDIGER GRUND", "Sandiger Grund"}, "Sandiger Grund", 3},
		{[]string{"Nachahmer", "Roger", "k thx bye"}, "thx", -1},
		{[]string{"Six Feet Under", "Rock&Roll", "k thx bye"}, "Six Feet Under", 0},
		{[]string{"Six Feet Under", "Rock&Roll", "k thx bye"}, "Six Feet UNDER", 0},
	}

	for _, test := range tests {
		found := FindStringNoCase(test.input, test.needle)
		if found != test.wants {
			t.Errorf(FormatTestResult("FindString", test.wants, found, test.input, test.needle))
		}
	}
}
