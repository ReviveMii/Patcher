package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	version           = "v0.0.1"
	versionCheckURL   = "https://theerrorexe.github.io/api-patcher-ver.txt"
	wineInstallURL    = "https://wiki.winehq.org/Download"
	patcherDownloadURL = "https://github.com/AdmiralCurtiss/WfcPatcher/releases/download/v1.6/WfcPatcher1.6.zip"
)

func checkVersion() bool {
	resp, err := http.Get(versionCheckURL)
	if err != nil {
		fmt.Println("Error: Could not check version. Make sure you have an internet connection.")
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: Could not read version check response.")
		return false
	}

	if strings.Contains(string(body), version) {
		return true
	}
	fmt.Println("Error: Version not supported. Make sure you have the latest Version.")
	return false
}

func isWineInstalled() bool {
	_, err := exec.LookPath("wine")
	return err == nil
}

func ensurePatcherExists() error {
	patcherDir := "./patcher"
	helperPath := filepath.Join(patcherDir, "helper.exe")

	// Check if patcher directory and helper.exe exist
	if _, err := os.Stat(helperPath); err == nil {
		// File already exists
		return nil
	}

	fmt.Println("Patcher not found. Downloading and setting it up...")

	// Ensure the patcher directory exists
	err := os.MkdirAll(patcherDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating patcher directory: %v", err)
	}

	// Download WfcPatcher.zip
	zipPath := "./WfcPatcher1.6.zip"
	err = downloadFile(patcherDownloadURL, zipPath)
	if err != nil {
		return fmt.Errorf("error downloading patcher: %v", err)
	}
	defer os.Remove(zipPath) // Clean up the zip file after extraction

	// Extract only WfcPatcher.exe from the zip file
	err = extractFileFromZip(zipPath, "WfcPatcher.exe", helperPath)
	if err != nil {
		return fmt.Errorf("error extracting WfcPatcher.exe: %v", err)
	}

	fmt.Println("Patcher successfully set up.")
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractFileFromZip(zipPath, targetFile, outputPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == targetFile {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}

			return nil // Success
		}
	}

	return fmt.Errorf("file %s not found in zip archive", targetFile)
}

func listNDSFiles(directory string) ([]string, error) {
	var ndsFiles []string
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".nds") {
			ndsFiles = append(ndsFiles, file.Name())
		}
	}
	return ndsFiles, nil
}

func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func patchGame(ndsFile string) error {
	// Ensure tmp directory exists
	err := os.MkdirAll("./tmp", os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error creating tmp directory: %v", err)
	}

	// Copy the selected NDS file to ./tmp/game.nds
	tmpPath := "./tmp/game.nds"
	err = copyFile(ndsFile, tmpPath)
	if err != nil {
		return fmt.Errorf("Error copying file: %v", err)
	}

	// Run the external patcher
	cmd := exec.Command("wine", "./patcher/helper.exe", "-d", "d.errexe.xyz", tmpPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running patcher: %v", err)
	}

	// Check if patching was successful
	outputFile := fmt.Sprintf("game(d.errexe.xyz).nds")
	if _, err := os.Stat("./tmp/" + outputFile); err == nil {
		err = os.Rename("./tmp/"+outputFile, "./output.nds")
		if err != nil {
			return fmt.Errorf("Error moving patched file: %v", err)
		}
	} else if _, err := os.Stat("./" + outputFile); err == nil {
		err = os.Rename("./"+outputFile, "./output.nds")
		if err != nil {
			return fmt.Errorf("Error moving patched file: %v", err)
		}
	} else {
		return fmt.Errorf("Patching failed: Patched file not found")
	}

	// Success message
	fmt.Println("-------------------------------------")
	fmt.Println("ReviveMii Patcher Public Beta v0.0.1")
	fmt.Println("Game was patched successfully! It's \"output.nds\" now")
	fmt.Println("-------------------------------------")
	fmt.Println("Credits:")
	fmt.Println("helper.exe is https://github.com/AdmiralCurtiss/WfcPatcher")
	fmt.Println("helper.exe is licensed under the GNU General Public License v3.0 and because of this this Program is also licensed under the GNU General Public License v3.0.")
	fmt.Println("")
	fmt.Println("You can get this Program Source code on https://github.com/ReviveMii/Patcher")
	fmt.Println("")
	fmt.Println("(c) 2024. ReviveMii Project. https://revivemii.fr.to/")
	return nil
}

func main() {
	// Check if wine is installed
	if !isWineInstalled() {
		fmt.Println("Error: Wine is not installed. Please install Wine to run this program.")
		fmt.Printf("You can download and install Wine from: %s\n", wineInstallURL)
		os.Exit(1)
	}

	// Check version before proceeding
	if !checkVersion() {
		os.Exit(1)
	}

	// Ensure the patcher exists
	err := ensurePatcherExists()
	if err != nil {
		fmt.Printf("Error setting up patcher: %v\n", err)
		os.Exit(1)
	}

	// Get .nds files in the current directory
	ndsFiles, err := listNDSFiles(".")
	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
		os.Exit(1)
	}

	if len(ndsFiles) == 0 {
		fmt.Println("Nothing to patch.")
		return
	}

	// Display the list of .nds files and let user choose one
	fmt.Println("Select a file to patch:")
	for i, file := range ndsFiles {
		fmt.Printf("%d) %s\n", i+1, file)
	}

	var choice int
	fmt.Print("Enter your choice: ")
	_, err = fmt.Scan(&choice)
	if err != nil || choice < 1 || choice > len(ndsFiles) {
		fmt.Println("Invalid choice.")
		return
	}

	selectedFile := ndsFiles[choice-1]
	fmt.Printf("You selected: %s\n", selectedFile)

	// Patch the selected game
	err = patchGame(selectedFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
