package main

import (
	"exercicio/config"
	"exercicio/internal/domain"
	"exercicio/internal/handler"
	"exercicio/internal/repository"
	"exercicio/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db := config.NewDBConnection()

	defer db.Close()
	
	userRepo    := repository.NewUserRepository(db)
	frutaRepo 	:= repository.NewFrutaRepository(db)

	userService := service.NewUserService(userRepo)
	frutaService := service.NewFrutaService(frutaRepo)
	
	userHandler := handler.NewUserHandler(userService)
	frutaHandler := handler.NewFrutaHandler(frutaService)

	r := gin.Default()
	
	userHandler.RegisterRoutes(r)
	frutaHandler.RegisterRoutes(r)


	_ = domain.Fruta{}

	r.Run(":8080")
}
