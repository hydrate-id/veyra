package model

type Rule struct {
	PublicPattern  string `json:"public_pattern"`
	InternalTarget string `json:"internal_target"`
}
