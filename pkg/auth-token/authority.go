package authtoken

import (
	"io/ioutil"
	
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"github.com/dgrijalva/jwt-go"
	"fmt"
)

type Authority interface {
	 IsInRole(role string) (bool, error)
} 

type ContextAuthority struct {
	pathToPublicKey string
	ctx             context.Context
}

func NewContextAuthority(pathToPublicKey string, ctx context.Context) Authority {
	return &ContextAuthority{
		pathToPublicKey: pathToPublicKey,
		ctx:             ctx,
	}
}

// IsInRole will inspect the context for a token, validate it, and signal if the
// token signals role affinity
func (ca *ContextAuthority) IsInRole(role string) (bool, error) {

	md, found := metadata.FromContext(ca.ctx)

	if !found {
		return false, TokenNotFound
	}

	t := md[tokenContextName]

	if t == nil {
		return false, TokenNotFound
	}

	ok, err := ca.isTokenValidAndContainsRole(t[0], role)

	if !ok {
		return false, err
	}

	return true, nil
}

// isTokenValidAndContainsRole checks if token for a role if valid, not expired and uses the correct signing method
// does a lot rather than exposing the token
func (ca *ContextAuthority) isTokenValidAndContainsRole(signedTokenAsString, role string) (bool, error) {

	// NOTE : loads keys on every request
	verifyBytes, err := ioutil.ReadFile(ca.pathToPublicKey)
	if err != nil {
		return false, err
	}

	authenticationVerifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return false, err
	}

	token, err := jwt.Parse(signedTokenAsString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return false, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return authenticationVerifyKey, nil
	})

	// TODO add some leeway on parsing
	ve, ok := err.(*jwt.ValidationError)
	if !ok {
		return false, ve
	}

	if token.Valid {
		_, ok := token.Claims[role]
		return ok, nil
	}

	panic("SignedToken is neither valid or invalid")

}
