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
	ELKHost     string `yaml:"elk_host"`
	ProjectName string `yaml:"project_name,omitempty"`
	Username    string `yaml:"username,omitempty"`
	Password    string `yaml:"password,omitempty"`
}

func initialize(project string) {

	appConfig = ReadConfigYaml(project)
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

func ReadConfigYaml(projectName string) Config {

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return updateConfig(config, projectName)
}

func updateConfig(config Config, projectName string) Config {
	config.ProjectName = strings.Replace(config.ProjectName, "{project_name}", projectName, -1)
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
