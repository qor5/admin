package main

import "time"

type Request struct {
	Sizes struct {
		Headers int `json:"headers"`
		Body    int `json:"body"`
	} `json:"sizes"`
	MimeType        string `json:"mimeType"`
	Charset         any    `json:"charset"`
	ContentEncoding any    `json:"contentEncoding"`
	Header          struct {
		FirstLine string `json:"firstLine"`
		Headers   []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
	} `json:"header"`
	Body struct {
		Text    string `json:"text"`
		Charset any    `json:"charset"`
	} `json:"body"`
}

type Response struct {
	Status int `json:"status"`
	Sizes  struct {
		Headers int `json:"headers"`
		Body    int `json:"body"`
	} `json:"sizes"`
	MimeType        string `json:"mimeType"`
	Charset         string `json:"charset"`
	ContentEncoding any    `json:"contentEncoding"`
	Header          struct {
		FirstLine string `json:"firstLine"`
		Headers   []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
	} `json:"header"`
	Body struct {
		Text    string `json:"text"`
		Charset string `json:"charset"`
	} `json:"body"`
}

type RequestResponse struct {
	Status          string `json:"status"`
	Method          string `json:"method"`
	ProtocolVersion string `json:"protocolVersion"`
	Scheme          string `json:"scheme"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ActualPort      int    `json:"actualPort"`
	Path            string `json:"path"`
	Query           string `json:"query"`
	Tunnel          bool   `json:"tunnel"`
	KeptAlive       bool   `json:"keptAlive"`
	WebSocket       bool   `json:"webSocket"`
	RemoteAddress   string `json:"remoteAddress"`
	ClientAddress   string `json:"clientAddress"`
	ClientPort      int    `json:"clientPort"`
	Times           struct {
		Start           time.Time `json:"start"`
		RequestBegin    time.Time `json:"requestBegin"`
		RequestComplete time.Time `json:"requestComplete"`
		ResponseBegin   time.Time `json:"responseBegin"`
		End             time.Time `json:"end"`
	} `json:"times"`
	Durations struct {
		Total    int `json:"total"`
		DNS      any `json:"dns"`
		Connect  any `json:"connect"`
		Ssl      any `json:"ssl"`
		Request  int `json:"request"`
		Response int `json:"response"`
		Latency  int `json:"latency"`
	} `json:"durations"`
	Speeds struct {
		Overall  int `json:"overall"`
		Request  int `json:"request"`
		Response int `json:"response"`
	} `json:"speeds"`
	TotalSize int      `json:"totalSize"`
	Request   Request  `json:"request"`
	Response  Response `json:"response"`
}
