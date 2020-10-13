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

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Daniel-WWU-IT/libreva/pkg/action"
	"github.com/Daniel-WWU-IT/libreva/pkg/common"
	"github.com/Daniel-WWU-IT/libreva/pkg/reva"
)

func runActions(session *reva.Session) {
	if act, err := action.NewUploadAction(session); err == nil {
		if _, err := act.UploadFile("/home/test.txt", strings.NewReader("Hello World!\n"), common.CreateDataDescriptor("test.txt", 13), false); err == nil {
			//log.Printf("Uploaded file: %q [%db]", info.Path, info.Size)
			log.Println("Upped the stuff")
		} else {
			log.Printf("Can't upload file: %v", err)
		}
	}
	fmt.Println()

	if act, err := action.NewEnumFilesAction(session); err == nil {
		if files, err := act.ListFiles("/home", true); err == nil {
			for _, info := range files {
				fmt.Printf("%s [%db] -- %s\n", info.Path, info.Size, info.Type)

				// Download the file
				if actDl, err := action.NewDownloadAction(session); err == nil {
					if data, err := actDl.DownloadFile(info); err == nil {
						log.Printf("Downloaded %d bytes for '%v'", len(data), info.Path)
					} else {
						log.Printf("Unable to download data for '%v': %v", info.Path, err)
					}
				}

				fmt.Println()
			}
		} else {
			log.Printf("Can't list files: %v", err)
		}
	}
	fmt.Println()
}

func main() {
	if session, err := reva.NewSession(); err == nil {
		if err := session.Initiate("sciencemesh-test.uni-muenster.de:9600", false); err != nil {
			log.Fatalf("Can't initiate Reva session: %v", err)
		}

		if methods, err := session.GetLoginMethods(); err == nil {
			fmt.Println("Supported login methods:")
			for _, m := range methods {
				fmt.Printf("* %v\n", m)
			}
			fmt.Println()
		} else {
			log.Fatalf("Can't list login methods: %v", err)
		}

		if err := session.BasicLogin("daniel", "danielpass"); err == nil {
			log.Printf("Successfully logged into Reva (token=%v)", session.Token())
			fmt.Println()
			runActions(session)
		} else {
			log.Fatalf("Can't log in to Reva: %v", err)
		}
	} else {
		log.Fatalf("Can't create Reva session: %v", err)
	}
}
