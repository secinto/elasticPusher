package pusher

import (
	"github.com/elastic/go-elasticsearch/v8"
	"time"
)

const VERSION = "0.1.9"

type Interaction struct {
	Timestamp             time.Time `json:"timestamp"`
	ProjectName           string    `json:"project"`
	HostName              string    `json:"host"`
	URL                   string    `json:"url"`
	Raw                   string    `json:"raw,omitempty"`
	ResponseHTTPVersion   string    `json:"responseHTTPVersion,omitempty"`
	ResponseStatusCode    string    `json:"responseStatusCode,omitempty"`
	ResponseStatusMessage string    `json:"responseStatusMessage,omitempty"`
	RequestHeader         string    `json:"requestHeader"`
	RequestBody           string    `json:"requestBody,omitempty"`
	ResponseHeader        string    `json:"responseHeader"`
	ResponseBody          string    `json:"responseBody"`
}

type Config struct {
	ELKHost  string `yaml:"elk_host"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Pusher struct {
	options *Options
}

// StoreConfig configures the store.
type StoreConfig struct {
	Client      *elasticsearch.Client
	IndexName   string
	ProjectName string
	DebugOutput bool
}

// Store allows to index and search documents.
type Store struct {
	es          *elasticsearch.Client
	indexName   string
	projectName string
	debug       bool
}
