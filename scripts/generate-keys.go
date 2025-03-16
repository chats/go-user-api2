package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Error generating Ed25519 key pair: %v\n", err)
		os.Exit(1)
	}

	// Convert keys to hex
	privateKeyHex := hex.EncodeToString(privateKey)
	publicKeyHex := hex.EncodeToString(publicKey)

	fmt.Println("PASETO Ed25519 Key Pair:")
	fmt.Println("-----------------------")
	fmt.Printf("Private Key (for .env): %s\n", privateKeyHex)
	fmt.Printf("Public Key (for .env) : %s\n", publicKeyHex)
	fmt.Println()
	fmt.Println("Add these to your .env file as:")
	fmt.Println("PASETO_PRIVATE_KEY=" + privateKeyHex)
	fmt.Println("PASETO_PUBLIC_KEY=" + publicKeyHex)
}
