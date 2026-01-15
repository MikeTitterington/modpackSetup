package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"strings"
)

type FabricMod struct {
	Environment string `json:"environment"`
}

func IsClientSideOnly(jarPath string) (bool, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return false, err
	}
	defer r.Close()

	// Check for fabric.mod.json (Fabric/Quilt)
	for _, f := range r.File {
		if f.Name == "fabric.mod.json" {
			rc, err := f.Open()
			if err != nil {
				return false, err
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return false, err
			}

			var mod FabricMod
			if err := json.Unmarshal(data, &mod); err == nil {
				if mod.Environment == "client" {
					return true, nil
				}
			}
		}
	}

	// Check for META-INF/mods.toml (Forge/NeoForge)
	for _, f := range r.File {
		if f.Name == "META-INF/mods.toml" {
			rc, err := f.Open()
			if err != nil {
				return false, err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return false, err
			}

			content := string(data)
			// Heuristic: Check for displayTest="IGNORE_ALL_VERSION" which often indicates client-side cosmetic mods
			if strings.Contains(content, "displayTest=\"IGNORE_ALL_VERSION\"") {
				return true, nil
			}

			// Heuristic: specific side="CLIENT"
			// This is harder to match perfectly with string contains but "side=\"CLIENT\"" is a good guess
			if strings.Contains(content, "side=\"CLIENT\"") {
				return true, nil
			}
		}
	}

	return false, nil
}
