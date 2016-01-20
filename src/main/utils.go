package main

import (
	"io/ioutil"
	"net/http"
)

type JsonMap map[string]interface{}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Any(vs []string, dst string) bool {
	for _, v := range vs {
		if v == dst {
			return true
		}
	}
	return false
}

func DoGetWithCookies(path string, cookies *string) []byte {
	httpReq, err := http.NewRequest("GET", path, nil)
	if cookies != nil {
		header := http.Header{}
		header.Add("Cookie", *cookies)
		httpReq.Header = header
	}
	resp, err := qBTConn.Client.Do(httpReq)
	Check(err)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data
}
