package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	inputPtr := flag.String("input", "", "Path or URL to the modpack zip file")
	outputPtr := flag.String("output", "./server-files", "Output directory for server files")
	flag.Parse()

	if *inputPtr == "" {
		fmt.Println("Please provide an input modpack using --input")
		os.Exit(1)
	}

	workDir := *outputPtr
	if err := os.MkdirAll(workDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	tempDir := filepath.Join(workDir, "temp_modpack_extract")
	defer os.RemoveAll(tempDir) // Clean up

	modpackPath := *inputPtr

	// Check if input is URL
	if len(*inputPtr) > 4 && ((*inputPtr)[:4] == "http" || (*inputPtr)[:4] == "HTTP") {
		fmt.Println("Downloading modpack...")
		downloadPath := filepath.Join(workDir, "modpack.zip")
		if err := DownloadFile(*inputPtr, downloadPath); err != nil {
			fmt.Printf("Error downloading modpack: %v\n", err)
			os.Exit(1)
		}
		modpackPath = downloadPath
		defer os.Remove(downloadPath)
	}

	fmt.Println("Extracting modpack...")
	if err := Unzip(modpackPath, tempDir); err != nil {
		fmt.Printf("Error extracting modpack: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Reading manifest...")
	manifestPath := filepath.Join(tempDir, "manifest.json")
	// Some modpacks have the manifest inside a subdirectory (rare but possible), we assume root for now
	// Actually, sometimes there is a single folder inside the zip.
	// Let's check if manifest exists, if not, check first child dir?
	// For standard CurseForge packs, it shouldn't be nested.

	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		fmt.Printf("Error reading manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Modpack: %s (Version: %s)\n", manifest.Name, manifest.Version)
	fmt.Printf("Minecraft: %s\n", manifest.Minecraft.Version)

	if err := InstallServer(tempDir, manifest, workDir); err != nil {
		fmt.Printf("Error installing server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done! Server files generated in:", workDir)
	fmt.Println("Note: You may need to accept the EULA in eula.txt before running the server.")
}
