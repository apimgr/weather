package services

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// TorKeyManager handles Tor hidden service key operations
type TorKeyManager struct {
	dataDir string
}

// NewTorKeyManager creates a new key manager
func NewTorKeyManager(dataDir string) *TorKeyManager {
	return &TorKeyManager{
		dataDir: dataDir,
	}
}

// ExportKeys exports the current Tor hidden service keys
func (km *TorKeyManager) ExportKeys() (publicKey, privateKey []byte, err error) {
	keysDir := filepath.Join(km.dataDir, "site")

	// Read private key
	privKeyPath := filepath.Join(keysDir, "hs_ed25519_secret_key")
	privKeyData, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Read public key
	pubKeyPath := filepath.Join(keysDir, "hs_ed25519_public_key")
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %w", err)
	}

	// Tor keys have a 32-byte header, actual key starts at byte 32
	if len(privKeyData) < 64 {
		return nil, nil, fmt.Errorf("invalid private key file: too short")
	}
	if len(pubKeyData) < 64 {
		return nil, nil, fmt.Errorf("invalid public key file: too short")
	}

	// Extract actual keys (skip 32-byte header)
	actualPrivKey := privKeyData[32:64]
	actualPubKey := pubKeyData[32:64]

	return actualPubKey, actualPrivKey, nil
}

// ImportKeys imports Tor hidden service keys from raw bytes
func (km *TorKeyManager) ImportKeys(publicKey, privateKey []byte) error {
	// Validate key lengths
	if len(privateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key length: got %d, want %d", len(privateKey), ed25519.PrivateKeySize)
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length: got %d, want %d", len(publicKey), ed25519.PublicKeySize)
	}

	// Verify key pair consistency
	pubFromPriv := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	for i := range publicKey {
		if publicKey[i] != pubFromPriv[i] {
			return fmt.Errorf("public key does not match private key")
		}
	}

	// Create keys directory
	keysDir := filepath.Join(km.dataDir, "site")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Write keys in Tor format (32-byte header + 32-byte key)
	if err := km.writePrivateKey(filepath.Join(keysDir, "hs_ed25519_secret_key"), privateKey); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	if err := km.writePublicKey(filepath.Join(keysDir, "hs_ed25519_public_key"), publicKey); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	// Write hostname file
	address := publicKeyToOnionAddress(publicKey)
	hostnameFile := filepath.Join(keysDir, "hostname")
	if err := os.WriteFile(hostnameFile, []byte(address+".onion\n"), 0600); err != nil {
		return fmt.Errorf("failed to write hostname file: %w", err)
	}

	return nil
}

// writePrivateKey writes a private key in Tor's format
func (km *TorKeyManager) writePrivateKey(path string, key []byte) error {
	// Tor private key format:
	// 32 bytes: "== ed25519v1-secret: type0 =="
	// 32 bytes: actual secret key
	header := []byte("== ed25519v1-secret: type0 ==")
	if len(header) < 32 {
		// Pad header to 32 bytes
		padding := make([]byte, 32-len(header))
		header = append(header, padding...)
	}

	data := make([]byte, 0, 64)
	data = append(data, header[:32]...)
	data = append(data, key...)

	return os.WriteFile(path, data, 0600)
}

// writePublicKey writes a public key in Tor's format
func (km *TorKeyManager) writePublicKey(path string, key []byte) error {
	// Tor public key format:
	// 32 bytes: "== ed25519v1-public: type0 =="
	// 32 bytes: actual public key
	header := []byte("== ed25519v1-public: type0 ==")
	if len(header) < 32 {
		// Pad header to 32 bytes
		padding := make([]byte, 32-len(header))
		header = append(header, padding...)
	}

	data := make([]byte, 0, 64)
	data = append(data, header[:32]...)
	data = append(data, key...)

	return os.WriteFile(path, data, 0600)
}

// ImportFromFile imports keys from a Tor key file
func (km *TorKeyManager) ImportFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if it's a PEM file
	if block, _ := pem.Decode(data); block != nil {
		// Handle PEM format
		return fmt.Errorf("PEM format not yet supported")
	}

	// Check if it's a raw Tor key file (64 bytes)
	if len(data) == 64 {
		// Extract key (skip 32-byte header)
		key := data[32:64]

		// Determine if private or public based on header
		header := string(data[:28])
		if header == "== ed25519v1-secret: type0" {
			// Private key - derive public key
			privKey := ed25519.PrivateKey(key)
			pubKey := privKey.Public().(ed25519.PublicKey)
			return km.ImportKeys(pubKey, key)
		} else if header == "== ed25519v1-public: type0 " {
			return fmt.Errorf("cannot import public key alone, need private key")
		}
	}

	// Try base64 decode
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err == nil && len(decoded) == ed25519.PrivateKeySize {
		privKey := ed25519.PrivateKey(decoded)
		pubKey := privKey.Public().(ed25519.PublicKey)
		return km.ImportKeys(pubKey, decoded)
	}

	return fmt.Errorf("unsupported key file format")
}

// GetCurrentAddress returns the current .onion address
func (km *TorKeyManager) GetCurrentAddress() (string, error) {
	hostnameFile := filepath.Join(km.dataDir, "site", "hostname")
	data, err := os.ReadFile(hostnameFile)
	if err != nil {
		return "", fmt.Errorf("failed to read hostname file: %w", err)
	}

	return string(data), nil
}

// DeleteKeys removes all Tor hidden service keys
func (km *TorKeyManager) DeleteKeys() error {
	keysDir := filepath.Join(km.dataDir, "site")
	return os.RemoveAll(keysDir)
}

// KeysExist checks if Tor keys already exist
func (km *TorKeyManager) KeysExist() bool {
	privKeyPath := filepath.Join(km.dataDir, "site", "hs_ed25519_secret_key")
	pubKeyPath := filepath.Join(km.dataDir, "site", "hs_ed25519_public_key")

	_, err1 := os.Stat(privKeyPath)
	_, err2 := os.Stat(pubKeyPath)

	return err1 == nil && err2 == nil
}
