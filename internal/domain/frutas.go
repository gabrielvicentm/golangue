package domain

type Fruta struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Color string `json:"color"`
	Price float64 `json:"price"`
	Weight_grams int `json:"weight_grams"`
}


type FrutaInput struct{
	Name string `json:"name"  binding:"required"`
	Color string `json:"color"  binding:"required"`
	Price float64 `json:"price" binding:"required,min=1"`
	Weight_grams int `json:"weight_grams" binding:"required,min=1"`
}

type FrutaRepository interface{
	Save(fruta Fruta) error
	FindByName(name string) (*Fruta, error)
}