package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"go-stock/backend/provider/update"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: update-manifest <keygen|sign|verify> [flags]")
	}
	switch os.Args[1] {
	case "keygen":
		keygen(os.Args[2:])
	case "sign":
		sign(os.Args[2:])
	case "verify":
		verify(os.Args[2:])
	default:
		fail("unknown command: " + os.Args[1])
	}
}

func verify(args []string) {
	flags := flag.NewFlagSet("verify", flag.ExitOnError)
	manifestURL := flags.String("url", update.DefaultManifestURL, "signed update manifest URL")
	currentVersion := flags.String("current", "0.0.0", "version to compare with the manifest")
	_ = flags.Parse(args)
	provider, err := update.NewDefaultManifestProvider(*manifestURL)
	if err != nil {
		fail(err.Error())
	}
	result, err := provider.Check(context.Background(), *currentVersion)
	if err != nil {
		fail(err.Error())
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		fail(err.Error())
	}
	fmt.Println(string(encoded))
}

func keygen(args []string) {
	flags := flag.NewFlagSet("keygen", flag.ExitOnError)
	outDir := flags.String("out", ".update-signing", "directory for the Ed25519 key pair")
	_ = flags.Parse(args)
	if err := os.MkdirAll(*outDir, 0700); err != nil {
		fail(err.Error())
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fail(err.Error())
	}
	privatePath := filepath.Join(*outDir, "manifest-private.key")
	publicPath := filepath.Join(*outDir, "manifest-public.key")
	if err := os.WriteFile(privatePath, []byte(base64.StdEncoding.EncodeToString(privateKey)+"\n"), 0600); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(publicPath, []byte(base64.StdEncoding.EncodeToString(publicKey)+"\n"), 0644); err != nil {
		fail(err.Error())
	}
	fmt.Println(publicPath)
}

func sign(args []string) {
	flags := flag.NewFlagSet("sign", flag.ExitOnError)
	keyPath := flags.String("key", ".update-signing/manifest-private.key", "base64-encoded Ed25519 private key")
	manifestPath := flags.String("manifest", "manifest.json", "manifest JSON to sign")
	_ = flags.Parse(args)
	encodedKey, err := os.ReadFile(*keyPath)
	if err != nil {
		fail(err.Error())
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(string(trimSpace(encodedKey)))
	if err != nil || len(privateKeyBytes) != ed25519.PrivateKeySize {
		fail("invalid Ed25519 private key")
	}
	manifestBytes, err := os.ReadFile(*manifestPath)
	if err != nil {
		fail(err.Error())
	}
	var manifest update.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		fail(err.Error())
	}
	if err := update.SignManifest(&manifest, ed25519.PrivateKey(privateKeyBytes)); err != nil {
		fail(err.Error())
	}
	output, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*manifestPath, append(output, '\n'), 0644); err != nil {
		fail(err.Error())
	}
}

func trimSpace(value []byte) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\n' || value[start] == '\r' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\n' || value[end-1] == '\r' || value[end-1] == '\t') {
		end--
	}
	return string(value[start:end])
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
