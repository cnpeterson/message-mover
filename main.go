package main

import (
    "net/http"
    "fmt"
    "log"
    "os"
    "github.com/joho/godotenv"
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

    if !s.ValidateToken(os.Getenv("SLACK_VERIFICATION_TOKEN")) {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    switch s.Command {
    case "/move":
        params := &slack.Msg{Text: s.Text}
        response := fmt.Sprintf("You asked for the weather for %v", params.Text)
        w.Write([]byte(response))

    default:
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

}

func main() {
    err := godotenv.Load("environment.env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    http.HandleFunc("/command", moveMessagesHandler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        panic(err)
    }
}

