package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/playwright-community/playwright-go"
)

const sessionFilePath = ".session.json"

func SaveSession(storageState *playwright.StorageState) error {
	if storageState == nil {
		return fmt.Errorf("storage state is nil")
	}

	data, err := json.MarshalIndent(storageState, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling storage state: %v", err)
	}

	err = os.WriteFile(sessionFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing session file: %v", err)
	}

	return nil
}

func LoadSession() *playwright.StorageState {
	data, err := os.ReadFile(sessionFilePath)
	if err != nil {
		return nil
	}

	var storageState playwright.StorageState
	err = json.Unmarshal(data, &storageState)
	if err != nil {
		return nil
	}

	return &storageState
}

func DeleteSession() error {
	return os.Remove(sessionFilePath)
}
