package entities

type Resident struct {
	Apartamento string         `bson:"apartamento"`
	Moradores   []ResidentInfo `bson:"moradores"`
}

type ResidentInfo struct {
	Nome     string `bson:"nome"`
	Telefone string `bson:"telefone"`
}
