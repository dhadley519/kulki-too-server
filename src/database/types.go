package database

type EmailUser struct {
	Email string `form:"email"`
}

type PasswordUser struct {
	Password string `form:"password"`
}

type User struct {
	EmailUser
	PasswordUser
}
