package main

import (
	"github.com/genjidb/genji"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/types"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var db *genji.DB

func setupDb() *genji.DB {

	// Create a database instance, here we'll store everything in memory
	db, err := genji.Open(":memory:")
	if err != nil {
		log.Fatalln(err.Error())
		{
			panic(err)
		}
	}

	// Create a table. Genji tables are schemaless by default, you don't need to specify a schema.
	err = db.Exec("CREATE TABLE user (email, password)")
	if err != nil {
		log.Fatalln(err.Error())
		{
			panic(err)
		}
	}

	// Create an index.
	err = db.Exec("CREATE UNIQUE INDEX idx_email ON user (email)")
	if err != nil {
		log.Fatalln(err.Error())
		{
			panic(err)
		}
	}

	return db
}

func addUser(user User) error {

	passwordUser, hashError := hashPassword(user.PasswordUser)

	if hashError != nil {
		return hashError
	}

	// Insert some data
	insertError := db.Exec("INSERT INTO user (email, password) VALUES ('david.e.hadley@gmail.com',?)", passwordUser.Password)

	if insertError != nil {
		return insertError
	}
	return nil
}

func getUser(email EmailUser) *User {

	rows, queryError := db.Query("select email, password from user where email = ?", email.Email)
	if queryError != nil {
		if genji.IsNotFoundError(queryError) {
			return nil
		}
		panic(queryError)
	}

	var user User
	var userCount = 0
	iterError := rows.Iterate(func(d types.Document) error {
		if userCount > 0 {
			return errors.New("Found more than a single user for a given email address")
		}
		scanError := document.StructScan(d, &user)
		if scanError != nil {
			panic(scanError)
		}
		userCount++
		return nil
	})

	if iterError != nil || user.Email != email.Email {
		panic(iterError)
	}
	return &user
}

func hashPassword(password PasswordUser) (*PasswordUser, error) {
	passwordByteArray := []byte(password.Password)

	bcryptedPassword, hashError := bcrypt.GenerateFromPassword(passwordByteArray, bcrypt.DefaultCost)

	if hashError != nil {
		return nil, hashError
	}
	passStr := string(bcryptedPassword)
	return &PasswordUser{Password: passStr}, nil
}
