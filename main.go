package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
)

// web controllers
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "Hello bedrock api proxy...")
}

func chat(w http.ResponseWriter, r *http.Request) {
	var chatReq openai.ChatCompletionRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(chatReq.Model)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	if chatReq.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		converseInput := formatStreamInput(chatReq)
		convereStream(brc, w, converseInput)

	} else {
		converseInput := format(chatReq)
		response := converse(brc, converseInput)
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}

}

func main() {
	fmt.Println("Server started on port 8081...")
	http.HandleFunc("/", home)
	http.HandleFunc("POST /v1/chat/completions", chat)
	http.ListenAndServe(":8081", nil)
}
