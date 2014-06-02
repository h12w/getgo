// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func MustGet(url string) Response {
	resp, err := http.Get(url)
	checkError(err)
	checkErrorf(resp.StatusCode != http.StatusOK,
		"HTTP response code: %d.", resp.StatusCode)
	return Response{resp.Body}
}

type Response struct {
	rc io.ReadCloser
}

func (r Response) Close() {
	r.rc.Close()
}

func (r Response) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r.rc)
	if err != nil {
		return err
	}
	r.Close()
	return nil
}

func (r Response) SavePrettyJson(file string) error {
	var v interface{}
	if err := json.NewDecoder(r.rc).Decode(&v); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	if _, err := f.Write(buf); err != nil {
		return err
	}
	return nil
}
