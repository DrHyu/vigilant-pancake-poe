package wsocket

import (
	"fmt"
	"net/http"

	"drhyu.com/indexer/filters"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(*http.Request) bool { return true },
}

func wshandler(w http.ResponseWriter, r *http.Request, SM *filters.SearchManager) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	var client = NewClient(conn)
	go client.Handle(SM)
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
