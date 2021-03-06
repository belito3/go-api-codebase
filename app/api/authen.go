package api

import (
	"github.com/belito3/go-web-api/pkg/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Refs: https://www.sohamkamani.com/golang/2019-01-01-jwt-authentication/
// TODO: ranh bo sung refresh token
// Create the JWT key used to create the signature
//var jwtKey = []byte(config.C.JWTSecretKey)

var jwtKey = []byte("my_secret_key")

var users = map[string]string{
	"app_key1": "secret_key1",
	"app_key2": "secret_key2",
}


// Create a struct to read the username and password from the request
type Credentials struct {
	AppKey 		string `json:"app_key"`
	SecretKey 	string `json:"secret_key"`
}

// Create a struct that will be encoded to a JWT
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	AppKey string `json:"app_key"`
	jwt.StandardClaims
}


// Create the signin handler
func Signin(c *gin.Context) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	if err := c.ShouldBindJSON(&creds); err!=nil {
		responseError(c, http.StatusBadRequest, err.Error())
		logger.Errorf(nil, err.Error())
		//c.Abort()
		return
	}

	logger.Infof(nil, "abc %v", creds)
	// Get the expected password for our in memory map
	expectedSecretKey, ok := users[creds.AppKey]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an
	if !ok || expectedSecretKey != creds.SecretKey {
		responseError(c, http.StatusUnauthorized, "Status Unauthorized")
		return
	}

	// TODO: bo sung co che refresh token
	// Declare the expiration time of the token
	expirationTime := time.Now().Add(2 * 365 * 24 * time.Hour)
	// Create the JWT claims, which include the username and expiry time
	claims := &Claims{
		AppKey: creds.AppKey,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		responseError(c, http.StatusInternalServerError, "Internal Server Error")
		logger.Errorf(nil, err.Error())
		return
	}

	r := map[string]interface{}{
		"token": tokenString,
	}

	responseSuccess(c, http.StatusOK, r)
}


func TokenAuthMiddleware() gin.HandlerFunc{
	// Do some initialization logic here
	// Whatever you define before the return statement will be executed only once
	// Foo()

	return func(c *gin.Context){
		// Get token from header
		token := c.GetHeader("token")
		if len(token) == 0 {
			logger.Errorf(nil,"Token is not set")
			responseError(c, http.StatusUnauthorized, "API token required")
			c.Abort()
			return
		}

		claims := &Claims{}

		// Parse the JWT string and store the result in `claims`
		// This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token)(interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				logger.Errorf(nil,"Token is invalid. SignatureInvalid. Token=%v", token)
				responseError(c, http.StatusUnauthorized, "Token is invalid")
				c.Abort()
				return
			}
			responseError(c, http.StatusBadRequest, "Token is invalid")
			logger.Errorf(nil,"Token is invalid. Bad request. Token=%v", token)
			c.Abort()
			return
		}
		if !tkn.Valid {
			logger.Errorf(nil,"Token is invalid. Token=%v", token)
			responseError(c, http.StatusUnauthorized, "Token is invalid")
			c.Abort()
			return
		}
		// It means that after our middleware is done executing
		// we can pass on request handler to the next func in the chain.
		c.Next()
	}
}


// Create Welcome handle
func Welcome(c *gin.Context) {
	responseSuccess(c, http.StatusOK, nil)
}

