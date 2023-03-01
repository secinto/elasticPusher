package main

import (
	"elasticPusher/elastic"
	"elasticPusher/logging"
	"elasticPusher/types"
	"elasticPusher/utils"
	"encoding/json"
	"flag"
	"github.com/elastic/go-elasticsearch/v8"
	"os"
	"strings"
	"time"
)

var (
	log       = logging.NewLogger()
	appConfig types.Config
	esClient  *elasticsearch.Client
)

type Interaction struct {
	Timestamp   time.Time `json:"timestamp"`
	ProjectName string    `json:"project"`
	Hostname    string    `json:"host"`
	Data        string    `json:"interaction"`
}

func initialize(project string) {

	appConfig = utils.ReadConfigYaml(project)
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

func SaveInteractionToElk(file string, indexName string, project string, host string) {
	config := elastic.StoreConfig{Client: esClient, IndexName: indexName}
	store, err := elastic.NewStore(config)
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
		error := store.CreateFromData(jsonInteraction)
		if error != nil {
			log.Errorf("Pushing failed: %v", error)
		}
	}

}

func SaveAnyToElk(file string, indexName string, project string) {
	config := elastic.StoreConfig{Client: esClient, IndexName: indexName}
	store, err := elastic.NewStore(config)
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
		error := store.CreateFromData(jsonInteraction)
		if error != nil {
			log.Errorf("Pushing failed: %v", error)
		}
	}

}

func SaveAnyJSONLToElk(file string, indexName string, project string, debugOutput bool) {
	storeConfig := elastic.StoreConfig{Client: esClient, IndexName: indexName, ProjectName: project, DebugOutput: debugOutput}
	store, err := elastic.NewStore(storeConfig)
	if err != nil {
		log.Fatal("Error creating new ELK store: %v", err)
	}
	bytes, _ := os.ReadFile(file)
	error := store.CreateBulkFromData(bytes)
	if error != nil {
		log.Errorf("Pushing failed: %v", error)
	}
}

func main() {

	var fileToPush string
	var project string
	var indexName string
	var hostName string
	var inputType string
	var debugOutput bool

	flag.StringVar(&fileToPush, "f", "", "The name and location of the file which should be sent to ELK")
	flag.StringVar(&project, "p", "general", "Specify what type of file format is sent (e.g.: JSON, XML, CSV,...)")
	flag.StringVar(&indexName, "i", "", "Specify under what index name the types should be stored")
	flag.StringVar(&hostName, "h", "", "Specify the host for which the data has been obtained")
	flag.StringVar(&inputType, "t", "", "Specify what type of document should be sent (json, raw)")
	flag.BoolVar(&debugOutput, "d", false, "Provides debug output of the operation")
	flag.Parse()

	initialize(project)

	// 0.10 - First version
	// 0.11 - Updates on adding index name dynamically and bug fixes
	// 0.12 - Adding project name via parameters and storing via metadata (because max indices 500)
	// 0.13 - Added types and configuration handling which seem to have a limit
	// 0.14 - Adding handling of other files than JSON (project info not added here)
	// 0.15 - Adding storing of responses which also allow to add the host where it was obtained as metadata
	log.Info("Running elasticPusher v0.15")

	if strings.ToLower(inputType) == "json" {
		log.Infof("Storing file %s in index %s", fileToPush, indexName)
		SaveAnyJSONLToElk(fileToPush, indexName, project, debugOutput)
	} else if strings.ToLower(inputType) == "raw" {
		log.Infof("Storing file\n %s in index %s\n for project %s from host %s", fileToPush, indexName, project, hostName)
		SaveInteractionToElk(fileToPush, indexName, project, hostName)
	} else {
		log.Infof("Storing file %s in index %s", fileToPush, indexName)
		SaveAnyToElk(fileToPush, indexName, project)
	}
	log.Infof("elasticPusher successful. Exit now")

}
