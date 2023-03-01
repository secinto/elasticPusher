package pusher

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

var (
	log       = NewLogger()
	appConfig Config
	esClient  *elasticsearch.Client
)

const VERSION = "0.1"

type Interaction struct {
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project"`
	Hostname    string    `json:"host"`
	Data        string    `json:"interaction"`
}

type Config struct {
	ELKHost  string `yaml:"elk_host"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Pusher struct {
	options *Options
}

func NewPusher(options *Options) (*Pusher, error) {
	pusher := &Pusher{options: options}
	initialize(options.ConfigFile)
	return pusher, nil
}

func (p *Pusher) Push() error {
	if strings.ToLower(p.options.Type) == "json" {
		log.Infof("Pushing JSONL file %s to index %s for project %s", p.options.InputFile, p.options.Index, p.options.Project)
		SaveAnyJSONLToElk(p.options.InputFile, p.options.Index, p.options.Project, p.options.Verbose)
	} else if strings.ToLower(p.options.Type) == "raw" {
		log.Infof("Pushing RAW file %s to index %s for project %s and host %s", p.options.InputFile, p.options.Index, p.options.Project, p.options.Host)
		SaveInteractionToElk(p.options.InputFile, p.options.Index, p.options.Project, p.options.Host)
	} else {
		log.Infof("Pushing other file %s to index %s for project %s", p.options.InputFile, p.options.Index, p.options.Project)
		SaveAnyToElk(p.options.InputFile, p.options.Index, p.options.Project)
	}
	log.Infof("elasticPusher finished")
	return nil
}

func initialize(configLocation string) {
	appConfig = loadConfigFrom(configLocation)
	if appConfig.ELKHost == "" {
		appConfig.ELKHost = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			appConfig.ELKHost,
		},
		Username: appConfig.Username,
		Password: appConfig.Password,
	}

	log.Infof("Using %s as server", cfg.Addresses[0])

	var err error
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal("Error creating Elasticsearch client: %v", err)
	}
}

func loadConfigFrom(location string) Config {
	var config Config
	var yamlFile []byte
	var err error

	yamlFile, err = os.ReadFile(location)
	if err != nil {
		path, err := os.Getwd()
		if err != nil {
			log.Fatalf("yamlFile.Get err   #%v ", err)
		}

		yamlFile, err = os.ReadFile(path + "\\config.yaml")
		if err != nil {
			log.Fatalf("yamlFile.Get err   #%v ", err)
		}
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return config
}

func SaveInteractionToElk(file string, indexName string, project string, host string) {
	config := StoreConfig{Client: esClient, IndexName: indexName}
	store, err := NewStore(config)
	if err != nil {
		log.Fatal("Error creating new ELK store: %v", err)
	}

	bytes, _ := os.ReadFile(file)

	interaction := Interaction{
		Timestamp:   time.Now(),
		ProjectName: project,
		Hostname:    host,
		Data:        string(bytes),
	}

	if jsonInteraction, err := json.Marshal(interaction); err == nil {
		err := store.CreateFromData(jsonInteraction)
		if err != nil {
			log.Errorf("Pushing failed: %v", err)
		}
	}

}

func SaveAnyToElk(file string, indexName string, project string) {
	config := StoreConfig{Client: esClient, IndexName: indexName}
	store, err := NewStore(config)
	if err != nil {
		log.Fatal("Error creating new ELK store: %v", err)
	}
	bytes, _ := os.ReadFile(file)

	interaction := Interaction{
		Timestamp:   time.Now(),
		ProjectName: project,
		Data:        string(bytes),
	}

	if jsonInteraction, err := json.Marshal(interaction); err == nil {
		err = store.CreateFromData(jsonInteraction)
		if err != nil {
			log.Errorf("Pushing failed: %v", err)
		}
	}

}

func SaveAnyJSONLToElk(file string, indexName string, project string, debugOutput bool) {
	storeConfig := StoreConfig{Client: esClient, IndexName: indexName, ProjectName: project, DebugOutput: debugOutput}
	store, err := NewStore(storeConfig)
	if err != nil {
		log.Fatal("Error creating new ELK store: %v", err)
	}
	bytes, _ := os.ReadFile(file)
	err = store.CreateBulkFromData(bytes)
	if err != nil {
		log.Errorf("Pushing failed: %v", err)
	}
}
