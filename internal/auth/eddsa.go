package auth

import "crypto/ed25519"

type EddsaKeys struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
}
