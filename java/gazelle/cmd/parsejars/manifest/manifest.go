// Package manifest provide the data structure to pre-compute the Maven JAR mapping.
//
// Inspired from https://github.com/bazelbuild/rules_python/blob/27d0c7bb8e663dd2e2e9b295ecbfed680e641dfd/gazelle/manifest/manifest.go#L15
package manifest

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type File struct {
	Integrity string    `json:"integrity"`
	Manifest  *Manifest `json:"manifest,omitempty"`
}

func NewFile(manifest *Manifest) *File {
	return &File{Manifest: manifest}
}

// Encode encodes the manifest file to the given writer.
func (f *File) Encode(w io.Writer, mavenInstallPath string) error {
	mavenInstallChecksum, err := sha256File(mavenInstallPath)
	if err != nil {
		return fmt.Errorf("failed to encode manifest file: %w", err)
	}
	integrityBytes, err := f.calculateIntegrity(mavenInstallChecksum)
	if err != nil {
		return fmt.Errorf("failed to encode manifest file: %w", err)
	}
	f.Integrity = fmt.Sprintf("%x", integrityBytes)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(f); err != nil {
		return fmt.Errorf("failed to encode manifest file: %w", err)
	}
	return nil
}

// VerifyIntegrity verifies if the integrity set in the File is valid.
func (f *File) VerifyIntegrity(mavenInstallPath string) (bool, error) {
	mavenInstallChecksum, err := sha256File(mavenInstallPath)
	if err != nil {
		return false, fmt.Errorf("failed to verify integrity: %w", err)
	}
	integrityBytes, err := f.calculateIntegrity(mavenInstallChecksum)
	if err != nil {
		return false, fmt.Errorf("failed to verify integrity: %w", err)
	}
	valid := (f.Integrity == fmt.Sprintf("%x", integrityBytes))
	return valid, nil
}

func (f *File) calculateIntegrity(mavenInstallChecksum []byte) ([]byte, error) {
	hash := sha256.New()
	if err := json.NewEncoder(hash).Encode(f.Manifest); err != nil {
		return nil, fmt.Errorf("failed to calculate integrity: %w", err)
	}
	if _, err := hash.Write(mavenInstallChecksum); err != nil {
		return nil, fmt.Errorf("failed to calculate integrity: %w", err)
	}
	return hash.Sum(nil), nil
}

// Decode decodes the manifest file from the given path.
func (f *File) Decode(manifestPath string) error {
	file, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to decode manifest file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(f); err != nil {
		return fmt.Errorf("failed to decode manifest file: %w", err)
	}

	return nil
}

type Manifest struct {
	ArtifactsMapping map[string][]string `json:"artifacts_mapping"`
}

func sha256File(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sha256 sum for file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate sha256 sum for file: %w", err)
	}

	return hash.Sum(nil), nil
}
