package filters

import "drhyu.com/indexer/models"

type ItemPropery struct {
	properyID int
	// poperyType reflection.type
}

const (
	P_BASETYPE     = iota // BaseType     string
	P_FRAMETYPE    = iota // FrameType    int
	P_H            = iota // H            int
	P_ICON         = iota // Icon         string
	P_ID           = iota // ID           string
	P_IDENTIFIED   = iota // Identified   bool
	P_ILVL         = iota // Ilvl         int
	P_INVENTORYID  = iota // InventoryID  string
	P_LEAGUE       = iota // League       string
	P_NAME         = iota // Name         string
	P_TYPELINE     = iota // TypeLine    string
	P_VERIFIED     = iota // Verified    bool
	P_W            = iota // W           int
	P_X            = iota // X           int
	P_Y            = iota // Y           int
	P_CORRUPTED    = iota // Corrupted    bool
	P_TALISMANTIER = iota // TalismanTier int
	P_DESCRTEXT    = iota // DescrText    string

	P_EXPLICITMODS = iota // ExplicitMods []string
	P_IMPLICITMODS = iota // ImplicitMods []string
	P_FLAVOURTEXT  = iota // FlavourText  []string
	P_UTILITYMODS  = iota // UtilityMods  []string
	P_CRAFTEDMODS  = iota // CraftedMods  []string
	P_ENCHANTMODS  = iota // EnchantMods  []string

	P_PROPERTIES = iota // Properties  []struct {
	// 	DisplayMode int
	// 	Name        string
	// 	Type        int
	// 	Values      [][]interface{}
	// }
	// Extended     struct {
	P_SUBBASETYPE   = iota // BaseType      string
	P_CATEGORY      = iota // Category      string
	P_PREFIXES      = iota // Prefixes      int
	P_SUBCATEGORIES = iota // Subcategories []string
	P_SUFFIXES      = iota // Suffixes      int
	// }
)

func (filter *Filter) GetItemProperty(item *models.Item) interface{} {

	switch filter.PropertyID {

	// Simple types
	case P_BASETYPE:
		return item.BaseType
	case P_FRAMETYPE:
		return item.FrameType
	case P_H:
		return item.H
	case P_ICON:
		return item.Icon
	case P_ID:
		return item.ID
	case P_IDENTIFIED:
		return item.Identified
	case P_ILVL:
		return item.Ilvl
	case P_INVENTORYID:
		return item.InventoryID
	case P_LEAGUE:
		return item.League
	case P_NAME:
		return item.Name
	case P_TYPELINE:
		return item.TypeLine
	case P_VERIFIED:
		return item.Verified
	case P_W:
		return item.W
	case P_X:
		return item.X
	case P_Y:
		return item.Y
	case P_CORRUPTED:
		return item.Corrupted
	case P_TALISMANTIER:
		return item.TalismanTier
	case P_DESCRTEXT:
		return item.DescrText
	case P_SUBBASETYPE:
		return item.Extended.BaseType
	case P_CATEGORY:
		return item.Extended.Category
	case P_PREFIXES:
		return item.Extended.Prefixes
	case P_SUFFIXES:
		return item.Extended.Suffixes

	// Arrays
	case P_EXPLICITMODS:
		return &item.ExplicitMods
	case P_IMPLICITMODS:
		return &item.ImplicitMods
	case P_FLAVOURTEXT:
		return &item.FlavourText
	case P_UTILITYMODS:
		return &item.UtilityMods
	case P_CRAFTEDMODS:
		return &item.CraftedMods
	case P_ENCHANTMODS:
		return &item.CraftedMods
	case P_SUBCATEGORIES:
		return &item.Extended.Subcategories

	case P_PROPERTIES:
		// check if subproperty exists
		for i := range item.Properties {
			if item.Properties[i].Name == filter.SubProperty {
				if len(item.Properties[i].Values) > 0 && len(item.Properties[i].Values[0]) > 0 {
					return item.Properties[i].Values[0][0]
				} else {
					return nil
				}
			}
		}
		return nil
	}

	return nil
}
