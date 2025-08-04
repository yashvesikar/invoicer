package backup

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/invoicer/config"
)

type Metadata struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
}

func CreateBackup(outputPath string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no configuration found")
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("invoicer_backup_%s.tar.gz", timestamp)
	
	if outputPath != "" {
		backupName = filepath.Join(outputPath, backupName)
	}

	file, err := os.Create(backupName)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	hostname, _ := os.Hostname()
	metadata := Metadata{
		Version:   "1.0",
		Timestamp: time.Now(),
		Hostname:  hostname,
	}
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := addToTar(tarWriter, "metadata.json", metadataBytes); err != nil {
		return fmt.Errorf("failed to add metadata: %w", err)
	}

	dataFiles := []string{"clients.json", "invoices.json", "audit.json"}
	for _, fileName := range dataFiles {
		filePath := filepath.Join(cfg.DataDir(), fileName)
		if _, err := os.Stat(filePath); err == nil {
			if err := addFileToTar(tarWriter, filePath, filepath.Join("data", fileName)); err != nil {
				return fmt.Errorf("failed to add %s: %w", fileName, err)
			}
		}
	}

	configPath, err := config.ConfigPath()
	if err == nil {
		if _, err := os.Stat(configPath); err == nil {
			if err := addFileToTar(tarWriter, configPath, "config/config.json"); err != nil {
				return fmt.Errorf("failed to add config: %w", err)
			}
		}
	}

	templatePath := filepath.Join(cfg.TemplatesDir(), "invoice.tex")
	if _, err := os.Stat(templatePath); err == nil {
		if err := addFileToTar(tarWriter, templatePath, "templates/invoice.tex"); err != nil {
			return fmt.Errorf("failed to add template: %w", err)
		}
	}

	fmt.Printf("Backup created successfully: %s\n", backupName)
	return nil
}

func RestoreBackup(backupPath string) error {
	if err := ValidateBackup(backupPath); err != nil {
		return fmt.Errorf("backup validation failed: %w", err)
	}

	fmt.Print("This will overwrite existing data. Continue? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return fmt.Errorf("restore cancelled by user")
	}

	fmt.Println("Creating backup of current data before restore...")
	tempBackupPath := filepath.Join(os.TempDir(), fmt.Sprintf("pre_restore_backup_%s", time.Now().Format("20060102_150405")))
	if err := CreateBackup(tempBackupPath); err != nil {
		fmt.Printf("Warning: Could not create pre-restore backup: %v\n", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		var destPath string
		switch {
		case strings.HasPrefix(header.Name, "data/"):
			destPath = filepath.Join(cfg.DataDir(), strings.TrimPrefix(header.Name, "data/"))
		case header.Name == "config/config.json":
			configPath, err := config.ConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}
			destPath = configPath
		case strings.HasPrefix(header.Name, "templates/"):
			destPath = filepath.Join(cfg.TemplatesDir(), strings.TrimPrefix(header.Name, "templates/"))
		case header.Name == "metadata.json":
			continue
		default:
			fmt.Printf("Skipping unknown file: %s\n", header.Name)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", destPath, err)
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
		outFile.Close()

		fmt.Printf("Restored: %s\n", header.Name)
	}

	fmt.Println("Restore completed successfully!")
	return nil
}

func ValidateBackup(backupPath string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("invalid backup file format: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var hasMetadata bool
	var metadata Metadata

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("corrupted backup file: %w", err)
		}

		if header.Name == "metadata.json" {
			hasMetadata = true
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return fmt.Errorf("failed to read metadata: %w", err)
			}
			if err := json.Unmarshal(data, &metadata); err != nil {
				return fmt.Errorf("invalid metadata: %w", err)
			}
		}
	}

	if !hasMetadata {
		return fmt.Errorf("backup file missing metadata")
	}

	if metadata.Version != "1.0" {
		return fmt.Errorf("unsupported backup version: %s", metadata.Version)
	}

	return nil
}

func addToTar(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(data); err != nil {
		return err
	}

	return nil
}

func addFileToTar(tw *tar.Writer, sourcePath, tarPath string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    tarPath,
		Mode:    0644,
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}