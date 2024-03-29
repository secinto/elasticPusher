package pusher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
	"strings"
)

var (
	buf bytes.Buffer
)

// NewStore returns a new instance of the store.
func NewStore(c StoreConfig) (*Store, error) {
	indexName := c.IndexName
	projectName := c.ProjectName

	if indexName == "" {
		indexName = ""
	}

	s := Store{es: c.Client, indexName: indexName, projectName: projectName}
	return &s, nil
}

// CreateIndex creates a new index with mapping.
func (s *Store) CreateIndex(mapping string) error {
	res, err := s.es.Indices.Create(s.indexName, s.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error: %s", res)
	}
	return nil
}

// Create indexes a new document into store.
func (s *Store) CreateFromObject(item any) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return s.CreateFromData(payload)
}

// CreateFromData takes a byte array and creates an IndexRequest from it
// and stores it at the defined index.
func (s *Store) CreateFromData(payload []byte) error {
	ctx := context.Background()

	res, err := esapi.CreateRequest{
		Index:      s.indexName,
		DocumentID: uuid.New().String(),
		Body:       bytes.NewReader(payload),
	}.Do(ctx, s.es)

	if err != nil {
		log.Fatalf("Pushing to ELK failed: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}
		return fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	return nil
}

// CreateBulkFromData takes a byte array and creates a BulkRequest from it
// and stores it at the defined index.
func (s *Store) CreateBulkFromData(payload []byte) error {
	createMetaDataForBulkData(payload, s.indexName, s.projectName)
	res, err := s.es.Bulk(bytes.NewReader(buf.Bytes()), s.es.Bulk.WithIndex(s.indexName))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	buf.Reset()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}
		return fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	return nil
}

func createMetaDataForBulkData(payload []byte, indexName string, projectName string) {
	// Prepare the metadata payload
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`, indexName, "\n"))
	// Each line of JSON requires the metadata as trailer to be accepted by elastic. Thus, we read each line
	// and process it separately
	lines := strings.Split(string(payload), "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "}")
		line = line + ",\"project\":\"" + projectName + "\"}"
		linePayload := []byte(line + "\n")
		//Grow the buffer accordingly
		buf.Grow(len(linePayload) + len(meta))
		// Write types to buffer
		buf.Write(meta)
		buf.Write(linePayload)
	}
}
