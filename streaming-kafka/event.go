package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type event struct {
	TS      int64
	Gender  string
	Age     int64
	Path    string
	Browser string
	OS      string
}

func getInt64(v interface{}) (int64, error) {
	switch v.(type) {
	// Golang decodes JSON number as float64
	case float64:
		return int64(v.(float64)), nil
	default:
		return 0, fmt.Errorf("Can't get int64 from %++v", v)
	}
}

func getString(v interface{}) (string, error) {
	switch v.(type) {
	case string:
		return strings.ToLower(v.(string)), nil
	default:
		return "", fmt.Errorf("Can't get string from %++v", v)
	}
}

// Custom unmarshal func because of:
// - count unknown fields in json
// - map multiple fields to one value
func (e *event) unmarshalJSON(j []byte) error {
	var rawStrings map[string]interface{}

	err := json.Unmarshal(j, &rawStrings)
	if err != nil {
		log.Printf("WARN: Can't unmashal INPUT JSON: %s, JSON Input: %s\n", err.Error(), string(j[:]))
		return err
	}

	for k, v := range rawStrings {
		switch strings.ToLower(k) {
		case "ts":
			e.TS, err = getInt64(v)
			if err != nil {
				return err
			}
		case "gender":
			e.Gender, err = getString(v)
			if err != nil {
				return err
			}
		case "age":
			e.Age, err = getInt64(v)
			if err != nil {
				return err
			}
		case "path":
			e.Path, err = getString(v)
			if err != nil {
				return err
			}
		case "browser":
			e.Browser, err = getString(v)
			if err != nil {
				return err
			}
		case "os":
			e.OS, err = getString(v)
			if err != nil {
				return err
			}
		default:
			jsonUnknownFields.Inc()
			log.Printf("WARN: Unknown field in json: %s = %v\n", k, v)
		}
	}
	return nil
}
