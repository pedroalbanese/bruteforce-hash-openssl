package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/sha3"
	"golang.org/x/crypto/ripemd160"

	"github.com/emmansun/gmsm/sm3"
)

const (
	Version     = "0.0.0"
	VersionDate = "02 Sep 2026"
	AppName     = "Bruteforce Hash OpenSSL"
)

type Config struct {
	Target     string
	MinLen     int
	MaxLen     int
	Prefix     string
	Suffix     string
	Charset    string
	Threads    int
	Verbose    int
	HashType   string
	OutputFile string
	InputFile  string
	XofLen     int // tamanho do SHAKE em bytes
}

type HashInfo struct {
	Name      string
	Length    int
	XofLen    int // tamanho padrão para SHAKE em bytes
	New       func() hash.Hash
	IsShake   bool
	HasXofLen bool // indica se suporta xoflen
}

var hashRegistry = map[string]HashInfo{
	// MD5
	"md5": {
		Name:   "MD5",
		Length: 32,
		New:    md5.New,
	},
	
	// MD4
	"md4": {
		Name:   "MD4",
		Length: 32,
		New:    md4.New,
	},
	
	// SHA1
	"sha1": {
		Name:   "SHA1",
		Length: 40,
		New:    sha1.New,
	},
	
	// RIPEMD-160
	"rmd160": {
		Name:   "RIPEMD-160",
		Length: 40,
		New:    ripemd160.New,
	},
	
	// SHA2
	"sha224": {
		Name:   "SHA224",
		Length: 56,
		New:    sha256.New224,
	},
	"sha256": {
		Name:   "SHA256",
		Length: 64,
		New:    sha256.New,
	},
	"sha384": {
		Name:   "SHA384",
		Length: 96,
		New:    sha512.New384,
	},
	"sha512": {
		Name:   "SHA512",
		Length: 128,
		New:    sha512.New,
	},
	"sha512-224": {
		Name:   "SHA512-224",
		Length: 56,
		New:    sha512.New512_224,
	},
	"sha512-256": {
		Name:   "SHA512-256",
		Length: 64,
		New:    sha512.New512_256,
	},
	
	// SHA3
	"sha3-224": {
		Name:   "SHA3-224",
		Length: 56,
		New:    sha3.New224,
	},
	"sha3-256": {
		Name:   "SHA3-256",
		Length: 64,
		New:    sha3.New256,
	},
	"sha3-384": {
		Name:   "SHA3-384",
		Length: 96,
		New:    sha3.New384,
	},
	"sha3-512": {
		Name:   "SHA3-512",
		Length: 128,
		New:    sha3.New512,
	},
	
	// SHAKE (OpenSSL 3.4.0+ usa -xoflen)
	"shake128": {
		Name:      "SHAKE128",
		Length:    32,  // hash em hex (16 bytes * 2)
		XofLen:    16,  // bytes padrão (128 bits)
		New:       func() hash.Hash { return sha3.NewShake128() },
		IsShake:   true,
		HasXofLen: true,
	},
	"shake256": {
		Name:      "SHAKE256",
		Length:    64,  // hash em hex (32 bytes * 2)
		XofLen:    32,  // bytes padrão (256 bits)
		New:       func() hash.Hash { return sha3.NewShake256() },
		IsShake:   true,
		HasXofLen: true,
	},
	
	// SM3
	"sm3": {
		Name:   "SM3",
		Length: 64,
		New:    sm3.New,
	},
	
	// BLAKE2
	"blake2b512": {
		Name:   "BLAKE2b-512",
		Length: 128,
		New:    func() hash.Hash { h, _ := blake2b.New(64, nil); return h },
	},
	"blake2s256": {
		Name:   "BLAKE2s-256",
		Length: 64,
		New:    func() hash.Hash { h, _ := blake2s.New256(nil); return h },
	},
}

type StringGenerator struct {
	charset  []rune
	prefix   string
	suffix   string
	minLen   int
	maxLen   int
	current  []int
	firstRun bool
}

func NewStringGenerator(config *Config) *StringGenerator {
	charset := []rune(config.Charset)
	if len(charset) == 0 {
		charset = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@#$%^&*()-_=+[]{}|;:,.<>?")
	}

	middleLen := config.MinLen - len(config.Prefix) - len(config.Suffix)
	if middleLen < 0 {
		middleLen = 0
	}

	return &StringGenerator{
		charset:  charset,
		prefix:   config.Prefix,
		suffix:   config.Suffix,
		minLen:   config.MinLen,
		maxLen:   config.MaxLen,
		current:  make([]int, middleLen),
		firstRun: true,
	}
}

func (sg *StringGenerator) Next() (string, bool) {
	if sg.firstRun {
		sg.firstRun = false
		for i := range sg.current {
			sg.current[i] = 0
		}
		return sg.buildString(), true
	}

	for i := len(sg.current) - 1; i >= 0; i-- {
		sg.current[i]++
		if sg.current[i] < len(sg.charset) {
			return sg.buildString(), true
		}
		sg.current[i] = 0

		if i == 0 {
			currentTotalLen := len(sg.prefix) + len(sg.suffix) + len(sg.current)
			if currentTotalLen < sg.maxLen {
				sg.current = append(sg.current, 0)
				return sg.buildString(), true
			}
			return "", false
		}
	}
	return "", false
}

func (sg *StringGenerator) buildString() string {
	middle := make([]rune, len(sg.current))
	for i, idx := range sg.current {
		middle[i] = sg.charset[idx]
	}
	return sg.prefix + string(middle) + sg.suffix
}

func (sg *StringGenerator) GetTotalCount() uint64 {
	var total uint64 = 0
	base := uint64(len(sg.charset))
	prefixLen := len(sg.prefix)
	suffixLen := len(sg.suffix)

	for length := sg.minLen; length <= sg.maxLen; length++ {
		middleLen := length - prefixLen - suffixLen
		if middleLen < 0 {
			continue
		}
		var combinations uint64 = 1
		for i := 0; i < middleLen; i++ {
			combinations *= base
		}
		total += combinations
	}
	return total
}

func computeHash(hashInfo HashInfo, str string, xofLen int) string {
	h := hashInfo.New()
	
	// Para SHAKE, precisamos ler os bytes com tamanho específico
	if hashInfo.IsShake {
		shake := h.(sha3.ShakeHash)
		shake.Write([]byte(str))
		result := make([]byte, xofLen)
		shake.Read(result)
		return hex.EncodeToString(result)
	}
	
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func worker(id int, config *Config, target string, hashInfo HashInfo, strChan <-chan string, results chan<- string, wg *sync.WaitGroup, stopFlag *atomic.Bool, attempts *uint64) {
	defer wg.Done()

	// Usa xoflen do config ou o padrão do hash
	xofLen := hashInfo.XofLen
	if config.XofLen > 0 {
		xofLen = config.XofLen
	}
	// Se ainda for 0 e for SHAKE, usa 16 como fallback
	if hashInfo.IsShake && xofLen == 0 {
		xofLen = 16
	}

	for str := range strChan {
		if stopFlag.Load() {
			return
		}

		atomic.AddUint64(attempts, 1)

		if config.Verbose > 0 && atomic.LoadUint64(attempts)%uint64(config.Verbose*1000) == 0 {
			fmt.Printf("\r[%s] Testing: %-20s | Attempts: %d", hashInfo.Name, str, atomic.LoadUint64(attempts))
		}

		computed := computeHash(hashInfo, str, xofLen)
		
		if computed == target {
			results <- str
			stopFlag.Store(true)
			return
		}
	}
}

func progressReporter(attempts *uint64, stopFlag *atomic.Bool, total uint64, hashName string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	for range ticker.C {
		if stopFlag.Load() {
			return
		}

		current := atomic.LoadUint64(attempts)
		elapsed := time.Since(startTime).Seconds()
		rate := float64(current) / elapsed

		var percent float64
		if total > 0 {
			percent = float64(current) / float64(total) * 100
		}

		fmt.Printf("\r[Progress] Tested: %d strings (%.2f%%) | Rate: %.0f str/s | Elapsed: %.0fs | Hash: %s",
			current, percent, rate, elapsed, hashName)
	}
}

func loadHashes(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(data), "\n")
	var hashes []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			hashes = append(hashes, line)
		}
	}
	return hashes, nil
}

func main() {
	var config Config

	version := flag.Bool("version", false, "Print version info")
	
	flag.StringVar(&config.Target, "hash", "", "Target hash to crack")
	flag.StringVar(&config.InputFile, "file", "", "File with list of hashes to crack")
	flag.IntVar(&config.MinLen, "min", 1, "Minimum string length")
	flag.IntVar(&config.MaxLen, "max", 8, "Maximum string length")
	flag.StringVar(&config.Prefix, "prefix", "", "String prefix")
	flag.StringVar(&config.Suffix, "suffix", "", "String suffix")
	flag.StringVar(&config.Charset, "charset", "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@#$%^&*()-_=+[]{}|;:,.<>?", "Character set")
	flag.IntVar(&config.Threads, "threads", runtime.NumCPU(), "Number of worker threads")
	flag.IntVar(&config.Verbose, "verbose", 1, "Verbose level (0=quiet, 1=normal)")
	flag.StringVar(&config.HashType, "type", "sha256", "Hash type")
	flag.IntVar(&config.XofLen, "xoflen", 0, "XOF length in bytes for SHAKE (default: 16 for shake128, 32 for shake256). Can be any value like 64, 128, etc.")
	flag.StringVar(&config.OutputFile, "output", "", "Output file for found strings")
	
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s (%s)\n", AppName, Version, VersionDate)
		fmt.Printf("Built with Go %s\n", runtime.Version())
		fmt.Println("\nSupported hash types (OpenSSL compatible):")
		fmt.Println("\n  MD5:")
		fmt.Println("    - md5 (openssl dgst -md5)")
		fmt.Println("\n  MD4:")
		fmt.Println("    - md4 (openssl dgst -md4)")
		fmt.Println("\n  SHA1:")
		fmt.Println("    - sha1 (openssl dgst -sha1)")
		fmt.Println("\n  RIPEMD-160:")
		fmt.Println("    - rmd160 (openssl dgst -rmd160)")
		fmt.Println("\n  SHA2:")
		fmt.Println("    - sha224 (openssl dgst -sha224)")
		fmt.Println("    - sha256 (openssl dgst -sha256)")
		fmt.Println("    - sha384 (openssl dgst -sha384)")
		fmt.Println("    - sha512 (openssl dgst -sha512)")
		fmt.Println("    - sha512-224 (openssl dgst -sha512-224)")
		fmt.Println("    - sha512-256 (openssl dgst -sha512-256)")
		fmt.Println("\n  SHA3:")
		fmt.Println("    - sha3-224 (openssl dgst -sha3-224)")
		fmt.Println("    - sha3-256 (openssl dgst -sha3-256)")
		fmt.Println("    - sha3-384 (openssl dgst -sha3-384)")
		fmt.Println("    - sha3-512 (openssl dgst -sha3-512)")
		fmt.Println("\n  SHAKE (requires -xoflen in OpenSSL 3.4.0+):")
		fmt.Println("    - shake128 (openssl dgst -shake128 -xoflen 16)")
		fmt.Println("    - shake256 (openssl dgst -shake256 -xoflen 32)")
		fmt.Println("    - Custom xoflen: -xoflen <bytes> (e.g., -xoflen 64 for 512-bit output)")
		fmt.Println("\n  SM3:")
		fmt.Println("    - sm3 (openssl dgst -sm3)")
		fmt.Println("\n  BLAKE2:")
		fmt.Println("    - blake2b512 (openssl dgst -blake2b512)")
		fmt.Println("    - blake2s256 (openssl dgst -blake2s256)")
		os.Exit(0)
	}

	if config.Target == "" && config.InputFile == "" {
		fmt.Println("ERROR: -hash or -file is required")
		flag.Usage()
		os.Exit(1)
	}

	hashInfo, ok := hashRegistry[config.HashType]
	if !ok {
		fmt.Printf("ERROR: Unsupported hash type: %s\n", config.HashType)
		fmt.Println("\nSupported types:")
		for k := range hashRegistry {
			fmt.Printf("  - %s\n", k)
		}
		os.Exit(1)
	}

	// Define xoflen
	xofLen := hashInfo.XofLen
	if config.XofLen > 0 {
		xofLen = config.XofLen
	}
	// Se for SHAKE e xofLen ainda for 0, usa 16 como fallback
	if hashInfo.IsShake && xofLen == 0 {
		xofLen = 16
	}

	var targets []string
	if config.InputFile != "" {
		var err error
		targets, err = loadHashes(config.InputFile)
		if err != nil {
			fmt.Printf("Error loading hash file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loaded %d hashes from file\n", len(targets))
	} else {
		targets = []string{config.Target}
	}

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		expectedLen := hashInfo.Length
		if hashInfo.IsShake {
			expectedLen = xofLen * 2 // em hex
		}

		if len(target) != expectedLen {
			fmt.Printf("WARNING: Hash length is %d but expected %d for %s (xoflen=%d bytes)\n", 
				len(target), expectedLen, hashInfo.Name, xofLen)
		}

		fmt.Printf("\n%s v%s\n", AppName, Version)
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("Hash type: %s\n", hashInfo.Name)
		fmt.Printf("Target hash: %s\n", target)
		if hashInfo.IsShake {
			fmt.Printf("XOF length: %d bytes (%d hex chars)\n", xofLen, xofLen*2)
		} else {
			fmt.Printf("Hash length: %d bytes\n", hashInfo.Length/2)
		}
		fmt.Println(strings.Repeat("=", 60))

		sg := NewStringGenerator(&config)
		totalCombinations := sg.GetTotalCount()

		fmt.Printf("\n[BRUTE-FORCE MODE]\n")
		fmt.Printf("Total combinations: %d\n", totalCombinations)
		fmt.Printf("Charset size: %d\n", len(config.Charset))
		fmt.Printf("String length: %d to %d\n", config.MinLen, config.MaxLen)
		fmt.Printf("Prefix: '%s'\n", config.Prefix)
		fmt.Printf("Suffix: '%s'\n", config.Suffix)
		fmt.Printf("Using %d threads\n", config.Threads)
		fmt.Println(strings.Repeat("-", 60))

		strChan := make(chan string, 10000)
		results := make(chan string, 10)
		stopFlag := &atomic.Bool{}
		var attempts uint64

		var wg sync.WaitGroup

		go func() {
			defer close(strChan)
			for {
				str, ok := sg.Next()
				if !ok {
					break
				}
				if stopFlag.Load() {
					return
				}
				strChan <- str
			}
		}()

		for i := 0; i < config.Threads; i++ {
			wg.Add(1)
			go worker(i, &config, target, hashInfo, strChan, results, &wg, stopFlag, &attempts)
		}

		go progressReporter(&attempts, stopFlag, totalCombinations, hashInfo.Name)

		go func() {
			wg.Wait()
			close(results)
			fmt.Println()
		}()

		var foundString string
		select {
		case str, ok := <-results:
			if ok {
				foundString = str
			}
		case <-time.After(24 * time.Hour):
			stopFlag.Store(true)
		}

		if foundString != "" {
			fmt.Printf("\nSUCCESS! String found: %s\n", foundString)
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("Hash type: %s\n", hashInfo.Name)
			fmt.Printf("Target hash: %s\n", target)
			if hashInfo.IsShake {
				fmt.Printf("XOF length: %d bytes\n", xofLen)
			}
			fmt.Printf("Found string: %s\n", foundString)
			fmt.Printf("Attempts: %d\n", atomic.LoadUint64(&attempts))
			fmt.Println(strings.Repeat("=", 60))
			
			if config.OutputFile != "" {
				output := fmt.Sprintf("Hash: %s\nType: %s\nString: %s\nAttempts: %d\n",
					target, hashInfo.Name, foundString, atomic.LoadUint64(&attempts))
				if hashInfo.IsShake {
					output += fmt.Sprintf("XOF Length: %d bytes\n", xofLen)
				}
				os.WriteFile(config.OutputFile, []byte(output), 0644)
				fmt.Printf("Results saved to: %s\n", config.OutputFile)
			}
		} else {
			fmt.Printf("\nString not found after %d attempts.\n", atomic.LoadUint64(&attempts))
		}
	}
}
