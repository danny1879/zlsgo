/*
 * @Author: seekwe
 * @Date:   2019-05-15 19:37:01
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-30 13:31:47
 */

package zhttp

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Res struct {
	r      *R
	req    *http.Request
	resp   *http.Response
	client *http.Client
	cost   time.Duration
	*multipartHelper
	requesterBody    []byte
	responseBody     []byte
	downloadProgress DownloadProgress
	err              error
}

func (r *Res) Request() *http.Request {
	return r.req
}

func (r *Res) Response() *http.Response {
	return r.resp
}

func (r *Res) Bytes() []byte {
	data, _ := r.ToBytes()
	return data
}

func (r *Res) ToBytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.responseBody != nil {
		return r.responseBody, nil
	}
	//noinspection GoUnhandledErrorResult
	defer r.resp.Body.Close()
	respBody, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.responseBody = respBody
	return r.responseBody, nil
}

func (r *Res) String() string {
	data, _ := r.ToBytes()
	return string(data)
}

func (r *Res) ToString() (string, error) {
	data, err := r.ToBytes()
	return string(data), err
}

func (r *Res) ToJSON(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (r *Res) ToXML(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

func (r *Res) ToFile(name string) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	if r.responseBody != nil {
		_, err = file.Write(r.responseBody)
		return err
	}

	if r.downloadProgress != nil && r.resp.ContentLength > 0 {
		return r.download(file)
	}
	
	//noinspection GoUnhandledErrorResult
	defer r.resp.Body.Close()
	_, err = io.Copy(file, r.resp.Body)
	return err
}

func (r *Res) download(file *os.File) error {
	p := make([]byte, 1024)
	b := r.resp.Body
	//noinspection GoUnhandledErrorResult
	defer b.Close()
	total := r.resp.ContentLength
	var current int64
	var lastTime time.Time
	for {
		l, err := b.Read(p)
		if l > 0 {
			_, _err := file.Write(p[:l])
			if _err != nil {
				return _err
			}
			current += int64(l)
			if now := time.Now(); now.Sub(lastTime) > 200*time.Millisecond {
				lastTime = now
				r.downloadProgress(current, total)
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}