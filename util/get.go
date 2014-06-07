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

// MustGet gets the response of a URL or panics if any error occurs.
func MustGet(url string) Response {
	resp, err := http.Get(url)
	checkError(err)
	checkErrorf(resp.StatusCode != http.StatusOK,
		"HTTP response code: %d.", resp.StatusCode)
	return Response{resp.Body}
}

// Response provides convenient methods to save an HTTP response's body.
type Response struct {
	rc io.ReadCloser
}

// Close the reponse.
func (r Response) Close() {
	r.rc.Close()
}

// Save the response body to a file.
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

// SavePrettyJSON prettily formats a JSON response and save it to a file.
func (r Response) SavePrettyJSON(file string) error {
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
