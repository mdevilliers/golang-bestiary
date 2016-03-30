package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"

	auth "github.com/mdevilliers/golang-bestiary/pkg/auth-token"
)

const (
	port = ":50000"
)

func main() {

	fmt.Println("auth service started")

	privateKey, err := generateNewPrivateKey()

	if err != nil {
		log.Fatal(err)
		return
	}

	s := &authenticationServer{
		privateKey: privateKey,
	}

	http.HandleFunc("/authenticate", s.authenticate)
	http.HandleFunc("/publickey", s.serverPublicKey)
	http.HandleFunc("/rotate", s.rotateKeys)
	http.ListenAndServe(port, nil)
}

type authenticationServer struct {
	privateKey *rsa.PrivateKey
}

func (au *authenticationServer) authenticate(w http.ResponseWriter, r *http.Request) {

	log.Println("authentication request received")

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		log.Println("/authenticate error", "username:", username, "password:", password)
		http.Error(w, "GoAway!", 403)
		return
	}

	var token auth.Token

	if username == "mark" && password == "secret" {

		token = auth.NewToken(
			map[string]interface{}{
				"role1": true,
				"role2": true,
				"role3": true,
				"role4": true,
			},
		)
	} else {
		token = auth.NewToken(
			map[string]interface{}{
				"read": true,
			},
		)
	}

	signer := auth.NewTokenSigner(au.privateKey)
	signedToken, err := signer.Sign(token)

	if err != nil {
		log.Println("/authenticate error", "username:", username, "password:", password)
		http.Error(w, "GoAway!", 403)
		return
	}

	fmt.Fprintf(w, "%s", signedToken)

}

func (au *authenticationServer) serverPublicKey(w http.ResponseWriter, r *http.Request) {

	log.Println("public key request received")

	publicDer, err := x509.MarshalPKIXPublicKey(&au.privateKey.PublicKey)
	if err != nil {
		log.Println("Failed to get der format for PublicKey.", err)
		http.Error(w, "Server Errors", 500)
		return
	}

	pem := string(pem.EncodeToMemory(&pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicDer,
	}))

	fmt.Fprintf(w, "%s", pem)
}

func (au *authenticationServer) rotateKeys(w http.ResponseWriter, r *http.Request) {

	log.Println("rotate key request received")

	newPrivateKey, err := generateNewPrivateKey()

	if err != nil {
		log.Println("/rotate error", err)
		http.Error(w, "GoAway!", 403)
		return
	}

	au.privateKey = newPrivateKey
}

func generateNewPrivateKey() (*rsa.PrivateKey, error) {

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	
	if err != nil {
		return nil, err
	}

	return priv, nil
}
