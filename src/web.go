package main

import (
	"github.com/genjidb/genji"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type ClaimsImpl struct {
	email *EmailUser
	jwt.RegisteredClaims
}

func setupWeb(db *genji.DB) *gin.Engine {
	var r *gin.Engine = gin.Default()
	r.POST("/register", register)
	r.POST("/login", login)
	return r
}

func login(c *gin.Context) {
	var user = User{EmailUser{Email: c.PostForm("email")}, PasswordUser{Password: c.PostForm("password")}}
	error := authUser(user)
	if error != nil {
		c.JSON(http.StatusUnauthorized, user.EmailUser)
	}
	refreshJwtCookie(c, user.Email)
	c.JSON(http.StatusOK, user.EmailUser)
}

func register(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	existingUser := getUser(user.EmailUser)
	if existingUser == nil {
		addUser(user)
		user.Password = ""
		c.JSON(http.StatusOK, user.EmailUser)
	} else {
		user.Password = ""
		c.JSON(http.StatusConflict, user.EmailUser)
	}
}

func authUser(user User) error {
	existingUser := getUser(user.EmailUser)
	error := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	return error
}

func refreshJwtCookie(ctx *gin.Context, email string) *gin.Context {
	user := getUser(EmailUser{Email: email})
	if user == nil {
		ctx.SetCookie("kulki", "", 0, "/", "localhost", false, true)
		return ctx
	}
	var token = jwt.NewWithClaims(jwt.SigningMethodRS256, ClaimsImpl{
		&user.EmailUser,
		jwt.RegisteredClaims{
			// Also fixed dates can be used for the NumericDate
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			Issuer:    "kulki",
		},
	})
	ss, err := token.SignedString(priv)
	if err != nil {
		panic(err)
	}
	ctx.SetCookie("kulki", ss, 3600, "/", "localhost", false, true)
	return ctx
}
