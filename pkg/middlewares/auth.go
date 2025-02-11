package middlewares

import (
	"context"
	"fmt"
	"strings"

	"github.com/CTO2BPublic/passage-server/pkg/errors"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/tracing"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
)

var OidcProvider *oidc.Provider

// Auth middleware function to handle both OIDC and JWT authentication
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, span := Tracer.Start(c.Request.Context(), "middlewares.Auth")
		defer span.End()

		// Skip midleware if Auth is disabled
		if !Config.Auth.OIDC.Enabled && !Config.Auth.JWT.Enabled {
			claims := models.NewUnauthenticatedUserClaims()

			// Extract specific claims dynamically
			username := claims.GetString(Config.Auth.JWT.UsernameClaim)
			groups := claims.GetStringSlice(Config.Auth.JWT.GroupsClaim)

			log.Debug().
				Str("username", username).
				Str("groups", strings.Join(groups, ",")).Msg("User details")

			c.Set("uid", username)
			c.Set("utype", "user")
			c.Set("groups", groups)
			c.Set("claims", claims)
			span.End()
			c.Next()
			return
		}

		bearerToken, err := RetrieveToken(c)
		if err != nil {
			c.AbortWithStatusJSON(errors.ErrorAuthMissingAuthHeader(err))
			return
		}

		// Check if it is static token
		static, err := IsStaticToken(c, ctx, bearerToken)
		if err != nil {
			c.AbortWithStatusJSON(errors.ErrorAuthInvalidToken(err, "issuer-selection"))
			return
		}
		if static {
			c.Next()
			return
		}

		// Continue with the OIDC
		if Config.Auth.OIDC.Enabled {
			if err := OIDCAuth(c, ctx, bearerToken); err != nil {
				c.AbortWithStatusJSON(errors.ErrorAuthOidcInit(err))
				return
			}
		}

		// Continue with JWT
		if Config.Auth.JWT.Enabled {
			if err := JWTAuth(c, ctx, bearerToken); err != nil {
				c.AbortWithStatusJSON(errors.ErrorAuthTokenVerificationFailed(err, "JWT"))
				return
			}
		}

		span.End()
		c.Next()
	}
}

func IsStaticToken(c *gin.Context, ctx context.Context, tokenString string) (bool, error) {
	_, span := Tracer.Start(ctx, "middlewares.Auth.TokenAuth")
	defer span.End()

	claims := models.NewClaimsMap()
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, _, err := parser.ParseUnverified(tokenString, &claims)
	if err != nil {
		return false, err
	}

	if claims.Issuer == "passage-server" {
		staticToken, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unsupported signing method")
			}
			return []byte(Config.SharedSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(errors.ErrorAuthInvalidToken(err, "passage-server-token"))
			return true, err
		}
		if !staticToken.Valid {
			c.AbortWithStatusJSON(errors.ErrorAuthInvalidToken(err, "passage-server-token"))
			return true, err
		}
		c.Set("uid", claims.ID)
		c.Set("utype", "token")

		return true, nil
	}
	return false, nil

}

// handleOIDCAuth verifies OIDC tokens using Dex's OIDC provider
func OIDCAuth(c *gin.Context, ctx context.Context, tokenString string) error {

	_, span := Tracer.Start(ctx, "middlewares.Auth.OIDCAuth")
	defer span.End()

	OidcProvider, err := oidc.NewProvider(ctx, Config.Auth.OIDC.IssuerURL)
	if err != nil {
		span.AddEvent("OIDC provider initialization failed")
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	// Verifier for the OIDC tokens
	idTokenVerifier := OidcProvider.Verifier(&oidc.Config{ClientID: Config.Auth.OIDC.ClientID})

	// Verify and parse the token
	idToken, err := idTokenVerifier.Verify(ctx, tokenString)
	if err != nil {
		span.AddEvent("OIDC token verification failed")
		return fmt.Errorf("failed to verify OIDC token: %w", err)
	}

	// Extract claims from the ID token
	var claims struct {
		Email string `json:"email"`
		*jwt.RegisteredClaims
	}
	if err := idToken.Claims(&claims); err != nil {
		span.AddEvent("Failed to parse OIDC token claims")
		return fmt.Errorf("failed to parse OIDC token claims: %w", err)
	}

	// Set user context information
	c.Set("uid", claims.Email)
	c.Set("utype", "user")

	span.SetAttributes(attribute.String("identity.id", claims.Email))
	c.Request = c.Request.WithContext(tracing.AddBaggage(c.Request.Context(), "identity.id", claims.Email))

	return nil

}

// handleJWTAuth verifies JWT tokens (e.g., Teleport's tokens)
func JWTAuth(c *gin.Context, ctx context.Context, tokenString string) error {

	_, span := Tracer.Start(ctx, "middlewares.Auth.JWTAuth")
	defer span.End()

	claims := models.NewClaimsMap()
	token, err := jwt.ParseWithClaims(tokenString, &claims, getJWT, jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	if claims.Issuer != Config.Auth.JWT.Issuer {
		return fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	// Extract specific claims dynamically
	username := claims.GetString(Config.Auth.JWT.UsernameClaim)
	groups := claims.GetStringSlice(Config.Auth.JWT.GroupsClaim)

	log.Debug().
		Str("username", username).
		Str("groups", strings.Join(groups, ",")).Msg("User details")

	// Set user context
	c.Set("uid", username)
	c.Set("utype", "user")
	c.Set("groups", groups)
	c.Set("claims", claims)

	span.SetAttributes(attribute.String("identity.id", username))

	c.Request = c.Request.WithContext(tracing.AddBaggage(c.Request.Context(), "identity.id", username))

	return nil
}

func getJWT(token *jwt.Token) (interface{}, error) {
	ctx := context.Background()
	// TODO: cache response so we don't have to make a request every time
	// we want to verify a JWT
	jwksURL := Config.Auth.JWT.JWKSURL

	set, err := jwk.Fetch(ctx, jwksURL)
	if err != nil {
		return nil, err
	}

	key, _ := set.Key(0)

	var pubKey interface{}

	err = key.Raw(&pubKey)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func RetrieveToken(c *gin.Context) (token string, err error) {

	// API client will always use authorization header with Bearer prefix
	headerName := "Authorization"
	headerPrefix := "Bearer "

	// First check if thats the case
	authHeader := c.GetHeader(headerName)
	if strings.HasPrefix(authHeader, headerPrefix) {

		token = authHeader[len(headerPrefix):]
		return token, nil
	}

	// Override header and prefix if specified
	if Config.Auth.JWT.Enabled {
		headerName = Config.Auth.JWT.TokenHeader
		headerPrefix = Config.Auth.JWT.HeaderPrefix
	}

	// Get auth header contents
	authHeader = c.GetHeader(headerName)

	// Check if content is not blank
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Check if header contains headerPrefix if prefix is specified
	if headerPrefix != "" {
		log.Debug().Msgf("Header prefix found: %s", headerPrefix)
		if !strings.HasPrefix(authHeader, headerPrefix) {
			return "", fmt.Errorf("invalid authorization header")
		}
		token = authHeader[len(headerPrefix):]
		return token, nil
	}

	return authHeader, nil
}
