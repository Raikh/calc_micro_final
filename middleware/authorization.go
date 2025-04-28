package middleware

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/raikh/calc_micro_final/internal/config"
	"github.com/raikh/calc_micro_final/model"
)

type jwtCustomClaims struct {
	Id    int64 `json:"id"`
	Admin bool  `json:"admin"`
	jwt.RegisteredClaims
}

var secret_key, refresh_secret_key string
var expirationHours time.Duration

func Init(cfg *config.Config) {
	secret_key = cfg.GetKey("APP_JWT_SECRET_KEY")
	refresh_secret_key = cfg.GetKey("APP_JWT_REFRESH_SECRET_KEY")
	value, err := strconv.ParseUint(cfg.GetKey("APP_JWT_EXPIRATION_HOURS"), 10, 64)
	if err != nil {
		log.Fatalf("Value for APP_JWT_EXPIRATION_HOURS is unsupported")
	}
	expirationHours = time.Duration(value)
}

func JwtAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authHeader := ctx.Request().Header.Get("Authorization")

		if len(authHeader) == 0 {
			return ctx.JSON(http.StatusUnauthorized, "")
		}

		authFields := strings.Fields(authHeader)
		if len(authFields) != 2 || strings.ToLower(authFields[0]) != "bearer" {
			return errors.New("bad authorization header")
		}
		tokenString := authFields[1]

		token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret_key), nil
		})

		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, "")
		}

		if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
			if time.Now().Unix() > claims.ExpiresAt.Unix() {
				return ctx.JSON(http.StatusUnauthorized, "")
			}

			user, err := model.GetById(claims.Id)

			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, "")
			}

			if user.Email == "" {
				return ctx.JSON(http.StatusUnauthorized, "")
			}
			ctx.Set("user", user)
		} else {
			return ctx.JSON(http.StatusUnauthorized, "")
		}

		return next(ctx)
	}
}

func GenerateAccessToken(user *model.User) (string, time.Time, error) {
	// Declare the expiration time of the token (1h).
	expirationTime := time.Now().Add(time.Hour * expirationHours)

	return generateToken(user, expirationTime)
}

// Pay attention to this function. It holds the main JWT token generation logic.
func generateToken(user *model.User, expirationTime time.Time) (string, time.Time, error) {
	claims := &jwtCustomClaims{
		Id:    user.Id,
		Admin: false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret_key))
	if err != nil {
		return "", time.Now(), err
	}

	return tokenString, expirationTime, nil
}
