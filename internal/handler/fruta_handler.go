package handler

import (
 	"net/http"
	"exercicio/internal/domain"
	"exercicio/internal/service"

	"github.com/gin-gonic/gin"
)


type FrutaHandler struct {
	service *service.FrutaService
	}


func NewFrutaHandler(service *service.FrutaService) *FrutaHandler {
	return &FrutaHandler{service: service}
}


func (h *FrutaHandler) RegisterRoutes(r *gin.Engine) {
	frutas := r.Group("/frutas")
	
	{
		frutas.POST("/registraFrutinha", h.CreateFruta)
	}
}

func (h *FrutaHandler) CreateFruta(c *gin.Context) {

	
	var input domain.FrutaInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// PASSO 2: Chamar o service com o input já validado.
	fruta, err := h.service.CreateFruta(input)


	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// PASSO 3: Responder com sucesso.
	c.JSON(http.StatusCreated, fruta)
}
