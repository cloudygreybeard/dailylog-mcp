package cmd

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// outputJSON outputs data as formatted JSON
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputYAML outputs data as formatted YAML
func outputYAML(data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}
	fmt.Println(string(yamlData))
	return nil
}

