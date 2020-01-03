package service

import (
	"crypto/rsa"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs"
	"github.com/dgrijalva/jwt-go"

	"github.com/klyngen/flightlogger/common"
	"github.com/klyngen/flightlogger/configuration"
)

// FlightLogService describes service containing business-logic
type FlightLogService struct {
	database     common.FlightLogDatabase
	config       configuration.ApplicationConfig
	email        common.EmailServiceInterface
	sessionstore *scs.SessionManager
	signingkey   *rsa.PrivateKey
	verifykey    *rsa.PublicKey
}

// NewService creates a new flightlogservice
func NewService(database common.FlightLogDatabase, email common.EmailServiceInterface, config configuration.ApplicationConfig) *FlightLogService {
	sign, verify := getSigningKeys(config.PrivateKeyPath, config.PublicKeyPath)

	return &FlightLogService{database: database,
		config:       config,
		signingkey:   sign,
		verifykey:    verify,
		email:        email,
		sessionstore: createSessionStore(config),
	}
}

// NewServiceWithPersistedSession requires some sort of persisted storage
func NewServiceWithPersistedSession(database common.FlightLogDatabase, email common.EmailServiceInterface, config configuration.ApplicationConfig, sessionStore scs.Store) *FlightLogService {
	sign, verify := getSigningKeys(config.PrivateKeyPath, config.PublicKeyPath)

	sessionManager := createSessionStore(config)
	sessionManager.Store = sessionStore

	return &FlightLogService{database: database,
		config:       config,
		signingkey:   sign,
		verifykey:    verify,
		email:        email,
		sessionstore: sessionManager,
	}
}

func createSessionStore(config configuration.ApplicationConfig) *scs.SessionManager {
	sessionManager := scs.New()
	sessionManager.Lifetime = time.Duration(config.Tokenexpiration) * time.Minute
	sessionManager.IdleTimeout = time.Duration(config.Tokenexpiration/2) * time.Minute
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = true
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.Domain = config.ServerURL

	return sessionManager
}

func getSigningKeys(privatekeypath string, publickeypath string) (*rsa.PrivateKey, *rsa.PublicKey) {
	var signBytes, verifyBytes []byte
	var signKey *rsa.PrivateKey
	var verifyKey *rsa.PublicKey
	var err error

	if signBytes, err = ioutil.ReadFile(privatekeypath); err != nil {
		log.Fatalf("Unable to read PrivateKey: %v", err)
	}

	if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
		log.Fatalf("Unable to parse PrivateKey: %v", err)
	}

	if verifyBytes, err = ioutil.ReadFile(publickeypath); err != nil {
		log.Fatalf("Unable to read PublicKey: %v", err)
	}

	if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
		log.Fatalf("Unable to parse PublicKey: %v", err)
	}
	return signKey, verifyKey
}

// VerifyTokenString - tries to verify the token using the certificate
func (f *FlightLogService) VerifyTokenString(tokenString string) (jwt.Claims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return f.verifykey, nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims, err
}

// GetSessionManager returns the configured sessionmanager
func (f *FlightLogService) GetSessionManager() *scs.SessionManager {
	return f.sessionstore
}

// SessionParameters is used to define consts
type sessionParameters string

const (
	sessionParamUserID sessionParameters = "userid"
	sessionParamRoles  sessionParameters = "roles"
)
