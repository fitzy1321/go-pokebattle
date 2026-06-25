package setup

type fullPokeData struct {
	Id              uint
	Name            string
	Type_1          string
	Type_2          *string // nullable
	Base_experience int
	Stats           statsData
	Moves           []moveData
	Next_evolutions []nextEvoData
	Growth_Rate     *string // nullable
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

type moveData struct {
	Name           string
	Level_learned  uint
	Learn_method   string
	Max_pp         int
	Power          *int         // nullable
	Accuracy       *int         // nullable
	Type           *string      // TODO: should this be nullable?
	Damage_class   *string      // nullable
	Ailment        *string      // nullable
	Ailment_chance *int         // nullable
	Move_category  *string      // nullable
	Healing        *int         // nullable
	Drain          *int         // nullable
	Stat_changes   []statChange // TODO: maybe nullable?

}

type statChange struct {
	Stat   string
	Change any // TODO: check type
}

type nextEvoData struct {
	Evolves_into_id uint
	Trigger         string
	Min_level       uint
	Item            *string // nullable
}

type _mvIR struct {
	name   string
	level  int
	url    string
	method string
}
