package database

type EmailUser struct {
	Email string `json:"email"`
}

type PasswordUser struct {
	Password string `json:"password"`
}

type User struct {
	EmailUser
	PasswordUser
}
