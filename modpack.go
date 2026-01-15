package main

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	Minecraft       Minecraft `json:"minecraft"`
	ManifestType    string    `json:"manifestType"`
	ManifestVersion int       `json:"manifestVersion"`
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Author          string    `json:"author"`
	Files           []File    `json:"files"`
	Overrides       string    `json:"overrides"`
}

type Minecraft struct {
	Version    string      `json:"version"`
	ModLoaders []ModLoader `json:"modLoaders"`
}

type ModLoader struct {
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
}

type File struct {
	ProjectID int  `json:"projectID"`
	FileID    int  `json:"fileID"`
	Required  bool `json:"required"`
}

func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}
