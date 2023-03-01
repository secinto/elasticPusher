package utils

import (
	"elasticPusher/logging"
	"elasticPusher/types"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	log = logging.NewLogger()
	// Create default settings which ignore insecure certificates and doesn't perform automatic redirect loading
)

func ReadConfigYaml(projectName string) types.Config {

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}

	var config types.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return updateConfig(config, projectName)
}

func updateConfig(config types.Config, projectName string) types.Config {
	config.ProjectName = strings.Replace(config.ProjectName, "{project_name}", projectName, -1)
	return config
}
