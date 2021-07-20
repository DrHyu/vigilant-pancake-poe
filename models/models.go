package models

import "fmt"

//easyjson
type RespStruct struct {
	NextChangeID string  `json:"next_change_id"`
	Stashes      []Stash `json:"stashes"`
}

type Stash struct {
	AccountName       string `json:"accountName"`
	ID                string `json:"id"`
	Items             []Item `json:"items"`
	LastCharacterName string `json:"lastCharacterName"`
	League            string `json:"league"`
	Public            bool   `json:"public"`
	Stash             string `json:"stash"`
	StashType         string `json:"stashType"`
}

type Item struct {
	BaseType     string   `json:"baseType"`
	ExplicitMods []string `json:"explicitMods,omitempty"`
	FrameType    int      `json:"frameType"`
	H            int      `json:"h"`
	Icon         string   `json:"icon"`
	ID           string   `json:"id"`
	Identified   bool     `json:"identified"`
	Ilvl         int      `json:"ilvl"`
	ImplicitMods []string `json:"implicitMods,omitempty"`
	InventoryID  string   `json:"inventoryId"`
	League       string   `json:"league"`
	Name         string   `json:"name"`
	Extended     struct {
		BaseType      string   `json:"baseType"`
		Category      string   `json:"category"`
		Prefixes      int      `json:"prefixes"`
		Subcategories []string `json:"subcategories"`
		Suffixes      int      `json:"suffixes"`
	} `json:"extended,omitempty"`
	Requirements []struct {
		DisplayMode int             `json:"displayMode"`
		Name        string          `json:"name"`
		Values      [][]interface{} `json:"values"`
	} `json:"requirements,omitempty"`
	TypeLine    string   `json:"typeLine"`
	Verified    bool     `json:"verified"`
	W           int      `json:"w"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	FlavourText []string `json:"flavourText,omitempty"`
	Properties  []struct {
		DisplayMode int             `json:"displayMode"`
		Name        string          `json:"name"`
		Type        int             `json:"type"`
		Values      [][]interface{} `json:"values"`
	} `json:"properties,omitempty"`
	SocketedItems []interface{} `json:"socketedItems,omitempty"`
	Sockets       []struct {
		Attr    string `json:"attr"`
		Group   int    `json:"group"`
		SColour string `json:"sColour"`
	} `json:"sockets,omitempty"`
	Corrupted    bool     `json:"corrupted,omitempty"`
	TalismanTier int      `json:"talismanTier,omitempty"`
	DescrText    string   `json:"descrText,omitempty"`
	UtilityMods  []string `json:"utilityMods,omitempty"`
	CraftedMods  []string `json:"craftedMods,omitempty"`
	EnchantMods  []string `json:"enchantMods,omitempty"`
}

func (i *Item) Describe(fields ...int) string {

	out := "Item:\n"

	for _, field := range fields {

		prop := GetItemProperty(field, "", i)
		label := GetItemPropertyName(field, "", i)

		out += fmt.Sprintf("\t%v: %v\n", label, prop)
	}
	return out
}
