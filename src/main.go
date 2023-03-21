package main

import (
	"crypto/rand"
	"math/big"
)

type EmailUser struct {
	Email string
}

type PasswordUser struct {
	Password string
}

type User struct {
	EmailUser
	PasswordUser
}

func main() {
	initPrivateKey()
	db = setupDb()
	web := setupWeb(db)
	defer db.Close()
	err := web.Run()
	if err != nil {
		panic(err)
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func randomInt(min int, max int) int {
	//rand.Int produces a number between 0 and n
	//we take max - min to get a random number range and then add min back to get a result between min and max.
	randomBigInteger, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		panic(err)
	}
	return min + int(randomBigInteger.Int64())
}
