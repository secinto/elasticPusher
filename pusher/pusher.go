package pusher

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	log       = NewLogger()
	appConfig Config
	esClient  *elasticsearch.Client
)

func NewPusher(options *Options) (*Pusher, error) {
	pusher := &Pusher{options: options}
	initialize(options.ConfigFile)
	return pusher, nil
}

func (p *Pusher) Push() error {
	if p.options.InputFile != "" {
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
	} else {
		log.Infof("No input specified (piping currently not supported). Exiting program!")
	}

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
		log.Fatalf("yamlFile.Get err   #%v ", err)
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

	interaction := Interaction{
		Timestamp:   time.Now(),
		ProjectName: project,
		HostName:    host,
	}

	parseResponses(file, &interaction)

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
		Raw:         string(bytes),
	}

	parseResponses(file, &interaction)

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

func parseResponses(fileName string, interaction *Interaction) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileLines := strings.Split(string(file), "\n")

	var isResponse = false
	var isBody = false
	var newlineReceived = false
	var header []string
	var body []string
	var responseBodyStartIndex int
	var responseBodyEndIndex int
	var responseHeaderStart bool = false

	for index, line := range fileLines {
		if strings.TrimSpace(line) != "" {
			if index == 0 {
				interaction.URL = strings.TrimSpace(line)
				newlineReceived = false
			} else if index >= 2 && !isResponse {
				// Request header
				header = append(header, line)
				newlineReceived = false
			} else if isResponse && !isBody {
				// Response header
				if responseHeaderStart {
					responseHeaderStart = false
					parts := strings.Split(line, " ")
					if len(parts) >= 3 {
						interaction.ResponseHTTPVersion = parts[0]
						interaction.ResponseStatusCode = parts[1]
						var message []string
						if len(parts) > 3 {
							for i := 2; i < len(parts); i++ {
								message = append(message, parts[i])
							}
						} else {
							message = append(message, parts[2])
						}
						interaction.ResponseStatusMessage = strings.Join(message, " ")
					}

				}
				header = append(header, line)
				newlineReceived = false
			} else if isResponse && isBody {
				// Response body
				// Currently the amount of body characters, which is sometimes the first line after the newline is not
				// evaluated and verified. If the first line is the length, then the last line is 0.
				if responseBodyStartIndex == index {
					_, err := strconv.ParseUint(line, 16, 64)
					if err != nil {
						body = append(body, line+"\n")
					}
				} else if responseBodyEndIndex == index {
					_, err := strconv.ParseUint(line, 16, 64)
					if err != nil {
						body = append(body, line+"\n")
					}
				} else {
					body = append(body, line+"\n")
				}
			}
		} else {
			if index > 3 && isResponse == false && !newlineReceived {
				isResponse = true
				isBody = false
				interaction.RequestHeader = strings.Join(header, "\n")
				header = nil
				//Response header
				newlineReceived = true
				responseHeaderStart = true
			} else if index > 3 && isResponse == true && !newlineReceived {
				//Response body
				isBody = true
				body = nil
				interaction.ResponseHeader = strings.Join(header, "\n")
				newlineReceived = true
				responseBodyStartIndex = index + 1
				responseBodyEndIndex = len(fileLines) - 2
			} else if isResponse && isBody {
				body = append(body, line+"\n")
			}
		}
	}

	interaction.ResponseBody = strings.Join(body, "\n")
}
