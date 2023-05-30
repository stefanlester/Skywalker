package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

func doAuth() error {
	// migrations
	dbType := sky.DB.DataType
	fileName := fmt.Sprintf("%d_create_auth_tables", time.Now().UnixMicro())
	upFile := sky.RootPath + "/migrations/" + fileName + ".up.sql"
	downFile := sky.RootPath + "/migrations/" + fileName + ".down.sql"

	err := copyFilefromTemplate("templates/migrations/auth_tables." + dbType + ".sql", upFile)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte("drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"), downFile)
	if err != nil {
		exitGracefully(err)
	}

	// run migrations
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/data/user.go.txt", sky.RootPath + "/data/user.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/data/token.go.txt", sky.RootPath + "/data/token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/data/remember_token.go.txt", sky.RootPath + "/data/remember_token.go")
	if err != nil {
		exitGracefully(err)
	}

	// copy over middleware
	err = copyFilefromTemplate("templates/middleware/auth.go.txt", sky.RootPath + "/middleware/auth.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/middleware/auth-token.go.txt", sky.RootPath + "/middleware/auth-token.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/middleware/remember.go.txt", sky.RootPath + "/middleware/remember.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/handlers/auth-handlers.go.txt", sky.RootPath + "/handlers/auth-handlers.go")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/mailer/password-reset.html.tmpl", sky.RootPath + "/mail/password-reset.html.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/mailer/password-reset.plain.tmpl", sky.RootPath + "/mail/password-reset.plain.tmpl")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/views/login.jet", sky.RootPath + "/views/login.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/views/forgot.jet", sky.RootPath + "/views/forgot.jet")
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilefromTemplate("templates/views/reset-password.jet", sky.RootPath + "/views/reset-password.jet")
	if err != nil {
		exitGracefully(err)
	}


	color.Yellow("  - users, tokens, and remember_tokens migrations created and executed")
	color.Yellow("  - user and token models created")
	color.Yellow("  - auth middleware created")
	color.Yellow("")
	color.Yellow("Don't forget to add user and token models in data/models.go, and to add appropriate middleware to your routes!")

	return nil
}