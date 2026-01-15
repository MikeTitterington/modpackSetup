package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"regexp"
)

type FabricMod struct {
	Environment string `json:"environment"`
}

func IsClientSideOnly(jarPath string) (bool, string, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return false, "", err
	}
	defer r.Close()

	// Check for fabric.mod.json (Fabric/Quilt)
	for _, f := range r.File {
		if f.Name == "fabric.mod.json" {
			rc, err := f.Open()
			if err != nil {
				return false, "", err
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return false, "", err
			}

			var mod FabricMod
			if err := json.Unmarshal(data, &mod); err == nil {
				if mod.Environment == "client" {
					return true, "", nil
				}
			}
		}
	}

	// Check for META-INF/mods.toml (Forge/NeoForge)
	for _, f := range r.File {
		if f.Name == "META-INF/mods.toml" {
			rc, err := f.Open()
			if err != nil {
				return false, "", err
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return false, "", err
			}

			content := string(data)
			// Heuristic: displayTest="IGNORE_ALL_VERSION" is too aggressive and deletes server mods like AttributeFix.
			// Removing this check. Rely on Blocklist for client-side filtering.

			// Extract Mod ID
			// modId="name" or modId = "name"
			re := regexp.MustCompile(`modId\s*=\s*"([^"]+)"`)
			matches := re.FindStringSubmatch(content)
			var modId string
			if len(matches) > 1 {
				modId = matches[1]
			}

			return false, modId, nil
		}
	}

	return false, "", nil
}
