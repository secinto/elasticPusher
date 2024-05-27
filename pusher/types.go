package pusher

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
	"time"
)

const VERSION = "0.3.6"

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

type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project"`
	HostName    string    `json:"host"`
	Entry       string    `json:"entry"`
	Level       string    `json:"level"`
}

type Options struct {
	ConfigFile string
	InputFile  string
	Project    string
	Index      string
	Host       string
	Type       string
	Silent     bool
	Version    bool
	NoColor    bool
	Verbose    bool
}

type Config struct {
	ELKHost  string `yaml:"elk_host"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	APIKey   string `yaml:"apikey,omitempty"`
}

type Pusher struct {
	InputFile string
	Project   string
	Index     string
	Host      string
	Type      string
}

// StoreConfig configures the store.
type StoreConfig struct {
	Client      *elasticsearch.Client
	IndexName   string
	ProjectName string
}

// Store allows to index and search documents.
type Store struct {
	es          *elasticsearch.Client
	indexName   string
	projectName string
}

// Hook is a hook that writes logs of specified LogLevels to specified Writer
type Hook struct {
	Pusher    *Pusher
	Formatter logrus.Formatter
	LogLevel  logrus.Level
}
