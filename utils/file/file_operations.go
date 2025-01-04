package file

import (
	"archive/tar"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pierrec/lz4/v4"
	"github.com/pterm/pterm"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

// Helper function to decompress LZ4 and extract tar
func DecompressAndExtractLz4Tar(lz4Path, destDir string) error {
	// Open the LZ4 file
	file, err := os.Open(lz4Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create LZ4 reader
	lz4Reader := lz4.NewReader(file)

	// Create tar reader
	tarReader := tar.NewReader(lz4Reader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return err
		}

		// Determine proper file path
		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create file
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			// Set permissions
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			// Unsupported type
			pterm.Warning.Printf(fmt.Sprintf("Unsupported file type: %v in %s", header.Typeflag, header.Name))
		}
	}

	return nil
}

// Helper function to download a file with a progress bar
func DownloadFileWithProgress(url, dest string) error {
	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get the size
	sizeStr := resp.Header.Get("Content-Length")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return errors.New("invalid Content-Length")
	}

	// Create progress bar
	p := mpb.New(
		mpb.WithWidth(64),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
	bar := p.AddBar(int64(size),
		mpb.PrependDecorators(
			decor.Name("Downloading:", decor.WC{W: len("Downloading: "), C: decor.DidentRight}),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 5}),
		),
	)

	// Create a proxy reader
	reader := bar.ProxyReader(resp.Body)
	defer reader.Close()

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	p.Wait()

	return nil
}

func DownloadFileWithAria2(url, dest string) error {
	// Check if aria2c is installed
	_, err := exec.LookPath("aria2c")
	if err != nil {
		pterm.Warning.Println("aria2c is not installed.")
		pterm.Info.Println("Attempting to install aria2...")

		// Attempt to install aria2
		installCmd := exec.Command("sudo", "apt-get", "install", "aria2", "-y")
		var stderr bytes.Buffer
		installCmd.Stderr = &stderr

		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install aria2: %v\nDetails: %s", err, stderr.String())
		}
	}

	// Prepare aria2c command
	cmdArgs := []string{
		"--split=16",
		"--max-connection-per-server=16",
		"--min-split-size=1M",
		"--enable-color=true",       // Disable colors for cleaner output
		"--console-log-level=error", // Suppress logs except errors
		"--summary-interval=1",      // Update progress every second
		"--dir=" + filepath.Dir(dest),
		"--out=" + filepath.Base(dest),
		url,
	}

	cmd := exec.Command("aria2c", cmdArgs...)

	// Capture stdout and filter progress lines
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stdout: %v", err)
	}

	cmd.Stderr = os.Stderr // Show errors if any

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start aria2c: %v", err)
	}

	// Filter progress lines and clean output
	scanner := bufio.NewScanner(stdout)
	lastLine := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[#") { // Filter only progress lines
			// Clean and update the line
			lastLine = strings.TrimSpace(line)
			fmt.Printf("\r%-80s", lastLine) // Clear previous line and overwrite
		}
	}

	// Print final progress summary after completion
	fmt.Printf("\r%-80s\n", "") // Clear line and show completion message
	pterm.Info.Println("Download complete!")

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("aria2c failed: %v", err)
	}

	return nil
}
