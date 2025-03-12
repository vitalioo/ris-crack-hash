package model

import "encoding/xml"

type HashCrackRequest struct {
	Hash      string `json:"hash"`
	MaxLength int    `json:"maxLength"`
}

type HashCrackResponse struct {
	RequestID string
}

type HashStatusRequest struct {
	RequestID string `json:"requestId"`
}

type HashStatusResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

type HashCrackManagerRequest struct {
	XMLName    xml.Name `xml:"CrackHashManagerRequest"`
	RequestId  string   `xml:"RequestId"`
	PartNumber int      `xml:"PartNumber"`
	PartCount  int      `xml:"PartCount"`
	Hash       string   `xml:"Hash"`
	MaxLength  int      `xml:"MaxLength"`
	Alphabet   Alphabet `xml:"Alphabet"`
}

type Alphabet struct {
	Symbols []string `xml:"symbols"`
}

type WorkerResult struct {
	RequestID string `json:"requestId"`
	Word      string `json:"word"`
}

const (
	READY           = "READY"
	IN_PROGRESS     = "IN_PROGRESS"
	ERROR           = "ERROR"
	PARTITION_READY = "PARTITION_READY"
)
