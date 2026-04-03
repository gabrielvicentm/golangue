package repository

import (
	"context"
	"exercicio/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type frutaRepository struct {
	db *pgxpool.Pool
}


func NewFrutaRepository(db *pgxpool.Pool) domain.FrutaRepository {
	return &frutaRepository{db: db}
}

func (r *frutaRepository) Save(fruta domain.Fruta) error {
	_, err := r.db.Exec(
		
		context.Background(),
	
		`INSERT INTO frutas(id, name, color, price, weight_grams)
		 VALUES ($1, $2, $3, $4, $5)`,
		fruta.ID, fruta.Name, fruta.Color, fruta.Price, fruta.Weight_grams,
	)
	return err
}


func (r *frutaRepository) FindByName(name string) (*domain.Fruta, error) {
	var fruta domain.Fruta
	
	err := r.db.QueryRow(
		
		context.Background(),
		`SELECT id, name, color, price, weight_grams
		 FROM frutas
		 WHERE name = $1`,
		
		name,
	).Scan(&fruta.ID, &fruta.Name, &fruta.Color, &fruta.Price, &fruta.Weight_grams)
	
	if err != nil && err.Error() == "no rows in result set" {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &fruta, nil
	
}
