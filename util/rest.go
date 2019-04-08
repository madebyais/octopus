package util

import (
	"gopkg.in/resty.v1"
)

// IRestAPI returns available method for rest API
type IRestAPI interface {
	SetHeader(key string, value string)
	Post(url string, body interface{}) (*resty.Response, error)
	Put(url string, body interface{}) (*resty.Response, error)
}

// RestAPI is a schema for rest API object
type RestAPI struct {
	Request *resty.Request
	URL     string
}

// NewRestAPI initiates RestAPI object
func NewRestAPI() IRestAPI {
	return &RestAPI{
		Request: resty.R(),
	}
}

// SetHeader is used to set headers
func (r *RestAPI) SetHeader(key string, value string) {
	r.Request.SetHeader(key, value)
}

// Post is used to send POST request
func (r *RestAPI) Post(url string, body interface{}) (*resty.Response, error) {
	return r.Request.SetBody(body).Post(url)
}

// Put is used to send PUT request
func (r *RestAPI) Put(url string, body interface{}) (*resty.Response, error) {
	return r.Request.SetBody(body).Put(url)
}
