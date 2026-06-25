package pokedata

// WARN: gorm struct, do not change the member names
type Pokemon struct {
	Id              uint `gorm:"primaryKey"`
	Name            string
	Type_1          string
	Type_2          *string
	Base_hp         int
	Base_attack     int
	Base_defense    int
	Base_sp_attack  int
	Base_sp_defense int
	Base_speed      int
	Base_experience *int
	Growth_rate     *string
	Front_sprite    []byte
	Back_sprite     []byte
}

func (Pokemon) TableName() string {
	return "dex_pokemon"
}

// WARN: gorm struct, do not change the member names
type Move struct {
	Id             uint `gorm:"primaryKey"`
	Name           string
	Power          *int
	Accuracy       *int
	Max_pp         int
	Type           *string
	Damage_class   *string
	Ailment        *string
	Ailment_chance *int
	Move_category  *string
	Healing        *int
	Drain          *int
}

func (Move) TableName() string {
	return "dex_move"
}
