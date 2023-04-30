package main

import (
	"kulki/database"
	"kulki/web"
	"os"
)

func main() {

	web.InitPrivateKey()
	web.InitPublicKey()
	k := database.SetupDb(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"))
	defer func(k *database.KulkiDatabase) {
		err := k.Close()
		if err != nil {
			return
		}
	}(k)
	var webServer = web.SetupWeb(k)
	err := webServer.Run()
	if err != nil {
		return
	}
}
