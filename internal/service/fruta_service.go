package service

import (
	"errors"
	"exercicio/internal/domain"

	"github.com/google/uuid"
)

type FrutaService struct {			
	repo domain.FrutaRepository
}


func NewFrutaService(repo domain.FrutaRepository) *FrutaService {
	return &FrutaService{repo: repo}
	
}


func (s *FrutaService) CreateFruta(input domain.FrutaInput) (*domain.Fruta, error) {

	// REGRA 1: nomezinho da fruta não pode ser duplicado.
	existing, err := s.repo.FindByName(input.Name)
	
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("Fruta já cadastrada paizão")
	}

	fruta := domain.Fruta{
		ID:           uuid.NewString(),
		Name:         input.Name,
		Color: 		  input.Color,
		Price:  	  input.Price,
		Weight_grams: input.Weight_grams
	}

	if err := s.repo.Save(fruta); err != nil {
		return nil, err
	}

	return &fruta, nil
}
