package authtoken

import (
	"crypto/rsa"
	"errors"
	
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
