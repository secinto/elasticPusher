package pusher

import (
	"crypto/tls"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	log      = NewLogger()
	esClient *elasticsearch.Client
)

func FromOptions(options *Options) (*Pusher, error) {
	pusher := &Pusher{InputFile: options.InputFile, Index: options.Index, Project: options.Project, Type: options.Type, Host: options.Host}
	config := loadConfigFrom(options.ConfigFile)

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.ELKHost,
		},
	}

	if config.APIKey != "" {
		cfg.Username = config.Username
		cfg.Password = config.Password
	} else {
		cfg.APIKey = config.APIKey
	}

	initialize(cfg)

	return pusher, nil
}

func initialize(config elasticsearch.Config) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	log.Infof("Using %s as elasticsearch server", config.Addresses[0])

	var err error
	esClient, err = elasticsearch.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}
}

func loadConfigFrom(location string) Config {
	var config Config
	var yamlFile []byte
	var err error

	yamlFile, err = os.ReadFile(location)
	if err != nil {
		log.Fatalf("Could read yaml config file: %v ", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return config
}

func (p *Pusher) PushFile() error {
	if p.InputFile != "" {
		if strings.ToLower(p.Type) == "json" {
			log.Infof("Pushing JSON file %s to index %s for project %s", p.InputFile, p.Index, p.Project)
			SaveAnyJSONLToElk(p.InputFile, p.Index, p.Project)
		} else if strings.ToLower(p.Type) == "raw" {
			log.Infof("Pushing RAW file %s to index %s for project %s and host %s", p.InputFile, p.Index, p.Project, p.Host)
			SaveInteractionToElk(p.InputFile, p.Index, p.Project, p.Host)
		} else {
			log.Infof("Pushing other file %s to index %s for project %s", p.InputFile, p.Index, p.Project)
			SaveAnyToElk(p.InputFile, p.Index, p.Project)
		}
		log.Infof("elasticPusher finished")
	} else {
		log.Infof("No input specified (piping currently not supported). Exiting program!")
	}

	return nil
}

func (p *Pusher) PushLog(json string, level string) error {
	sendLogToELK(json, p.Index, level, p.Host, p.Project)
	return nil
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

func SaveAnyJSONLToElk(file string, indexName string, project string) {
	storeConfig := StoreConfig{Client: esClient, IndexName: indexName, ProjectName: project}
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

func sendLogToELK(logEntry string, indexName string, level string, host string, project string) {
	config := StoreConfig{Client: esClient, IndexName: indexName}
	store, err := NewStore(config)
	if err != nil {
		log.Fatal("Error creating new ELK store: %v", err)
	}

	entry := LogEntry{
		Timestamp:   time.Now(),
		ProjectName: project,
		HostName:    host,
		Entry:       logEntry,
		Level:       level,
	}

	if jsonEntry, err := json.Marshal(entry); err == nil {
		err := store.CreateFromData(jsonEntry)
		if err != nil {
			log.Errorf("Pushing failed: %v", err)
		}
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
