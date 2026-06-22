package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const BASE_URL string = "https://pokeapi.co/api/v2"
const SPRITE_URL_BASE string = "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-i/red-blue/transparent"

const POKEMON_COUNT int = 151

func GetSqliteDb(db_path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(db_path), &gorm.Config{})
}

func FetchDataAndCreateSqliteDb(db_path string) error {
	return createSqliteDb(fetchFromPokeAPI(), db_path)
}

type fullPokeData struct {
	Id              uint
	Name            string
	Type_1          string
	Type_2          *string // nullable
	Base_experience uint
	Stats           statsData
	Moves           []movesData
	Next_evolutions []nextEvoData
	Growth_Rate     string
	Front_sprite    []byte
	Back_sprite     []byte
}

type statsData struct {
	Attack          int
	Defense         int
	Hp              int
	Special_Attack  int
	Special_Defense int
	Speed           int
}

type movesData struct {
}

type nextEvoData struct {
	Evolves_into_id uint
	Trigger         string
	Min_level       uint
	Item            *string // nullable
}

func fetchFromPokeAPI() []fullPokeData {
	// WARN: buffered channel, don't change unless you know what you're doing (more than me 🙃).
	// WARN: concurrency gremlins will appear
	pokeDataChan := make(chan fullPokeData, POKEMON_COUNT)
	var wg sync.WaitGroup
	for i := range POKEMON_COUNT {
		pokeId := uint(i + 1)
		pokemon_url := fmt.Sprintf("%s/pokemon/%d", BASE_URL, pokeId)

		// * Where the ✨Magic✨ happens
		wg.Add(1)
		go func(url string, pokeId uint) {
			defer wg.Done()

			// * First PokeAPI requesr
			pokemondata, err := fetchPokeAPI(url)
			if err != nil {
				// TODO: is this the best way to handle this error?
				fmt.Fprintln(os.Stderr, err)
				return
			}

			// TODO: Move Data
			// TODO: Collect/Fetch Species and Next Evolution Data

			type_1, type_2 := getTypes(pokemondata)
			frontSprite, backSprite, err := getSprites(pokeId)

			pokeDataChan <- fullPokeData{
				Id:              pokeId,
				Name:            pokemondata["name"].(string),
				Type_1:          type_1,
				Type_2:          type_2,
				Base_experience: uint(pokemondata["base_experience"].(float64)),
				Stats:           *getStats(pokemondata),
				Moves:           []movesData{},   // TODO
				Next_evolutions: []nextEvoData{}, // TODO
				Growth_Rate:     "",              // TODO
				Front_sprite:    frontSprite,
				Back_sprite:     backSprite,
			}
			// end go func
		}(pokemon_url, pokeId)
		// end forloop
	}

	// * Wait for all goroutines and close the channel
	wg.Wait()
	// this may not get called if the buffered channel is changed, btw
	close(pokeDataChan)

	// * Allocate memory for our slice
	fullAPIData := make([]fullPokeData, 0, len(pokeDataChan))

	// * Get data out of channel
	for item := range pokeDataChan {
		fullAPIData = append(fullAPIData, item)
		fmt.Printf("Showing results I guess. %v\n", item)
	}

	return fullAPIData
}

func createSqliteDb(data []fullPokeData, path string) error {
	_, err := GetSqliteDb(path)
	if err != nil {
		return err
	}

	// TODO: Create Table Schema
	// TODO: Make sure foreign Keys are on for sqlite
	// TODO: ETL go struct to sql inserts
	// TODO: commit, cleanup, exit

	return nil
}

// func osLevelStuff() error {
// 	home_path, ok := os.LookupEnv("HOME")
// 	if !ok {
// 		return fmt.Errorf("No Home ENV, something is wrong ...\n")

// 	}
// 	fmt.Println("Home path:", home_path)

// 	xdg_data := os.Getenv("XDG_DATA_HOME")
// 	fmt.Println("idk if this is real? :", xdg_data)

// 	xdg_config := os.Getenv("XDG_CONFIG_HOME")
// 	fmt.Println("XDG_CONFIG_HOME:", xdg_config)

// 	osname := runtime.GOOS
// 	switch osname {
// 	case "windows":
// 		fmt.Println("Windows specific stuff")
// 	case "darwin":
// 		fmt.Println("MacOS stuff")
// 	case "linux":
// 		fmt.Println("linux stuff")
// 	default:
// 		fmt.Println("I have no idea what you're on ...")
// 	}

// 	return nil
// }

func _spriteHandler(resp *http.Response) ([]byte, error) {
	if resp.Header.Get("Content-Type") == "image/png" {
		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("Wrong Content-Type from network response.%v", resp.Header.Get("Content-Type"))
}

func getSprites(pokeId uint) (ftSprite []byte, bkSprite []byte, err error) {
	err = nil

	url := fmt.Sprintf("%s/%d.png", SPRITE_URL_BASE, pokeId)
	ftResp, ftRspErr := http.Get(url)
	if ftRspErr != nil {
		err = ftRspErr
	}
	defer ftResp.Body.Close()

	ftSprite, ftErr := _spriteHandler(ftResp)
	if ftErr != nil {
		err = ftErr
	}

	url = fmt.Sprintf("%s/back/%d.png", SPRITE_URL_BASE, pokeId)
	bkResp, bkRspErr := http.Get(url)
	if bkRspErr != nil {
		err = bkRspErr
	}
	defer bkResp.Body.Close()

	bkSprite, bkErr := _spriteHandler(bkResp)
	if bkErr != nil {
		err = bkErr
	}
	return
}

func getStats(data map[string]any) *statsData {
	stats := make(map[string]int)
	for _, v := range data["stats"].([]any) {
		tm := v.(map[string]any)
		name := tm["stat"].(map[string]any)["name"].(string)
		baseStat := int(tm["base_stat"].(float64))
		stats[name] = baseStat
	}
	return &statsData{
		Attack:          stats["attack"],
		Defense:         stats["defense"],
		Hp:              stats["hp"],
		Special_Attack:  stats["special-attack"],
		Special_Defense: stats["special-defense"],
		Speed:           stats["speed"],
	}
}

func fetchPokeAPI(url string) (map[string]any, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pokemondata map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&pokemondata); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pokemondata, nil
}

func getTypes(data map[string]any) (string, *string) {
	var type_1 string
	var type_2 *string
	for _, t := range data["types"].([]any) {
		tm := t.(map[string]any)
		slot := int(tm["slot"].(float64))
		tmpVar := tm["type"].(map[string]any)["name"].(string)
		var name *string = nil
		if tmpVar != "" {
			name = &tmpVar
		}

		switch slot {
		case 1:
			// type_1 should always be there, so normal string works
			type_1 = *name
		case 2:
			// type_2 can be null, so this should be a pointer
			type_2 = name
		default:
			// pass
		}
	}
	return type_1, type_2
}
