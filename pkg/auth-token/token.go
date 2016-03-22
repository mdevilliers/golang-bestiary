package authtoken

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"time"
)

var (
	// TokenNotFound signals no signed token found in Context
	TokenNotFound = errors.New("SignedToken not found in Context")

	tokenContextName = "signed-token"
)

// Token is a collections of claims in plain sight
type Token struct {
	claims map[string]interface{} 
}

// NewToken returns a Token with a set of claims
func NewToken(claims map[string]interface{}) Token {
	return Token{
		claims: claims,
	}
}

// SignedToken is an signed JWT token
type SignedToken struct {
	encryptedToken string
}

func (st SignedToken) String() string {
	return st.encryptedToken
}

type TokenSigner struct {
	key *rsa.PrivateKey
}

func NewTokenSigner(pathToPrivateKey string) (*TokenSigner, error) {

	signingBytes, err := ioutil.ReadFile(pathToPrivateKey)

	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(signingBytes)
	if err != nil {
		return nil, err
	}

	return &TokenSigner{key: key}, nil
}

// Sign will take a Token and return a SignedToken
func (ts *TokenSigner) Sign(token Token) (*SignedToken, error) {

	jwtToken := jwt.New(jwt.SigningMethodRS256)

	for claimName, claim := range token.claims {
		jwtToken.Claims[claimName] = claim
	}

	// add trusted claims last
	// this should really be dictated by a policy
	jwtToken.Claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	// Sign and get the complete encoded token as a string
	tokenString, err := jwtToken.SignedString(ts.key)

	if err != nil {
		return nil, err
	}

	return &SignedToken{encryptedToken: tokenString}, nil
}

func NewContextMarshaller(ctx context.Context) *ContextMarshaller {
	return &ContextMarshaller{
		ctx: ctx,
	}
}

type ContextMarshaller struct {
	ctx context.Context
}

// Marshal appends a signed token to a context
func (cm *ContextMarshaller) Marshal(token *SignedToken) context.Context {
	md := metadata.Pairs(tokenContextName, token.encryptedToken)
	ctx := metadata.NewContext(cm.ctx, md)
	return ctx
}

type ContextAuthority struct {
	pathToPublicKey string
	ctx             context.Context
}

func NewContextAuthority(pathToPublicKey string, ctx context.Context) *ContextAuthority {
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

	// WARNING : loads keys on every request
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
	// if ve, ok := err.(*jwt.ValidationError); ok {
	// 	if ve.Errors&jwt.ValidationErrorMalformed != 0 {

	// 		return false, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])

	// 		fmt.Println("That's not even a token")
	// 	} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
	// 		// Token is either expired or not active yet
	// 		fmt.Println("Timing is everything")
	// 	} else {
	// 		fmt.Println("UnkownError:", err)
	// 	}
	// } else {
	// 	fmt.Println("InValidToken:", err)
	// }
}
