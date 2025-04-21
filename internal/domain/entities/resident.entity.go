// package entities contém as definições de entidades do domínio, representando os dados e comportamentos principais do sistema.
package entities

type Resident struct {
	Apartamento string         `bson:"apartamento"`
	Resident    []ResidentInfo `bson:"moradores"`
}

type ResidentInfo struct {
	Nome     string `bson:"nome"`
	Telefone string `bson:"telefone"`
}
