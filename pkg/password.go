package pkg

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19 * 1024 // 19 MiB em KiB
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		saltEncoded,
		hashEncoded,
	)

	return encodedHash, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[4] == "" || parts[5] == "" {
		return false, fmt.Errorf("parse password hash: invalid argon2 format")
	}

	version, err := parseArgon2Parameter(parts[2], "v", 32)
	if err != nil {
		return false, fmt.Errorf("parse password hash version: %w", err)
	}

	parameters := strings.Split(parts[3], ",")
	if len(parameters) != 3 {
		return false, fmt.Errorf("parse password hash: invalid parameters")
	}

	memory, err := parseArgon2Parameter(parameters[0], "m", 32)
	if err != nil {
		return false, fmt.Errorf("parse password hash memory: %w", err)
	}

	argonTime, err := parseArgon2Parameter(parameters[1], "t", 32)
	if err != nil {
		return false, fmt.Errorf("parse password hash time: %w", err)
	}

	threads, err := parseArgon2Parameter(parameters[2], "p", 8)
	if err != nil {
		return false, fmt.Errorf("parse password hash threads: %w", err)
	}

	if version != uint64(argon2.Version) {
		return false, fmt.Errorf("unsupported argon2 version")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(argonTime),
		uint32(memory),
		uint8(threads),
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

func parseArgon2Parameter(raw, name string, bitSize int) (uint64, error) {
	key, value, ok := strings.Cut(raw, "=")
	if !ok || key != name || value == "" {
		return 0, fmt.Errorf("invalid %s parameter", name)
	}

	parsed, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %w", name, err)
	}

	return parsed, nil
}
