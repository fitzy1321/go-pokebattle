package store

import (
	"gorm.io/gorm"
)

func GetPokemon(db *gorm.DB) ([]Pokemon, error) {
	var pokemon []Pokemon
	result := db.Find(&pokemon)
	if result.Error != nil {
		return nil, result.Error
	}

	return pokemon, nil
}
