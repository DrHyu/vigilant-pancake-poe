package filters

import (
	"regexp"
	"strconv"

	"drhyu.com/indexer/models"
)

type Filter struct {
	Value1 interface{} `json:"Value1"`
	Value2 interface{} `json:"Value2,omitempty"`
	Value3 interface{} `json:"Value3,omitempty"`

	PropertyID  int    `json:"PropertyID,omitempty"`
	SubProperty string `json:"SubProperty,omitempty"`

	// How to compare the match of the regex
	RegexStr                   string         `json:"RegexStr,omitempty"`
	Regex                      *regexp.Regexp `json:"-"`
	RegexMatchComparisonMethod int            `json:"RegexMatchComparisonMethod,omitempty"`

	ComparisonMethod int  `json:"ComparisonMethod"`
	InverseMatch     bool `json:"InverseMatch"`
}

const (
	COMP_INT_EQ             = 0
	COMP_INT_GT             = 1
	COMP_INT_LT             = 2
	COMP_INT_GTE            = 3
	COMP_INT_LTE            = 4
	COMP_INT_BETWEEN        = 5
	COMP_STR_EQ             = 6
	COMP_REGEX_MATCH        = 7 // Check if regex matches the target field
	COMP_REGEX_FIND_COMPARE = 8 // Find a match in a string and compare the match vs another number. Regex -> Value1, CMP 1 -> Value2, CMP 2 -> Value3
)

func (filter *Filter) GetFilteredPropertyValue(item *models.Item) interface{} {
	return models.GetItemProperty(filter.PropertyID, filter.SubProperty, item)
}

func (filter *Filter) GetFilteredPropertyName(item *models.Item) string {
	return models.GetItemPropertyName(filter.PropertyID, filter.SubProperty, item)
}

type FilterError struct{}

func (f *FilterError) Error() string {
	return "Filter error !"
}

func (filter *Filter) ApplyTo(item *models.Item) (bool, error) {

	var result bool
	value := filter.GetFilteredPropertyValue(item)

	if value == nil {
		return false, nil
	}

	// Attempt string casts
	itemValStr, itemStrOk := value.(string)
	filterVal1Str, filter1StrOk := filter.Value1.(string)
	// filterVal2Str, filter2StrOk := filter.Value2.(string)
	// filterVal3Str, filter3StrOk := filter.Value3.(string)

	// Attempt int casts
	itemValInt, itemIntOk := value.(int)
	tempVal1, filter1IntOk := filter.Value1.(float64)
	filterVal1Int := int(tempVal1)
	tempVal2, filter2IntOk := filter.Value2.(float64)
	filterVal2Int := int(tempVal2)
	// filterVal3Int, filter3IntOk := filter.Value3.(int)

	switch filter.ComparisonMethod {

	case COMP_STR_EQ:
		result = itemStrOk && filter1StrOk && itemValStr == filterVal1Str
	case COMP_INT_EQ:
		result = itemIntOk && filter1IntOk && itemValInt == filterVal1Int
	case COMP_INT_GT:
		result = itemIntOk && filter1IntOk && itemValInt > filterVal1Int
	case COMP_INT_GTE:
		result = itemIntOk && filter1IntOk && itemValInt >= filterVal1Int
	case COMP_INT_LT:
		result = itemIntOk && filter1IntOk && itemValInt < filterVal1Int
	case COMP_INT_LTE:
		result = itemIntOk && filter1IntOk && itemValInt <= filterVal1Int
	case COMP_INT_BETWEEN:
		result = itemIntOk && filter1IntOk && filter2IntOk && itemValInt >= filterVal1Int && itemValInt <= filterVal2Int
	case COMP_REGEX_MATCH:
		if itemStrOk {
			result = filter.Regex.Match([]byte(itemValStr))
		} else {
			return false, nil
		}
	case COMP_REGEX_FIND_COMPARE:
		if !itemStrOk {
			return false, nil
		}
		match := filter.Regex.FindSubmatch([]byte(itemValStr))
		// Regex didn't match
		if match == nil {
			result = false
			break
		}

		foundInt, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return false, err
		}

		switch filter.RegexMatchComparisonMethod {
		case COMP_STR_EQ:
			return false, &FilterError{}
		case COMP_INT_EQ:
			result = filter1IntOk && foundInt == filterVal1Int
		case COMP_INT_GT:
			result = filter1IntOk && foundInt > filterVal1Int
		case COMP_INT_GTE:
			result = filter1IntOk && foundInt >= filterVal1Int
		case COMP_INT_LT:
			result = filter1IntOk && foundInt < filterVal1Int
		case COMP_INT_LTE:
			result = filter1IntOk && foundInt <= filterVal1Int
		case COMP_INT_BETWEEN:
			result = filter1IntOk && filter2IntOk && foundInt >= filterVal1Int && foundInt <= filterVal2Int
		default:
			return false, nil
		}

	default:
		return false, nil
	}

	if filter.InverseMatch {
		return !result, nil
	} else {
		return result, nil
	}
}

// func reportChanStatus(queues []chan *TrackedItem) {
// 	out := ""
// 	for i, c := range queues {

// 		out = out + fmt.Sprintf("CH%d %d/%d ", i, len(c), cap(c))
// 	}
// 	out += "\n"
// 	fmt.Print(out)
// }

// func debugChanLvlReporter(queues []chan *TrackedItem, inq chan<- models.Item, ouq chan *models.Item) {

// 	ticker := time.NewTicker(1000 * time.Millisecond)
// 	for {
// 		<-ticker.C
// 		reportChanStatus(queues)
// 		fmt.Printf("INQ %d/%d\n", len(inq), cap(inq))
// 		fmt.Printf("OUQ %d/%d\n", len(ouq), cap(ouq))
// 	}
// }
