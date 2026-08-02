# Bruteforce-Hash-OpenSSL
Cross-platform, multicore bruteforce-hash-openssl in Pure Go

# Bruteforce Hash OpenSSL

Ferramenta de brute-force para quebrar hashes de algoritmos suportados pelo OpenSSL.

## Compilação

```bash
go mod init bruteforce-hash-openssl
go get golang.org/x/crypto/blake2b golang.org/x/crypto/blake2s golang.org/x/crypto/sha3 golang.org/x/crypto/ripemd160 golang.org/x/crypto/md4 github.com/emmansun/gmsm/sm3
go build -ldflags=-s -o openssl-cracker main.go
```

## Uso

```bash
./openssl-cracker -hash <HASH> -type <ALG> -min <MIN> -max <MAX>
```

## Parâmetros

| Parâmetro | Descrição | Padrão |
|-----------|-----------|--------|
| `-hash` | Hash alvo | - |
| `-file` | Arquivo com hashes (um por linha) | - |
| `-type` | Algoritmo | `sha256` |
| `-min` | Tamanho mínimo | `1` |
| `-max` | Tamanho máximo | `8` |
| `-prefix` | Prefixo conhecido | `""` |
| `-suffix` | Sufixo conhecido | `""` |
| `-charset` | Conjunto de caracteres | `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!@#$%^&*()-_=+[]{}|;:,.<>?` |
| `-threads` | Número de threads | `runtime.NumCPU()` |
| `-verbose` | Nível de verbose (0=quiet, 1=normal) | `1` |
| `-xoflen` | Comprimento em bytes para SHAKE | `16`/`32` |
| `-output` | Arquivo de saída | - |
| `-version` | Exibe versão | - |

## Algoritmos

`md5`, `md4`, `sha1`, `rmd160`, `sha224`, `sha256`, `sha384`, `sha512`, `sha512-224`, `sha512-256`, `sha3-224`, `sha3-256`, `sha3-384`, `sha3-512`, `shake128`, `shake256`, `sm3`, `blake2b512`, `blake2s256`

## Exemplos

```bash
# MD5
./openssl-cracker -hash 5f4dcc3b5aa765d61d8327deb882cf99 -type md5 -min 1 -max 8

# SHA256
./openssl-cracker -hash 8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918 -type sha256 -min 1 -max 8

# SHA256 com prefixo
./openssl-cracker -hash 8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918 -type sha256 -prefix "abc" -min 4 -max 6

# SHAKE128 com xoflen 32
./openssl-cracker -hash 3f1c0f4b9a5e8d7c6b2a1f0e9d8c7b6a -type shake128 -xoflen 32 -min 1 -max 6

# SHAKE256 com xoflen 64
./openssl-cracker -hash 3f1c0f4b9a5e8d7c6b2a1f0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c -type shake256 -xoflen 64 -min 1 -max 6

# BLAKE2b-512
./openssl-cracker -hash a71079d42853dea26e453004338670a53814b78137ffbed07603a41d76a483aa9bc33b582f77d30a65e6f29a896c0411f38312e1d66e0bf16386c86a89bea572 -type blake2b512 -min 1 -max 6

# BLAKE2s-256
./openssl-cracker -hash 8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918 -type blake2s256 -min 1 -max 6

# SM3
./openssl-cracker -hash 1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b -type sm3 -min 1 -max 6

# Múltiplos hashes
./openssl-cracker -file hashes.txt -type sha256 -min 1 -max 6

# Com charset personalizado
./openssl-cracker -hash hash.txt -type sha1 -charset "abcdefghijklmnopqrstuvwxyz" -min 3 -max 5

# Com prefixo e sufixo
./openssl-cracker -hash hash.txt -type sha512 -prefix "user_" -suffix "_2024" -min 4 -max 10

# Salvar resultado
./openssl-cracker -hash hash.txt -type sha256 -min 1 -max 6 -output resultado.txt

# Versão
./openssl-cracker -version
```

## Gerando Hashes

```bash
echo -n "minha_string" | openssl dgst -sha256
echo -n "minha_string" | openssl dgst -shake128 -xoflen 16
echo -n "minha_string" | openssl dgst -blake2b512
```
