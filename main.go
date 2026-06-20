package main

import (
	"log"
	"os"
	"runtime"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Pokemon struct {
	Id              uint `gorm:"primaryKey"`
	Name            string
	Type_1          string
	Type_2          *string
	Base_hp         uint
	Base_attack     uint
	Base_defense    uint
	Base_sp_attack  uint
	Base_sp_defense uint
	Base_speed      uint
	Base_experience *uint
	Growth_rate     *string
	Front_sprite    []byte
	Back_sprite     []byte
}

func (Pokemon) TableName() string {
	return "dex_pokemon"
}

func osLevelStuff() {
	xdg_data := os.Getenv("XDG_DATA_HOME")
	log.Println("idk if this is real? : ", xdg_data)

	xdg_config := os.Getenv("XDG_CONFIG_HOME")
	log.Println("XDG_CONFIG_HOME:", xdg_config)

	osname := runtime.GOOS
	switch osname {
	case "windows":
		log.Println("Windows specific stuff")
	case "darwin":
		log.Println("MacOS stuff")
	case "linux":
		log.Println("linux stuff")
	default:
		log.Println("I have no idea what you're on ...")
	}
}

func main() {
	// open gorm and sqlite db
	db, err := gorm.Open(sqlite.Open("pokedata.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to pokemon database: %v", err)
	}

	// get all pokemon
	var pokedex []Pokemon
	result := db.Find(&pokedex)
	if result.Error != nil {
		log.Fatalf("Error getting pokemon data:%v", result.Error)
	}

	// print pokemon
	for _, k := range pokedex {
		log.Printf("Id: %d Name: %s Type: %s", k.Id, k.Name, k.Type_1)
	}
}
