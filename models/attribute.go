package models

type Attribute struct {
	ID    int    `json:"id"`
	LogID string `json:"log_id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}