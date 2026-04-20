package auth

import (
	"log"
	"net/http"
	"strings"

	"todo/shared/appcontext"
	userpkg "todo/user"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
)

func Auth(authClient *firebaseauth.Client, userUsecase userpkg.Usecase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Bearer token is required",
				})
			}

			token, err := authClient.VerifyIDToken(c.Request().Context(), tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			email, _ := token.Claims["email"].(string)
			name, _ := token.Claims["name"].(string)

			user, err := userUsecase.FindOrCreateByFirebaseUID(c.Request().Context(), token.UID, email, name)
			if err != nil {
				log.Printf("[ERROR] FindOrCreateByFirebaseUID failed: uid=%s, error=%v", token.UID, err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Internal server error",
				})
			}

			ctx := appcontext.SetUserID(c.Request().Context(), user.ID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
