package plugingetter

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// A ChecksumError is returned when a checksum differs
type ChecksumError struct {
	Hash     hash.Hash
	Actual   []byte
	Expected []byte
	File     string
}

func (cerr *ChecksumError) Error() string {
	if cerr == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"Checksums did not match for %s.\nExpected: %s\nGot: %s\n%T",
		cerr.File,
		hex.EncodeToString(cerr.Expected),
		hex.EncodeToString(cerr.Actual),
		cerr.Hash, // ex: *sha256.digest
	)
}

type Checksum struct {
	Expected []byte
	Checksummer
}

type Checksummer struct {
	// Something like md5 or sha256
	Type string
	// Hash function
	hash.Hash
}

func (c *Checksummer) FileExt() string {
	return "_" + strings.ToUpper(c.Type) + "SUM"
}

// GetChecksumOfFile will extract the checksum from file `filePath + c.FileExt()`.
// It expects the checksum file to only contains the checksum and nothing else.
func (c *Checksummer) GetChecksumOfFile(filePath string) ([]byte, error) {
	checksumFile := filePath + c.FileExt()

	f, err := os.Open(checksumFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return c.ParseChecksum(f)
}

// ParseChecksum expects the checksum reader to only contains the checksum and
// nothing else.
func (c *Checksummer) ParseChecksum(f io.Reader) ([]byte, error) {
	res := make([]byte, c.Hash.Size())
	_, err := hex.NewDecoder(f).Read(res)
	return res, err
}

// ChecksumFile compares the expected checksum to the checksum of the file in
// filePath using the hash function.
func (c *Checksummer) ChecksumFile(expected []byte, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Checksum: failed to open file for checksum: %s", err)
	}
	defer f.Close()

	c.Hash.Reset()
	if _, err := io.Copy(c.Hash, f); err != nil {
		return fmt.Errorf("Failed to hash: %s", err)
	}

	if actual := c.Hash.Sum(nil); !bytes.Equal(actual, expected) {
		return &ChecksumError{
			Hash:     c.Hash,
			Actual:   actual,
			Expected: expected,
			File:     filePath,
		}
	}

	return nil
}
