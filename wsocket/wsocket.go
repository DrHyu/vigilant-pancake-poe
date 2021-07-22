package wsocket

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"drhyu.com/indexer/filters"
	"drhyu.com/indexer/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wshandler(w http.ResponseWriter, r *http.Request, SM *filters.SearchManager) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	connHandler(conn, SM)

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

func connHandler(conn *websocket.Conn, SM *filters.SearchManager) {

	var foundItems = make(chan *models.Item, 1000)
	var search = filters.NewSearch(mSearchGrps2, foundItems)
	SM.NewSearch(search)

	for {
		item := <-foundItems
		err := conn.WriteMessage(websocket.TextMessage, []byte(item.Name+" -- "+strconv.Itoa(int(search.ID))))

		if err != nil {
			fmt.Print(err)
			break
		}
	}
	SM.EndSearch(search.ID)
}

func StartWS(SM *filters.SearchManager) {
	r := gin.Default()
	r.LoadHTMLFiles("index.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request, SM)
	})

	r.Run("localhost:12312")
}
