package models

const (
	P_BASETYPE     = 0  // BaseType     string
	P_FRAMETYPE    = 1  // FrameType    int
	P_H            = 2  // H            int
	P_ICON         = 3  // Icon         string
	P_ID           = 4  // ID           string
	P_IDENTIFIED   = 5  // Identified   bool
	P_ILVL         = 6  // Ilvl         int
	P_INVENTORYID  = 7  // InventoryID  string
	P_LEAGUE       = 8  // League       string
	P_NAME         = 9  // Name         string
	P_TYPELINE     = 10 // TypeLine    string
	P_VERIFIED     = 11 // Verified    bool
	P_W            = 12 // W           int
	P_X            = 13 // X           int
	P_Y            = 14 // Y           int
	P_CORRUPTED    = 15 // Corrupted    bool
	P_TALISMANTIER = 16 // TalismanTier int
	P_DESCRTEXT    = 17 // DescrText    string
	P_NOTE         = 18 // Note         string

	P_EXPLICITMODS = 19 // ExplicitMods []string
	P_IMPLICITMODS = 20 // ImplicitMods []string
	P_FLAVOURTEXT  = 21 // FlavourText  []string
	P_UTILITYMODS  = 22 // UtilityMods  []string
	P_CRAFTEDMODS  = 23 // CraftedMods  []string
	P_ENCHANTMODS  = 24 // EnchantMods  []string

	P_PROPERTIES = 25 // Properties  []struct {
	// 	DisplayMode int
	// 	Name        string
	// 	Type        int
	// 	Values      [][]interface{}
	// }
	// Extended     struct {
	P_SUBBASETYPE   = 26 // BaseType      string
	P_CATEGORY      = 27 // Category      string
	P_PREFIXES      = 28 // Prefixes      int
	P_SUBCATEGORIES = 29 // Subcategories []string
	P_SUFFIXES      = 30 // Suffixes      int
	// }

	// Virtual properties (made up)
	P_RARITY = 31
)

const (
	FRAME_TYPE_NORMAL         = 0
	FRAME_TYPE_MAGIC          = 1
	FRAME_TYPE_RARE           = 2
	FRAME_TYPE_UNIQUE         = 3
	FRAME_TYPE_GEM            = 4
	FRAME_TYPE_CURRENCY       = 5
	FRAME_TYPE_DIVINATIONCARD = 6
	FRAME_TYPE_QUESTITEM      = 7
	FRAME_TYPE_PROPHECY       = 8
	FRAME_TYPE_RELIC          = 9
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
	case P_NOTE:
		return item.Note

	case P_RARITY:
		switch item.FrameType {
		case FRAME_TYPE_NORMAL:
			return "NORMAL"
		case FRAME_TYPE_MAGIC:
			return "MAGIC"
		case FRAME_TYPE_RARE:
			return "RARE"
		case FRAME_TYPE_UNIQUE:
			return "UNIQUE"
		default:
			return ""
		}

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
	case P_NOTE:
		return "Note"

	case P_RARITY:
		return "Rarity"

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
