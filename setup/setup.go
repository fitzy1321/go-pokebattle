package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	BaseUrl          string = "https://pokeapi.co/api/v2"
	Gen1PokemonCount int    = 151
	SpriteUrlBase    string = "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-i/red-blue/transparent"
)

type dict = map[string]any

func GetSqliteDb(db_path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(db_path), &gorm.Config{})
}

func FetchDataAndCreateSqliteDb(db_path string) error {
	return CreateSqliteDb(FetchPokemonData(), db_path)
}

func FetchPokemonData() []fullPokeData {
	// WARN: buffered channel, don't change unless you know what you're doing (more than me 🙃).
	// WARN: concurrency gremlins will appear
	pokeDataChan := make(chan fullPokeData, Gen1PokemonCount)
	var wg sync.WaitGroup
	for i := range Gen1PokemonCount {
		pokeId := uint(i + 1)
		pokemon_url := fmt.Sprintf("%s/pokemon/%d", BaseUrl, pokeId)

		// * Where the ✨Magic✨ happens
		wg.Add(1)
		go func(url string, pokeId uint) {
			defer wg.Done()

			// * Top level PokeAPI pokemon object request
			pokemondata, err := fetchPokeAPIData(url)
			if err != nil {
				// TODO: is this the best way to handle this error?
				fmt.Fprintln(os.Stderr, err)
				return
			}

			// * Get Pokemon Type data, no additional requests
			type_1, type_2 := getPokemonTypes(pokemondata)

			stats := getStats(pokemondata)
			if stats == nil {
				stats = &statsData{}
			}

			// * Get moves, will perform additional network requests for detailed move data...
			mvData, err := getMovesData(pokemondata)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			// * Get Sprites PNGs (2 network requests, front and back PNGs)
			frontSprite, backSprite, err := getSprites(pokeId)

			// * Species Data PokeAPI request, network request
			var growth_rate *string = nil
			if speciesUrl, ok := pokemondata["species"].(dict)["url"].(string); ok {
				speciesData, spErr := fetchPokeAPIData(speciesUrl)
				if spErr != nil {
					fmt.Fprintln(os.Stderr, spErr)
					return
				}
				grstr := speciesData["growth_rate"].(dict)["name"].(string)
				growth_rate = &grstr
				// TODO: Evolution Data
			}

			pokeDataChan <- fullPokeData{
				Id:              pokeId,
				Name:            pokemondata["name"].(string),
				Type_1:          type_1,
				Type_2:          type_2,
				Base_experience: int(pokemondata["base_experience"].(float64)),
				Stats:           *stats,
				Moves:           mvData,
				Next_evolutions: []nextEvoData{}, // TODO: fix later
				Growth_Rate:     growth_rate,
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

	// * Get data out of channel
	fullAPIData := make([]fullPokeData, 0, len(pokeDataChan))
	for item := range pokeDataChan {
		fullAPIData = append(fullAPIData, item)
		fmt.Printf("Showing results I guess. %+v\n", item)
	}
	return fullAPIData
}

func CreateSqliteDb(data []fullPokeData, path string) error {
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

func fetchPokeAPIData(url string) (dict, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pokemondata dict
	if err := json.NewDecoder(resp.Body).Decode(&pokemondata); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return pokemondata, nil
}

func getSprites(pokeId uint) (ftSprite []byte, bkSprite []byte, err error) {
	err = nil
	frontUrl := fmt.Sprintf("%s/%d.png", SpriteUrlBase, pokeId)
	backUrl := fmt.Sprintf("%s/back/%d.png", SpriteUrlBase, pokeId)

	ftResp, ftRspErr := http.Get(frontUrl)
	if ftRspErr != nil {
		err = ftRspErr
	}
	defer ftResp.Body.Close()

	ftSprite, ftErr := spriteHandler(ftResp)
	if ftErr != nil {
		err = ftErr
	}

	bkResp, bkRspErr := http.Get(backUrl)
	if bkRspErr != nil {
		err = bkRspErr
	}
	defer bkResp.Body.Close()

	bkSprite, bkErr := spriteHandler(bkResp)
	if bkErr != nil {
		err = bkErr
	}
	return
}

func spriteHandler(resp *http.Response) ([]byte, error) {
	if resp.Header.Get("Content-Type") == "image/png" {
		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("Wrong Content-Type from network response.%v", resp.Header.Get("Content-Type"))
}

func getMovesData(pokeData dict) ([]moveData, error) {
	var rb_moves []_mvIR
	names := mapset.NewSet[string]()

	pokeMoves, ok := pokeData["moves"].([]any)
	if !ok {
		return nil, fmt.Errorf("No move data ...")
	}

	for _, pmMv := range pokeMoves {
		md := pmMv.(dict)
		vgdTop, ok := md["version_group_details"].([]any)
		if !ok {
			// TODO: error handle, idk man ...
		}
		for _, vgdIR := range vgdTop {
			vgd := vgdIR.(dict)
			if vgd["version_group"].(dict)["name"].(string) == "red-blue" {
				moveName := md["move"].(dict)["name"].(string)
				if !names.Contains(moveName) {
					names.Add(moveName)
					rb_moves = append(rb_moves, _mvIR{
						name:   moveName,
						level:  int(vgd["level_learned_at"].(float64)),
						url:    md["move"].(dict)["url"].(string),
						method: vgd["move_learn_method"].(dict)["name"].(string),
					})
				}
			}
		} // end for
	} // end for

	var detailed []moveData
	for _, move := range rb_moves {
		mvData, err := fetchPokeAPIData(move.url)
		if err != nil {
			// TODO: error handle, idk man ..
		}
		meta, ok := mvData["meta"].(dict)
		if !ok {
			// TODO: error handle idk man ...
		}

		// TODO: implement []statChange data
		// statChanges := []statChange{}

		var power *int = nil
		if tp, ok := mvData["power"].(int); ok {
			power = &tp
		}

		var acc *int = nil
		if tacc, ok := mvData["accuracy"].(int); ok {
			acc = &tacc
		}

		var mpp int = 0
		if tmpp, ok := mvData["pp"].(int); ok {
			mpp = tmpp
		}

		var mtype *string = nil
		if tmtype, ok := mvData["type"].(dict)["name"].(string); ok {
			mtype = &tmtype
		}

		var dc *string = nil
		if tdc, ok := mvData["damage_class"].(dict)["name"].(string); ok {
			dc = &tdc
		}
		var ailment *string = nil
		if tailment, ok := meta["ailment"].(dict)["name"].(string); ok {
			ailment = &tailment
		}

		var ailmentChance *int = nil
		if tailChnc, ok := meta["ailment_chance"].(int); ok {
			ailmentChance = &tailChnc
		}

		detailed = append(detailed, moveData{
			Name:           move.name,
			Level_learned:  uint(move.level),
			Learn_method:   move.method,
			Max_pp:         mpp,
			Power:          power,
			Accuracy:       acc,
			Type:           mtype,
			Damage_class:   dc,
			Ailment:        ailment,
			Ailment_chance: ailmentChance,
			// Move_category:  meta["category"].(dict)["name"].(string),
			// Healing:        meta["healing"].(int),
			// Drain:          meta["drain"].(int),
			Stat_changes: []statChange{}, // TODO: fix later
		})
	}
	return detailed, nil
}

func getPokemonTypes(data dict) (string, *string) {
	var type_1 string
	var type_2 *string
	for _, t := range data["types"].([]any) {
		tm := t.(dict)
		slot := int(tm["slot"].(float64))
		tmpVar := tm["type"].(dict)["name"].(string)
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
			// TODO: ? pass ?
		}
	}
	return type_1, type_2
}

func getStats(data dict) *statsData {
	stats := make(map[string]int)
	for _, v := range data["stats"].([]any) {
		tm := v.(dict)
		name := tm["stat"].(dict)["name"].(string)
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

// TODO: Keep or delete this func?
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
