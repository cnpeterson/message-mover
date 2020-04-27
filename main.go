package main

import (
    "net/http"
    "fmt"
    "log"
    "os"
    "strings"
    "strconv"
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

type moveCommandArgs struct {
    messageStart        int
    messageEnd          int
    messagesPlaceHolder string
    toPlaceholder       string
    channel             string
    titlePlaceHolder    string
    title               string
}


func parseMoveCommandArgs (cmdArgs string) moveCommandArgs {
    splitArgs := strings.Fields(cmdArgs)
    messagesTotal := strings.Split(splitArgs[0], "-")
    ms := messagesTotal[0]
    me := messagesTotal[1]
    msstr, err := strconv.Atoi(ms)
    if err != nil {
        fmt.Println(err)
    }
    mestr, err := strconv.Atoi(me)
    if err != nil {
        fmt.Println(err)
    }
    mph := splitArgs[1]
    tph := splitArgs[2]
    channel := splitArgs[3]
    tiph := splitArgs[4]
    ti := strings.Join(splitArgs[5:], " ")

    cmds := moveCommandArgs{msstr, mestr, mph, tph, channel, tiph, ti}
    return cmds

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
        cmds := parseMoveCommandArgs(params.Text)
        response := fmt.Sprintf("You want to move messages for %v", cmds)
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

