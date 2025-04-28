package controller

import (
	"net/http"
	"time"

	"github.com/raikh/calc_micro_final/middleware"
	"github.com/raikh/calc_micro_final/model"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

type UserRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required,minstringlength(3)~Password too short,type(string)"`
}

type TokenResponse struct {
	Token     string     `json:"access_token"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func Register() echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := new(UserRequest)
		if err := c.Bind(requestUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		_, err := govalidator.ValidateStruct(requestUser)

		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, err.Error())
		}

		_, err = model.Create(requestUser.Email, requestUser.Password)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Unprocessable Entity")
		}

		return c.JSON(http.StatusCreated, nil)
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		requestUser := new(UserRequest)
		if err := c.Bind(requestUser); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		_, err := govalidator.ValidateStruct(requestUser)

		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		user, err := model.GetByEmail(requestUser.Email)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
		}

		if user.Id == 0 ||
			!user.CheckPasswordHarsh(requestUser.Password) {
			return echo.NewHTTPError(http.StatusForbidden, "Incorrect credentials")
		}

		token, time, err := middleware.GenerateAccessToken(&user)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, TokenResponse{token, &time})
	}
}
