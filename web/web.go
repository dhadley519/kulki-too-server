package web

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"kulki/database"
	"kulki/game"
	"net/http"
	"time"
)

type KulkiWebServer struct {
	engine *gin.Engine
	db     *database.KulkiDatabase
}

func (k *KulkiWebServer) Run() error {
	return k.engine.Run()
}

func SetupWeb(db *database.KulkiDatabase) *KulkiWebServer {
	var router = gin.Default()

	var config = cors.Config{}
	config.AllowOrigins = []string{"http://localhost:3030"}
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}

	w := &KulkiWebServer{engine: router, db: db}

	router.Use(cors.New(config))
	router.POST("/register", w.register)
	router.POST("/login", w.login)
	router.POST("/hover", w.hover)
	router.POST("/move", w.move)
	router.GET("/statistics", w.statistics)
	router.GET("/start", w.start)
	router.GET("/reset", w.reset)
	router.StaticFile("/index.html", "./public/index.html")
	router.Static("/static", "./public")

	return w
}

func (k *KulkiWebServer) reset(ctx *gin.Context) {
	emailUser, err := k.getEmailAddressFromCookie(ctx)
	if err == nil && &emailUser != nil {
		dbError := k.db.DeleteBoard(emailUser)
		if dbError == nil {
			ctx.Status(http.StatusOK)
			return
		}
	}
	ctx.Status(http.StatusInternalServerError)
}

func (k *KulkiWebServer) start(ctx *gin.Context) {
	emailUser, err := k.getEmailAddressFromCookie(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if emailUser != nil {
		var board *game.Board
		var err error
		var commands []*game.BoardCommand
		board, commands, err = k.db.GetBoard(emailUser)
		if err != nil || board == nil {
			board = game.NewBoard(9, 9, 6)
			commands = board.OnStartMessageReceipt()
			setBoardErr := k.db.SetBoard(emailUser, board)
			if setBoardErr != nil {
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			ctx.JSON(http.StatusOK, commands)
			return
		}
		ctx.JSON(http.StatusOK, commands)
		return
	}
	ctx.JSON(http.StatusInternalServerError, nil)
}

type Statistic struct {
	EmptyTileCount int `json:"emptyTileCount"`
	Score          int `json:"score"`
}

func (k *KulkiWebServer) statistics(ctx *gin.Context) {
	emailUser, err := k.getEmailAddressFromCookie(ctx)
	if err == nil {
		b, _, err := k.getUsersBoard(emailUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		if b != nil {
			ctx.JSON(http.StatusOK, Statistic{
				EmptyTileCount: b.EmptyTileCount(),
				Score:          b.Score})
			return
		}
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
}

func (k *KulkiWebServer) move(ctx *gin.Context) {
	var move game.MovePath
	bindError := ctx.BindJSON(&move)
	err := bindError
	if bindError == nil {
		emailUser, getEmailUserError := k.getEmailAddressFromCookie(ctx)
		err = getEmailUserError
		if getEmailUserError == nil && emailUser != nil {
			b, _, getBoardError := k.getUsersBoard(emailUser)
			err = bindError
			if getBoardError == nil && b != nil {
				commands, success := b.OnMoveMessageReceipt(&move)
				if success {
					updateBoardError := k.db.UpdateBoard(emailUser, b)
					err = updateBoardError
					if updateBoardError == nil {
						println("move success")
						ctx.JSON(http.StatusOK, commands)
						return
					}
				}
			}
		}
	}
	ctx.JSON(http.StatusInternalServerError, err)
}

type HoverResponse struct {
	Success   bool          `json:"success"`
	FinalPath []*game.Point `json:"finalPath"`
	game.MovePath
}

func (k *KulkiWebServer) hover(ctx *gin.Context) {
	var move game.MovePath
	bindError := ctx.BindJSON(&move)
	if bindError == nil {
		emailUser, getEmailAddressFromCookieError := k.getEmailAddressFromCookie(ctx)
		if getEmailAddressFromCookieError == nil {
			b, _, getUsersBoardError := k.getUsersBoard(emailUser)
			if getUsersBoardError == nil && b != nil {
				path, success := b.OnPathFindReceipt(&move)
				if success {
					var points []*game.Point
					for _, p := range path {
						points = append(points, &game.Point{X: p.X, Y: p.Y})
					}
					ctx.JSON(http.StatusOK, HoverResponse{Success: true, FinalPath: points, MovePath: move})
					return
				} else {
					ctx.JSON(http.StatusOK, HoverResponse{Success: false, FinalPath: []*game.Point{}, MovePath: move})
				}
			}
		}
		ctx.JSON(http.StatusInternalServerError, nil)
	}
}

func (k *KulkiWebServer) getUsersBoard(emailUser *database.EmailUser) (*game.Board, []*game.BoardCommand, error) {
	if emailUser != nil {
		return k.db.GetBoard(emailUser)
	}
	return nil, nil, nil
}

func (k *KulkiWebServer) getEmailAddressFromCookie(ctx *gin.Context) (*database.EmailUser, error) {
	cookie, err := ctx.Cookie("kulki")
	if err == nil {
		token, parseErr := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return pub, nil
		})
		if parseErr == nil {
			claims, ok := token.Claims.(jwt.MapClaims)
			if ok && token.Valid {
				email := claims["email"].(string)
				user, parseErr := k.db.GetUser(email)
				if parseErr == nil {
					return &user.EmailUser, nil
				}
			}
			return nil, nil
		}
		return nil, parseErr
	}
	return nil, err
}

func (k *KulkiWebServer) login(ctx *gin.Context) {

	var user database.User
	bindError := ctx.BindJSON(&user)
	if bindError != nil {
		ctx.JSON(http.StatusBadRequest, nil)
	}

	//var user = database.User{EmailUser: database.EmailUser{Email: ctx.PostForm("email")}, PasswordUser: database.PasswordUser{Password: ctx.PostForm("password")}}
	err := k.authUser(user)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, user.EmailUser)
		return
	}
	ctx, err = k.refreshJwtCookie(ctx, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, user.EmailUser)
}

type formData struct {
	email    string
	password string
}

func (k *KulkiWebServer) register(ctx *gin.Context) {
	//var f = formData{email: ctx.PostForm("email"), password: ctx.PostForm("password")}
	//err := ctx.Bind(&f)
	//if err != nil {
	//	ctx.JSON(http.StatusInternalServerError, nil)
	//	return
	//}
	var postedUser database.User
	bindError := ctx.BindJSON(&postedUser)
	if bindError != nil {
		ctx.JSON(http.StatusBadRequest, nil)
	}

	hashedPasswordUser, err := hashPassword(database.PasswordUser{Password: postedUser.Password})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	var user = database.User{EmailUser: database.EmailUser{Email: postedUser.Email}, PasswordUser: *hashedPasswordUser}
	err = k.db.AddUser(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx, err = k.refreshJwtCookie(ctx, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	user.Password = ""
	ctx.JSON(http.StatusOK, user.EmailUser)
	return
}

func hashPassword(password database.PasswordUser) (*database.PasswordUser, error) {
	passwordByteArray := []byte(password.Password)

	bcryptedPassword, hashError := bcrypt.GenerateFromPassword(passwordByteArray, bcrypt.DefaultCost)

	if hashError != nil {
		return nil, hashError
	}
	passStr := string(bcryptedPassword)
	return &database.PasswordUser{Password: passStr}, nil
}

func (k *KulkiWebServer) authUser(user database.User) error {
	existingUser, err := k.db.GetUser(user.Email)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	return err
}

func (k *KulkiWebServer) refreshJwtCookie(ctx *gin.Context, email string) (*gin.Context, error) {
	user, err := k.db.GetUser(email)
	if err != nil {
		ctx.SetCookie("kulki", "", 0, "/", "127.0.0.1", false, false)
		return ctx, err
	}
	if user == nil {
		ctx.SetCookie("kulki", "", 0, "/", "127.0.0.1", false, false)
		return ctx, errors.New("user not found")
	}
	var token = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"email":     user.Email,
		"ExpiresAt": jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		"Issuer":    "kulki"})
	ss, err := token.SignedString(priv)
	if err != nil {
		ctx.SetCookie("kulki", "", 0, "/", "127.0.0.1", false, false)
		return ctx, err
	}
	ctx.SetCookie("kulki", ss, 9999999999, "/", "127.0.0.1", false, false)
	return ctx, nil
}
