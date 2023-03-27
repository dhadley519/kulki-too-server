package main

import (
	"awesomeProject/database"
	"awesomeProject/web"
)

func main() {
	web.InitPrivateKey()
	web.InitPublicKey()
	k := database.SetupDb()
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
