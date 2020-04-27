package main

import (
    "net/http"
    "strings"
    )

type slackRequest struct {
    token       string
    teamId      string
    teamDomain  string
    channelId   string
    userId      string
    userName    string
    command     string
    text        string
    responseUrl string
}

func moveMessages(w http.ResponseWriter, r *http.Request) {
    message := r.URL.Path
    message = strings.TrimPrefix(message, "/")
    message = "Hello " + message

    w.Write([]byte(message))
}

func main() {
    http.HandleFunc("/command/move", moveMessages)
    if err := http.ListenAndServe(":443", nil); err != nil {
        panic(err)
    }
}

