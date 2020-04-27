package main

import (
    "net/http"
    "strings"
    "fmt"
    "github.com/nlopes/slack"
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

func moveMessagesHandler(w http.ResponseWriter, r *http.Request) {
    s, err := slack.SlashCommandParse(r)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    if !s.ValidateToken("905b8f07df16deb786fd568f9e016dd6") {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    switch s.Command {
    case "/command/move":
        params := &slack.Msg{Text: s.Text}
        response := fmt.Sprintf("You asked for the weather for %v", params.Text)
        w.Write([]byte(response))

    default:
        w.WriteHeader(http.StatusInternalServerError)
        return
    }


}

func main() {
    http.HandleFunc("/command/move", moveMessagesHandler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        panic(err)
    }
}

