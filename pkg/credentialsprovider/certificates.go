package credentialsprovider

import (
	"fmt"
	"os"
	"path/filepath"
)

// Credentials contains the pem encoded certificate, private key, and CA.
type Credentials struct {
	certificate []byte
	privatekey  []byte
	ca          []byte
}

// WriteToFileSystem writes the credentials to the given path in the file system. The path is created if it does not exist.
// The following files are created:
//
// - private.key
// - certificate.crt
// - ca.crt
func (c *Credentials) WriteToFileSystem(rootDir string) error {

	// Create the directory if it does not exist
	err := os.MkdirAll(rootDir, 0700)
	if err != nil {
		return err
	}

	// Write the private key
	privateKeyPath := filepath.Join(rootDir, "private.key")
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("error creating private key file: %v", err)
	}

	_, err = privateKeyFile.Write(c.privatekey)
	if err != nil {
		return fmt.Errorf("error writing private key file: %v", err)
	}

	// Write the certificate
	certificatePath := filepath.Join(rootDir, "certificate.crt")
	certificateFile, err := os.Create(certificatePath)
	if err != nil {
		return fmt.Errorf("error creating certificate file: %v", err)
	}

	_, err = certificateFile.Write(c.certificate)
	if err != nil {
		return fmt.Errorf("error writing certificate file: %v", err)
	}

	// Write the CA
	caPath := filepath.Join(rootDir, "ca.crt")
	caFile, err := os.Create(caPath)
	if err != nil {
		return fmt.Errorf("error creating CA file: %v", err)
	}

	_, err = caFile.Write(c.ca)
	if err != nil {
		return fmt.Errorf("error writing CA file: %v", err)
	}

	// Write the certificate chain
	chainPath := filepath.Join(rootDir, "chain.crt")
	chainFile, err := os.Create(chainPath)
	if err != nil {
		return fmt.Errorf("error creating chain file: %v", err)
	}

	newLine := []byte("\n")

	_, err = chainFile.Write(append(c.certificate, append(newLine, c.ca...)...))
	if err != nil {
		return fmt.Errorf("error writing chain file: %v", err)
	}

	return nil
}
