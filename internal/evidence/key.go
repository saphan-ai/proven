package evidence

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"
)

// The settings that name the organisation's keys. Each holds a file path, never a key.
const (
	// EnvPackKey names the PEM file of the Ed25519 private key that signs a pack.
	EnvPackKey = "PROVEN_PACK_KEY"
	// EnvPackPublicKey names the PEM file of the Ed25519 public key that verifies a pack.
	EnvPackPublicKey = "PROVEN_PACK_PUBLIC_KEY"
)

// KeyError names a key setting that cannot be used and why. It carries the
// setting and the file path, never a byte of the key (ADR-0005).
type KeyError struct {
	Setting string
	Path    string
	Reason  string
}

func (e *KeyError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("evidence pack key refused: %s: %s", e.Setting, e.Reason)
	}
	return fmt.Sprintf("evidence pack key refused: %s=%q: %s", e.Setting, e.Path, e.Reason)
}

// LoadPrivateKey reads a PKCS #8 PEM Ed25519 private key from the file the
// setting names. An unset setting, an unreadable file, a file that group or
// others can reach (any bit of mode 0077) and a file that is not such a key
// are each named.
func LoadPrivateKey(setting, path string) (ed25519.PrivateKey, error) {
	if err := ownerOnly(setting, path); err != nil {
		return nil, err
	}
	der, err := readPEM(setting, path, "PRIVATE KEY")
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, &KeyError{Setting: setting, Path: path, Reason: "not a PKCS #8 private key"}
	}
	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, &KeyError{Setting: setting, Path: path, Reason: "not an Ed25519 key"}
	}
	return key, nil
}

// ownerOnly refuses a private key file that group or others can reach. A file
// that cannot be examined is left to readPEM, which names it.
func ownerOnly(setting, path string) error {
	if path == "" || runtime.GOOS == "windows" {
		return nil
	}
	st, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if mode := st.Mode().Perm(); mode&0o077 != 0 {
		return &KeyError{Setting: setting, Path: path, Reason: fmt.Sprintf("the file has mode %04o, and group or others can reach it; run chmod 600 on it", mode)}
	}
	return nil
}

// LoadPublicKey reads a PKIX PEM Ed25519 public key from the file the setting names.
func LoadPublicKey(setting, path string) (ed25519.PublicKey, error) {
	der, err := readPEM(setting, path, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, &KeyError{Setting: setting, Path: path, Reason: "not a PKIX public key"}
	}
	key, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, &KeyError{Setting: setting, Path: path, Reason: "not an Ed25519 key"}
	}
	return key, nil
}

// readPEM returns the DER bytes of the first PEM block of the given type.
func readPEM(setting, path, blockType string) ([]byte, error) {
	if path == "" {
		return nil, &KeyError{Setting: setting, Reason: "not set"}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		// The file system's sentence names the path and the cause, not the content.
		return nil, &KeyError{Setting: setting, Path: path, Reason: "cannot be read"}
	}
	block, _ := pem.Decode(body)
	if block == nil || block.Type != blockType {
		return nil, &KeyError{Setting: setting, Path: path, Reason: fmt.Sprintf("holds no %s PEM block", blockType)}
	}
	return block.Bytes, nil
}
