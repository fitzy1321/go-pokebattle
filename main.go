package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"go-pokebattle/dex"
	"go-pokebattle/setup"

	"gorm.io/gorm"
)

func dbPathExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func main() {
	// * Catch all panics!
	defer func() {
		r := recover()
		if r == nil {
			return
		}
	}()

	// * Get and or Create Gorm/Sqlite DB
	dbPath := "pokedata.db"
	var db *gorm.DB = nil
	if exists, err := dbPathExists(dbPath); err != nil {
		fmt.Fprintln(os.Stderr, "Error occured checking for sqlite file:", err)
		return
	} else if !exists {
		// * Fetch Data From PokeAPI, Create SQLite DB, seeded with API Data
		data := setup.FetchPokemonData()
		fmt.Println("Length of pokemon data from api:", len(data))
		db, err = setup.CreateSqliteDb(data, dbPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Something wrong creating the db:", err)
			return
		}
		// * Wait for terminal input
		fmt.Print("> ")
		fmt.Scanln()
	} else {
		db, err = setup.GetSqliteDb(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to pokemon database: %v\n", err)
			return
		}
	}

	// * Get all Pokemon from db
	var pokedex []dex.Pokemon
	result := db.Find(&pokedex)
	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "Error getting pokemon data: %v\n", result.Error)
		return
	}

	// * Print Pokemons
	for _, k := range pokedex {
		fmt.Printf("Pokemon Id: %d Name: %s Type: %s\n", k.ID, k.Name, k.Type1)
	}

	// // * Get all moves from db
	// var movedex []Move
	// result = db.Find(&movedex)
	// if result.Error != nil {
	// 	fmt.Fprintf(os.Stderr, "Error getting move data: %v\n", result.Error)
	// 	return
	// }

	// // * Print moves
	// for _, k := range movedex {
	// 	fmt.Printf("Move id: %d, Name: %s\n", k.Id, k.Name)
	// }
}
