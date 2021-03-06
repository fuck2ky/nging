/*

   Copyright 2016 Wenhui Shen <www.webx.top>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

*/

package upload

import (
	"bytes"
	"io"
	"io/ioutil"
	"path"

	"github.com/webx-top/client/upload/watermark"
	"github.com/webx-top/echo"
)

var clients = make(map[string]func() Client)
var defaults = New(nil)

func Register(name string, c func() Client) {
	clients[name] = c
}

func Get(name string) Client {
	fn, _ := clients[name]
	if fn == nil {
		fn = func() Client {
			return defaults
		}
	}
	return fn()
}

func Has(name string) bool {
	_, ok := clients[name]
	return ok
}

func Delete(name string) {
	if _, ok := clients[name]; ok {
		delete(clients, name)
	}
}

type Storer interface {
	Put(dstFile string, body io.Reader, size int64) (savePath string, fileURL string, err error)
}

func Upload(ctx echo.Context, clientName string, result *Result, storer Storer, markFile string, checkers ...func(*Result) error) Client {
	client := Get(clientName)
	client.Init(ctx, result)
	body, err := client.Body()
	if err != nil {
		return client.SetError(err)
	}
	defer body.Close()
	if len(checkers) > 0 {
		err = checkers[0](result)
		if err != nil {
			return client.SetError(err)
		}
	}
	dstFile, err := result.GenFileName()
	if err != nil {
		return client.SetError(err)
	}
	if len(markFile) > 0 && result.FileType.String() == `image` {
		var b []byte
		b, err = ioutil.ReadAll(body)
		if err != nil {
			return client.SetError(err)
		}
		b, err = watermark.Bytes(b, path.Ext(result.FileName), markFile)
		if err != nil {
			return client.SetError(err)
		}
		byteReader := bytes.NewReader(b)
		result.SavePath, result.FileURL, err = storer.Put(dstFile, byteReader, byteReader.Size())
	} else {
		result.SavePath, result.FileURL, err = storer.Put(dstFile, body, body.Size())
	}
	if err != nil {
		return client.SetError(err)
	}
	return client
}
