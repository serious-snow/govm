package models

import "encoding/xml"

type ListBucketResult struct {
	XMLName     xml.Name `xml:"ListBucketResult"`
	Text        string   `xml:",chardata"`
	Xmlns       string   `xml:"xmlns,attr"`
	Name        string   `xml:"Name"`
	Prefix      string   `xml:"Prefix"`
	Marker      string   `xml:"Marker"`
	NextMarker  string   `xml:"NextMarker"`
	IsTruncated string   `xml:"IsTruncated"`
	Contents    []struct {
		Text           string `xml:",chardata"`
		Key            string `xml:"Key"`
		Generation     string `xml:"Generation"`
		MetaGeneration string `xml:"MetaGeneration"`
		LastModified   string `xml:"LastModified"`
		ETag           string `xml:"ETag"`
		Size           int64  `xml:"Size,string"`
	} `xml:"Contents"`
}

func (list *ListBucketResult) Reset() {
	list.Name = ""
	list.Prefix = ""
	list.Marker = ""
	list.NextMarker = ""
	list.IsTruncated = ""
	list.Contents = nil
}
