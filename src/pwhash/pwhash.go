package pwhash

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

var Iterations uint32
var Threads    uint8
var Memory_KiB uint32
var KeyLen     uint32

func ValidateSettings() (err error) {

	mem_ok     := Memory_KiB >= uint32(Threads)*8
	keylen_ok  := KeyLen >= 4
	threads_ok := Threads >= 1
	if !mem_ok     { errors.Join(err, errors.New(fmt.Sprintf("Password hashing algorithm Argon2 requires that the given memory be at least %vKiB (8 times the number of threads, which is set to %v)", Threads*8, Threads))) }
	if !keylen_ok  { errors.Join(err, errors.New("Password hashing algorithm Argon2 requires that the key length be at least 4")) }
	if !threads_ok { errors.Join(err, errors.New("Password hashing algorithm Argon2 needs at least 1 thread to run")) }

	return
}

// package docs: https://pkg.go.dev/golang.org/x/crypto/argon2?utm_source=godoc#hdr-Argon2id
// How to choose argon2 parameters: https://argon2-cffi.readthedocs.io/en/stable/parameters.html
// site for experimenting with argon2: https://argon2.online/
func EncodePassword(password string) (string, error) {
	iterations := Iterations
	threads    := Threads
	memory     := Memory_KiB
	key_len    := KeyLen

	salt := make([]byte, 16) // must be at least 8 according to argon2 paper
	_, err := crypto_rand.Read(salt)
	if err != nil { return "", err }

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, key_len)

	// log.Println("original hash:   ", hash)
	// log.Println("original salt:   ", salt)
	// log.Println("original memory: ", memory)
	// log.Println("original threads:", threads)
	// log.Println("original key_len:", key_len)

	return encode_hash(hash, salt, iterations, memory, threads, key_len), nil
}

func VerifyPassword(password string, encoded_hash string) (match bool) {
	target_hash, salt, iterations, memory, threads, err := decode_hash(encoded_hash)
	if err != nil { log.Print(err); return false }

	// log.Println("derived hash:   ", hash)
	// log.Println("derived salt:   ", salt)
	// log.Println("derived memory: ", memory)
	// log.Println("derived threads:", threads)
	// log.Println("derived key_len:", len(hash))

	password_hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(target_hash)))
	return bytes.Equal(password_hash, target_hash)
}

func ShowcaseHashSpeed() {
	test_count := 5
	var (avg1, avg2 time.Duration)

	for i := 0; i < test_count; i++ {
		func (pw string, iteration int) {
			start1 := time.Now()
			hash, _ := EncodePassword(pw)
			time_took1 := time.Since(start1)
			avg1 += time_took1

			start2 := time.Now()
			match := VerifyPassword(pw, hash)
			time_took2 := time.Since(start2)
			avg2 += time_took2

			log.Printf("[%v] generating password with current config: %v", iteration, time_took1)
			log.Printf("[%v] verifying  password with current config: %v", iteration, time_took2)

			if !match {
				log.Fatal("There's a problem with the password hashing, the testrun generated a hash but failed to match that hash against the same password.\nGenerated hash: %v\nTest-Password used: %v", hash, pw)
			}

		} ("Hello, Sailor", i)
	}

	log.Printf("avg to generate: %vms", int(avg1.Milliseconds()) / test_count)
	log.Printf("avg to compare:  %vms", int(avg2.Milliseconds()) / test_count)
}


func encode_hash(hash []byte, salt []byte, iterations uint32, memory uint32, threads uint8, key_len uint32) string {
	return fmt.Sprintf("$%v$v=%v$m=%v,t=%v,p=%v$%v$%v",
		"argon2id",
		argon2.Version,
		memory,
		iterations,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
}

func decode_hash(encoded_hash string) (hash []byte, salt []byte, iterations uint32, memory uint32, threads uint8, err error) {

	// Example hash: $argon2id$v=19$m=16384,t=1,p=4$5uj5oafRu1CoRecpBo/WRg$eILSo35LhFTZsWt4PgFYZCi4NwQNorBvBWhphSeYZWU
	// generated with string "whatever, man"
	// $ algorithm $ v=version $ m=memory_in_KiB, t=iterations, p=thread_count $ base64salt $ base64hash
	// $ v=version is optional
	// hash and salt are without the base64 padding characters '=' at the end

	if encoded_hash[0] != '$' { err = errors.New("encoded hash doesn't start with '$'"); return }
	str := encoded_hash[1:]

	algorithm, str := split_on_byte(str, '$')
	if algorithm != "argon2id" { err = errors.New("encoded hash doesn't use argon2id"); return }

	settings, str := split_on_byte(str, '$')
	if strings.HasPrefix(settings, "v=") { // it wasn't the settings, but the optional version section
		var version int64; version, err = strconv.ParseInt(settings[2:], 10, 64)
		if err != nil { return }
		if version != argon2.Version { err = errors.New(fmt.Sprintf("hash was encoded with v=%v, but current library works with v=%v", version, argon2.Version)); return }
		settings, str = split_on_byte(str, '$')
	}

	salt_str, str := split_on_byte(str, '$')
	salt, err = base64.RawStdEncoding.DecodeString(salt_str)
	if err != nil { return }
	hash, err = base64.RawStdEncoding.DecodeString(str) // hash is the last part, so whatever is left of `str` is the hash
	if err != nil { return }

	var m, t, p bool
	for {
		// I'm of the opinion that the settings should be allowed in any order, $m,t,p$, $t,p,m$, etc.
		// Therefore this code doesn't assume the order. Most other generators do, but whatever.
		var key_eq_val string
		key_eq_val, settings = split_on_byte(settings, ',')
		if key_eq_val == "" { break }
		key, val := split_on_byte(key_eq_val, '=')
		var tmp int64
		switch key {
			case "m":
				if m { err = errors.New("duplicate m="); return }; m = true
				tmp, err = strconv.ParseInt(val, 10, 32)
				if err != nil { return }
				memory = uint32(tmp)

			case "t":
				if t { err = errors.New("duplicate t="); return }; t = true
				tmp, err = strconv.ParseInt(val, 10, 32)
				if err != nil { return }
				iterations = uint32(tmp)

			case "p":
				if p { err = errors.New("duplicate p="); return }; p = true
				tmp, err = strconv.ParseInt(val, 10, 8)
				if err != nil { return }
				threads = uint8(tmp)

			default:
				err = errors.New(fmt.Sprintf("unexpected setting %v=", key))
				return
		}
	}

	if !m || !t || !p {
		err = errors.New(fmt.Sprintf("missing settings. m = %v, t = %v, p = %v", present(m), present(t), present(p)))
		return
	}

	return
}

func present(b bool) string {
	if b { return "present" } else { return "not present" }
}

func split_on_byte(str string, separator byte) (string, string) {
	idx := strings.IndexByte(str, separator)
	if idx != -1 {
		return str[:idx], str[idx+1:]
	} else {
		return str, ""
	}
}