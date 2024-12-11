package middleware

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"
	"vector-ai/model"

	"github.com/golang-jwt/jwt"
	"goyave.dev/goyave/v4"
)

var (
	verifyKey *rsa.PublicKey
)

func Authentication(next goyave.Handler) goyave.Handler {

	return func(res *goyave.Response, req *goyave.Request) {

		// get JWT
		tokenString, ok := req.BearerToken()
		if !ok {
			fmt.Println("NOT OK")
			//log.Fatal(ok)
		}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &model.ClerkClaims{}, func(token *jwt.Token) (interface{}, error) {

			var rsa string
			if env := os.Getenv("GOYAVE_ENV"); env == "production" {
				rsa = "resources/rsa/public.pem"
			} else {
				rsa = "resources/rsa/local.pem"
			}

			// Use public key to verify
			data, err := os.ReadFile(rsa)
			check(err)

			verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(data)
			check(err)

			return verifyKey, nil
		})
		check(err)

		req.Extra["jwt_claims"] = token.Claims.(*model.ClerkClaims)

		next(res, req) // Pass to the next handler
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(time.Now())
		fmt.Println(err)
		//log.Fatal(err)
	}
}
