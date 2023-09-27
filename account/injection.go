package main

import (
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/j03hanafi/hapalin-app/account/handler"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/j03hanafi/hapalin-app/account/repository"
	"github.com/j03hanafi/hapalin-app/account/service"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

// inject will initialize a handler starting from data sources
// which inject into repository layer
// which inject into service layer
// which inject into handler layer
func inject(d *dataSources) (*gin.Engine, error) {
	l := logger.Get()

	l.Info("injecting data sources into repository layer")

	userRepository := repository.NewUserRepository(d.DB)
	tokenRepository := repository.NewTokenRepository(d.RedisClient)

	l.Info("injecting repository layer into service layer")

	userService := service.NewUserService(&service.USConfig{
		UserRepository: userRepository,
	})

	// load rsa keys
	privateKeyPath := os.Getenv("PRIVATE_KEY_FILE")
	privateKeyFile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		l.Fatal("failed to load private key",
			zap.Error(err),
		)
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyFile)
	if err != nil {
		l.Fatal("failed to parse private key",
			zap.Error(err),
		)
		return nil, err
	}

	publicKeyPath := os.Getenv("PUBLIC_KEY_FILE")
	publicKeyFile, err := os.ReadFile(publicKeyPath)
	if err != nil {
		l.Fatal("failed to load public key",
			zap.Error(err),
		)
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyFile)
	if err != nil {
		l.Fatal("failed to parse public key",
			zap.Error(err),
		)
		return nil, err
	}

	// load token service env variable
	refreshSecret := os.Getenv("REFRESH_SECRET")
	idTokenExp := os.Getenv("ID_TOKEN_EXP")
	refreshTokenExp := os.Getenv("REFRESH_TOKEN_EXP")

	idExp, err := strconv.ParseInt(idTokenExp, 0, 64)
	if err != nil {
		l.Fatal("failed to parse id token expiration",
			zap.Error(err),
		)
		return nil, err
	}

	refreshExp, err := strconv.ParseInt(refreshTokenExp, 0, 64)
	if err != nil {
		l.Fatal("failed to parse refresh token expiration",
			zap.Error(err),
		)
		return nil, err
	}

	tokenService := service.NewTokenService(&service.TSConfig{
		PrivateKey:            privateKey,
		PublicKey:             publicKey,
		RefreshSecret:         refreshSecret,
		IDExpirationSecs:      idExp,
		RefreshExpirationSecs: refreshExp,
		TokenRepository:       tokenRepository,
	})

	// read HANDLER_TIMEOUT
	handlerTimeout := os.Getenv("HANDLER_TIMEOUT")
	ht, err := strconv.ParseInt(handlerTimeout, 0, 64)
	if err != nil {
		l.Error("failed to parse handler timeout",
			zap.Error(err),
		)
		return nil, err
	}

	// initialize gin.Engine
	router := gin.New()
	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))

	handler.NewHandler(&handler.Config{
		R:               router,
		UserService:     userService,
		TokenService:    tokenService,
		BaseURL:         os.Getenv("ACCOUNT_API_URL"),
		TimeoutDuration: time.Duration(ht) * time.Second,
	})

	return router, nil
}
