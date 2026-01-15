package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func InstallServer(manifestDir string, manifest *Manifest, outputDir string) error {
	modsDir := filepath.Join(outputDir, "mods")
	if err := os.MkdirAll(modsDir, os.ModePerm); err != nil {
		return err
	}

	fmt.Printf("Downloading %d mods...\n", len(manifest.Files))

	for i, file := range manifest.Files {
		url := GetModDownloadURL(file.ProjectID, file.FileID)
		fmt.Printf("[%d/%d] Downloading Mod ID: %d File ID: %d...\n", i+1, len(manifest.Files), file.ProjectID, file.FileID)

		// Download Mod
		destPath, err := DownloadMod(url, modsDir)
		if err != nil {
			fmt.Printf("Failed to download mod %d (File %d): %v\n", file.ProjectID, file.FileID, err)
			if file.Required {
				return err // Only return error if mod is required
			}
			// Optional mod? For now, we assume all are required or we just skip
			continue
		}

		filename := filepath.Base(destPath)
		fmt.Printf("Downloaded: %s\n", filename)

		// Rename to real filename if possible?
		// The generic download URL redirects to the real filename.
		// Our DownloadFile implementation saves to the destPath provided.
		// So we have "ProjectID-FileID.jar".
		// We can't easily get the real filename without inspecting the HTTP response headers in DownloadFile.

		// Check Client Side
		isClient, modId, err := IsClientSideOnly(destPath)
		if err != nil {
			fmt.Printf("Warning: Failed to check if mod is client side: %v\n", err)
		} else {
			// Check Blocklist (internal mod ID)
			blocklist := []string{
				"essential", "essential-mod", "essentials",
				"essential_mod", "essentialclient",
				// seemingly client-side only mods that might be missed
				"oculus", "rubidium", "embeddium",
				"controlling", "searchables",
				"toastcontrol", "mouseedits",
				"catalogue", "configured",
			}

			for _, blocked := range blocklist {
				if strings.Contains(strings.ToLower(modId), blocked) || strings.Contains(strings.ToLower(filename), blocked) {
					isClient = true // Force client side
					fmt.Printf("Mod '%s' (ID: %s) matched blocklist '%s'\n", filename, modId, blocked)
					break
				}
			}

			if isClient {
				fmt.Printf("Removing client-side mod: %s (ModID: %s)\n", filename, modId)
				os.Remove(destPath)
			}
		}
	}

	// Copy Overrides
	if manifest.Overrides != "" {
		overridesPath := filepath.Join(manifestDir, manifest.Overrides)
		fmt.Printf("Copying overrides from %s to %s...\n", overridesPath, outputDir)
		err := CopyDir(overridesPath, outputDir)
		if err != nil {
			fmt.Printf("Warning: Failed to copy overrides: %v\n", err)
		}
	}

	// Handle ModLoader (Forge)
	if len(manifest.Minecraft.ModLoaders) > 0 {
		loader := manifest.Minecraft.ModLoaders[0] // Assume first is primary
		if strings.HasPrefix(loader.ID, "forge") {
			forgeVersion := strings.TrimPrefix(loader.ID, "forge-")
			mcVersion := manifest.Minecraft.Version
			installerUrl := fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/forge-%s-%s-installer.jar", mcVersion, forgeVersion, mcVersion, forgeVersion)

			installerPath := filepath.Join(outputDir, "forge-installer.jar")
			fmt.Printf("Downloading Forge Installer %s...\n", installerUrl)
			if err := DownloadFile(installerUrl, installerPath); err != nil {
				fmt.Printf("Failed to download forge installer: %v\n", err)
			} else {
				fmt.Println("Forge Installer downloaded. Run: java -jar forge-installer.jar --installServer")
				// Create a helper script
				scriptContent := "java -jar forge-installer.jar --installServer\n"
				os.WriteFile(filepath.Join(outputDir, "install_forge.sh"), []byte(scriptContent), 0755)
			}
		}
	}

	return nil
}

func CopyDir(src string, dest string) error {
	// Simple recursive copy
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}
