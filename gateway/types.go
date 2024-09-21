package main

type Click struct {
	AdID         string `json:"clickID"`
	ImpressionID string `json:"impressionID"`
}

type CreateAdRequestData struct {
	AdvertiserID string `json:"advertiserID"`
	AdTitle      string `json:"adTitle"`
	AdURL        string `json:"adURL"`
}

type SendClickRequestData struct {
	AdID         string `json:"adID"`
	ImpressionID string `json:"impressionID"`
}

type ClickResponseData struct {
	AdID string `json:"adID"`
	TS   string `json:"ts"`
}
