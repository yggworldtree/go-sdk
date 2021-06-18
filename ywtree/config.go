package ywtree

type Config struct {
	Host    string `json:"host"`
	Org     string `json:"org"`
	Name    string `json:"name"`
	Alias   string `json:"alias"`
	Subs    string `json:"subs"`
	MaxFreq string `json:"maxFreq"`
	Sign    string `json:"sign"`
	Secret  string `json:"secret"`
}
