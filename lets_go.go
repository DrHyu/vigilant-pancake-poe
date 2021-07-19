package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"drhyu.com/indexer/tradeApi"
)

func read_unstructured_json(r io.Reader) ([]byte, error) {
	var result interface{}
	json.NewDecoder(r).Decode(&result)

	return json.MarshalIndent(result, "", "\t")
}

func read_structured_json(r io.Reader) ([]byte, error) {
	var result tradeApi.JsonStruct

	json.NewDecoder(r).Decode(&result)

	return json.MarshalIndent(result, "", "\t")
}

func main() {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://www.pathofexile.com/api/public-stash-tabs?id=1224902497-1229569941-1187442478-1328449025-1276721127", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, _ := client.Do(req)
	structed, _ := read_structured_json(resp.Body)

	defer resp.Body.Close()

	f2, _ := os.Create("structed.txt")
	defer f2.Close()
	f2.Write(structed)

}
