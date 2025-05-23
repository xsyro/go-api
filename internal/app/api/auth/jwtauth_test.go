package auth

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	TokenAuthHS256 *JWTAuth
	TokenSecret    = []byte("secretpass")

	TokenAuthRS256 *JWTAuth

	PrivateKeyRS256String = `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBALxo3PCjFw4QjgOX06QCJIJBnXXNiEYwDLxxa5/7QyH6y77nCRQy
J3x3UwF9rUD0RCsp4sNdX5kOQ9PUyHyOtCUCAwEAAQJARjFLHtuj2zmPrwcBcjja
IS0Q3LKV8pA0LoCS+CdD+4QwCxeKFq0yEMZtMvcQOfqo9x9oAywFClMSlLRyl7ng
gQIhAOyerGbcdQxxwjwGpLS61Mprf4n2HzjwISg20cEEH1tfAiEAy9dXmgQpDPir
C6Q9QdLXpNgSB+o5CDqfor7TTyTCovsCIQDNCfpu795luDYN+dvD2JoIBfrwu9v2
ZO72f/pm/YGGlQIgUdRXyW9kH13wJFNBeBwxD27iBiVj0cbe8NFUONBUBmMCIQCN
jVK4eujt1lm/m60TlEhaWBC3p+3aPT2TqFPUigJ3RQ==
-----END RSA PRIVATE KEY-----
`

	PublicKeyRS256String = `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBALxo3PCjFw4QjgOX06QCJIJBnXXNiEYw
DLxxa5/7QyH6y77nCRQyJ3x3UwF9rUD0RCsp4sNdX5kOQ9PUyHyOtCUCAwEAAQ==
-----END PUBLIC KEY-----
`
)

func init() {
	TokenAuthHS256 = New(jwa.HS256.String(), TokenSecret, nil, jwt.WithAcceptableSkew(30*time.Second))
}

//
// Tests
//

func TestSimple(t *testing.T) {
	r := chi.NewRouter()

	r.Use(
		Verifier(TokenAuthHS256),
		Authenticator(TokenAuthHS256),
	)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	// sending unauthorized requests
	if status, resp := testRequest(t, ts, "GET", "/", nil, nil); status != 401 || resp != "no token found\n" {
		t.Fatalf(resp)
	}

	h := http.Header{}
	h.Set("Authorization", "BEARER "+newJwtToken([]byte("wrong"), map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	h.Set("Authorization", "BEARER asdf")
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	// wrong token secret and wrong alg
	h.Set("Authorization", "BEARER "+newJwt512Token([]byte("wrong"), map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	// correct token secret but wrong alg
	h.Set("Authorization", "BEARER "+newJwt512Token(TokenSecret, map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}

	// correct token, but has expired within the skew time
	h.Set("Authorization", "BEARER "+newJwtToken(TokenSecret, map[string]interface{}{"exp": time.Now().Unix() - 29}))
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 200 || resp != "welcome" {
		fmt.Println("status", status, "resp", resp)
		t.Fatalf(resp)
	}

	// correct token, but has expired outside of the skew time
	h.Set("Authorization", "BEARER "+newJwtToken(TokenSecret, map[string]interface{}{"exp": time.Now().Unix() - 31}))
	if status, resp := testRequest(t, ts, "GET", "/", h, nil); status != 401 || resp != "token is expired\n" {
		t.Fatalf(resp)
	}

	// sending authorized requests
	if status, resp := testRequest(t, ts, "GET", "/", newAuthHeader(), nil); status != 200 || resp != "welcome" {
		t.Fatalf(resp)
	}
}

func TestSimpleRSA(t *testing.T) {
	privateKeyBlock, _ := pem.Decode([]byte(PrivateKeyRS256String))

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		t.Fatalf(err.Error())
	}

	publicKeyBlock, _ := pem.Decode([]byte(PublicKeyRS256String))

	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		t.Fatalf(err.Error())
	}

	TokenAuthRS256 = New(jwa.RS256.String(), privateKey, publicKey)

	claims := map[string]interface{}{
		"key":  "val",
		"key2": "val2",
		"key3": "val3",
	}

	_, tokenString, err := TokenAuthRS256.Encode(claims)
	if err != nil {
		t.Fatalf("Failed to encode claims %s\n", err.Error())
	}

	token, err := TokenAuthRS256.Decode(tokenString)
	if err != nil {
		t.Fatalf("Failed to decode token string %s\n", err.Error())
	}

	tokenClaims, err := token.AsMap(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(claims, tokenClaims) {
		t.Fatalf("The decoded claims don't match the original ones\n")
	}
}

func TestSimpleRSAVerifyOnly(t *testing.T) {
	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJrZXkiOiJ2YWwiLCJrZXkyIjoidmFsMiIsImtleTMiOiJ2YWwzIn0.kLEK3FZZPsAlQNKR5yHyjRyrlCJFhvKmrh7o-GqDT_zaGQgvb0Dufp8uNSMeOFAlLGK5FbKX7BckjJqfvEyrTQ"
	claims := map[string]interface{}{
		"key":  "val",
		"key2": "val2",
		"key3": "val3",
	}

	publicKeyBlock, _ := pem.Decode([]byte(PublicKeyRS256String))
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		t.Fatalf(err.Error())
	}

	TokenAuthRS256 = New(jwa.RS256.String(), nil, publicKey)

	_, _, err = TokenAuthRS256.Encode(claims)
	if err == nil {
		t.Fatalf("Expecting error when encoding claims without signing key")
	}

	token, err := TokenAuthRS256.Decode(tokenString)
	if err != nil {
		t.Fatalf("Failed to decode token string %s\n", err.Error())
	}

	tokenClaims, err := token.AsMap(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(claims, tokenClaims) {
		t.Fatalf("The decoded claims don't match the original ones\n")
	}
}

func TestMore(t *testing.T) {
	r := chi.NewRouter()

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(Verifier(TokenAuthHS256))

		authenticator := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token, _, err := FromContext(r.Context())

				if err != nil {
					http.Error(w, ErrorReason(err).Error(), http.StatusUnauthorized)
					return
				}

				if err := jwt.Validate(token); err != nil {
					http.Error(w, ErrorReason(err).Error(), http.StatusUnauthorized)
					return
				}

				// Token is authenticated, pass it through
				next.ServeHTTP(w, r)
			})
		}
		r.Use(authenticator)

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := FromContext(r.Context())

			if err != nil {
				w.Write([]byte(fmt.Sprintf("error! %v", err)))
				return
			}

			w.Write([]byte(fmt.Sprintf("protected, user:%v", claims["user_id"])))
		})
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	// sending unauthorized requests
	if status, resp := testRequest(t, ts, "GET", "/admin", nil, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}

	h := http.Header{}
	h.Set("Authorization", "BEARER "+newJwtToken([]byte("wrong"), map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	h.Set("Authorization", "BEARER asdf")
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	// wrong token secret and wrong alg
	h.Set("Authorization", "BEARER "+newJwt512Token([]byte("wrong"), map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}
	// correct token secret but wrong alg
	h.Set("Authorization", "BEARER "+newJwt512Token(TokenSecret, map[string]interface{}{}))
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 401 || resp != "token is unauthorized\n" {
		t.Fatalf(resp)
	}

	h = newAuthHeader(map[string]interface{}{"exp": EpochNow() - 1000})
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 401 || resp != "token is expired\n" {
		t.Fatalf(resp)
	}

	// sending authorized requests
	if status, resp := testRequest(t, ts, "GET", "/", nil, nil); status != 200 || resp != "welcome" {
		t.Fatalf(resp)
	}

	h = newAuthHeader((map[string]interface{}{"user_id": 31337, "exp": ExpireIn(5 * time.Minute)}))
	if status, resp := testRequest(t, ts, "GET", "/admin", h, nil); status != 200 || resp != "protected, user:31337" {
		t.Fatalf(resp)
	}
}

//
// Test helper functions
//

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header http.Header, body io.Reader) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}

	for k, v := range header {
		req.Header.Set(k, v[0])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func newJwtToken(secret []byte, claims ...map[string]interface{}) string {
	token := jwt.New()
	if len(claims) > 0 {
		for k, v := range claims[0] {
			token.Set(k, v)
		}
	}

	tokenPayload, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, secret))
	if err != nil {
		log.Fatal(err)
	}
	return string(tokenPayload)
}

func newJwt512Token(secret []byte, claims ...map[string]interface{}) string {
	// use-case: when token is signed with a different alg than expected
	token := jwt.New()
	if len(claims) > 0 {
		for k, v := range claims[0] {
			token.Set(k, v)
		}
	}
	tokenPayload, err := jwt.Sign(token, jwt.WithKey(jwa.HS512, secret))
	if err != nil {
		log.Fatal(err)
	}
	return string(tokenPayload)
}

func newAuthHeader(claims ...map[string]interface{}) http.Header {
	h := http.Header{}
	h.Set("Authorization", "BEARER "+newJwtToken(TokenSecret, claims...))
	return h
}
