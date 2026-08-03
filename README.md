# BruteForce-Hash-OpenSSL
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://github.com/pedroalbanese/bruteforce-hash-openssl/blob/master/LICENSE.md) 
[![GoDoc](https://godoc.org/github.com/pedroalbanese/bruteforce-hash-openssl?status.png)](http://godoc.org/github.com/pedroalbanese/bruteforce-hash-openssl)
[![GitHub downloads](https://img.shields.io/github/downloads/pedroalbanese/bruteforce-hash-openssl/total.svg?logo=github&logoColor=white)](https://github.com/pedroalbanese/bruteforce-hash-openssl/releases)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pedroalbanese/bruteforce-hash-openssl)](https://golang.org)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/pedroalbanese/bruteforce-hash-openssl)](https://github.com/pedroalbanese/bruteforce-hash-openssl/releases)

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
| `-charset` | Conjunto de caracteres | `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz` |
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
# Com charset personalizado
./bruteforce-hash-openssl -hash hash.txt -type sha1 -charset "abcdefghijklmnopqrstuvwxyz" -min 3 -max 5

# Com prefixo e sufixo
./bruteforce-hash-openssl-hash hash.txt -type sha512 -prefix "user_" -suffix "_2024" -min 4 -max 10

# Salvar resultado
./bruteforce-hash-openssl -hash hash.txt -type sha256 -min 1 -max 6 -output resultado.txt

# Versão
./bruteforce-hash-openssl -version
```

## Gerando Hashes

```bash
echo -n "minha_string" | openssl dgst -sha256
echo -n "minha_string" | openssl dgst -shake128 -xoflen 16
echo -n "minha_string" | openssl dgst -blake2b512
```

Confira também: https://github.com/pedroalbanese/bruteforce-salted-openssl

## License

This project is licensed under the ISC License.

#### Copyright (c) 2020-2026 Pedro F. Albanese - ALBANESE Research Lab.  
Todos os direitos de propriedade intelectual sobre este software pertencem ao autor, Pedro F. Albanese. Vide Lei 9.610/98, Art. 7º, inciso XII.
