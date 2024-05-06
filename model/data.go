package model

type UserMessageData struct {
	Token          string   `json:"Token"`
	AgreementStatus string  `json:"AgreementStatus"`
	Reason         []string `json:"Reason"`
	UserMessage    string   `json:"UserMessage"`
}