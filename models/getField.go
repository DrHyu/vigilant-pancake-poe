package models

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

func GetItemProperty(propertyID int, SubProperty string, item *Item) interface{} {

	switch propertyID {

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
			if item.Properties[i].Name == SubProperty {
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

func GetItemPropertyName(propertyID int, subProperty string, item *Item) string {

	switch propertyID {
	case P_BASETYPE:
		return "BaseType"
	case P_FRAMETYPE:
		return "FrameType"
	case P_H:
		return "H"
	case P_ICON:
		return "Icon"
	case P_ID:
		return "ID"
	case P_IDENTIFIED:
		return "Identified"
	case P_ILVL:
		return "Ilvl"
	case P_INVENTORYID:
		return "InventoryID"
	case P_LEAGUE:
		return "League"
	case P_NAME:
		return "Name"
	case P_TYPELINE:
		return "TypeLine"
	case P_VERIFIED:
		return "Verified"
	case P_W:
		return "W"
	case P_X:
		return "X"
	case P_Y:
		return "Y"
	case P_CORRUPTED:
		return "Corrupted"
	case P_TALISMANTIER:
		return "TalismanTier"
	case P_DESCRTEXT:
		return "DescrText"
	case P_SUBBASETYPE:
		return "Extended.BaseType"
	case P_CATEGORY:
		return "Extended.Category"
	case P_PREFIXES:
		return "Extended.Prefixes"
	case P_SUFFIXES:
		return "Extended.Suffixes"

	// Arrays
	case P_EXPLICITMODS:
		return "ExplicitMods"
	case P_IMPLICITMODS:
		return "ImplicitMods"
	case P_FLAVOURTEXT:
		return "FlavourText"
	case P_UTILITYMODS:
		return "UtilityMods"
	case P_CRAFTEDMODS:
		return "CraftedMods"
	case P_ENCHANTMODS:
		return "CraftedMods"
	case P_SUBCATEGORIES:
		return "Extended.Subcategories"

	case P_PROPERTIES:
		// check if subproperty exists
		return "Properties.[*]." + subProperty
	default:
		return ""
	}
}
