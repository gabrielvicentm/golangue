CREATE TABLE frutas(
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    price FLOAT(10,2),
    weight_grams INT NOT NULL, 
    created_at TIMESTAPZ NOT NULL DEFAULT NOW()
)

