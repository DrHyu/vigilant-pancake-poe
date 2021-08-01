package wsocket

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"drhyu.com/indexer/filters"
	"drhyu.com/indexer/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn

	requests chan []byte

	items  chan *models.Item
	search *filters.Search
}

func NewClient(c *websocket.Conn) *Client {

	return &Client{
		conn:     c,
		requests: make(chan []byte),
		// items:    make(chan *models.Item, 500),
	}
}

var mSearchGrps2 = []filters.SearchGroup{
	{
		SearchMode: filters.SEARCH_MODE_AND,
		Filters: []filters.Filter{
			{
				PropertyID:       models.P_NAME,
				Regex:            regexp.MustCompile("Tabula "),
				ComparisonMethod: filters.COMP_REGEX_MATCH,
			},
		},
	},
}

// type wsData struct {
// 	sGroups
// }

func decodeIncMsg(incMsg []byte) ([]filters.SearchGroup, bool) {

	var data []filters.SearchGroup
	err := json.Unmarshal(incMsg, &data)

	// var data2 []filters.SearchGroup
	// temp := strconv.Unquote(string(incMsg))

	if err != nil {
		return nil, false
	}

	ttt := `hello world \d\d`
	_ = ttt

	for i := range data {
		for y := range data[i].Filters {
			// If they passed a regex string, compile it now
			if data[i].Filters[y].RegexStr != "" {
				// decodedValue, err := url.QueryUnescape(data[i].Filters[y].RegexStr)
				// if err != nil {
				// 	return nil, false
				// }
				temp, err := regexp.Compile(data[i].Filters[y].RegexStr)
				if err != nil {
					return nil, false
				}
				data[i].Filters[y].Regex = temp
			}
		}
	}

	return data, true
}
func (c *Client) Handle(SM *filters.SearchManager) {

	go c.ProcessIncMsg()

	defer func() {
		if c.search != nil {
			SM.EndSearch(c.search.ID)
		}
		c.conn.Close()
	}()

	for {
		select {
		case newMsg, ok := <-c.requests:
			if !ok {
				return
			}
			fmt.Print("Got msg ", string(newMsg), "\n")
			searchGroups, ok := decodeIncMsg(newMsg)
			if !ok {
				fmt.Print("Malformed msg")
				continue
			}

			if c.search != nil {
				// End previous search for this client (if there was one)
				SM.EndSearch(c.search.ID)
				close(c.items)
			}

			c.items = make(chan *models.Item, 1000)
			c.search = filters.NewSearch(searchGroups, c.items)
			SM.NewSearch(c.search)
		case newItem := <-c.items:

			// Send more than 1 item in one go if available
			var lgth = len(c.items)
			var itemBundle = make([]models.Item, lgth+1)

			itemBundle[0] = *newItem
			for i := 0; i < lgth; i++ {
				temp := <-c.items
				itemBundle[i+1] = *temp
			}

			var byteStream, err = json.Marshal(itemBundle)

			if err != nil {
				fmt.Print(err)
				return
			}

			err = c.conn.WriteMessage(websocket.TextMessage, byteStream)

			if err != nil {
				fmt.Print(err)
				return
			}

		}
	}

}

func (c *Client) ProcessIncMsg() {

	defer func() {
		close(c.requests)
		c.conn.Close()
	}()

	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.requests <- message
	}
}
