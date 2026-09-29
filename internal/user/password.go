package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// OWASP Argon2id parameters: 19 MiB, 2 iterations, 1 lane.
const (
	argonMemory      = 19 * 1024
	argonIterations  = 2
	argonParallelism = 1
	argonSaltLength  = 16
	argonKeyLength   = 32
)

// SetPassword hashes plaintext with Argon2id and stores the PHC encoded hash.
func (u *User) SetPassword(plaintext string) error {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(plaintext), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	u.PasswordHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return nil
}

// PasswordMatches compares plaintext with the stored Argon2id hash.
func (u *User) PasswordMatches(plaintext string) (bool, error) {
	salt, hash, memory, iterations, parallelism, err := decodeArgon2ID(u.PasswordHash)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey([]byte(plaintext), salt, iterations, memory, parallelism, uint32(len(hash)))
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeArgon2ID(encoded string) (salt, hash []byte, memory, iterations uint32, parallelism uint8, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, 0, 0, 0, fmt.Errorf("unsupported argon2id version %d", version)
	}

	var parsedParallelism uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parsedParallelism); err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id parameters: %w", err)
	}
	if parsedParallelism == 0 || parsedParallelism > 255 {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id parallelism")
	}
	parallelism = uint8(parsedParallelism)

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id salt: %w", err)
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, 0, 0, 0, fmt.Errorf("invalid argon2id hash: %w", err)
	}
	return salt, hash, memory, iterations, parallelism, nil
}
