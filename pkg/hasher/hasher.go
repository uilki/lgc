package hasher

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
)

const (
	HashSHA256 uint8 = iota
	HashSHA512
)

// Hasher is an interface used to hash and verify passwords using
// the using different hash algorithms.
type Hasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

type hasher uint8

// Error
var (
	ErrEmptyPassword       = errors.New("empty password provided")
	ErrUnsupportedHashType = errors.New("hash: unsupported hash key")
)

var defaultHasher hasher

// init the default hasher.
func init() {
	defaultHasher = hasher(HashSHA256)
}

// New returns a new Hasher, configured with hash type.
// A non-nil error will be returned if hash type is invalid.
func New(t uint8) (Hasher, error) {
	switch t {
	case HashSHA256, HashSHA512:
		h := hasher(t)
		return &h, nil
	default:
		return nil, ErrUnsupportedHashType
	}

}

// HashPassword hashes the given password data using the default hasher.
// If empty password provided returns ErrEmptyPassword
func HashPassword(password string) (string, error) {
	return defaultHasher.HashPassword(password)
}

// HashPassword hashes the given password data based on hasher hash type.
// If empty password provided returns ErrEmptyPassword
func (h *hasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hash, err := hashType(*h)
	if err != nil {
		return "", err
	}

	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Returns a hash the given type.
// Rerturn not nil error if type is not supported
func hashType(t hasher) (hash.Hash, error) {
	switch uint8(t) {
	case HashSHA256:
		return sha256.New(), nil
	case HashSHA512:
		return sha512.New(), nil
	default:
		return nil, ErrUnsupportedHashType
	}
}

// CheckPasswordHash attempts to verifiy the password using the default hasher.
func CheckPasswordHash(password, hash string) bool {
	return defaultHasher.CheckPasswordHash(password, hash)
}

// CheckPasswordHash hashed the given password and compares it to the given hash data,
// returning a flag which determines whether or not the password matches the hash.
func (h *hasher) CheckPasswordHash(password, hash string) bool {
	newHash, err := h.HashPassword(password)

	if err != nil {
		return false
	}

	return newHash == hash
}
