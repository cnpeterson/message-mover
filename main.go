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


func parseMoveCommandArgs (cmds *moveCommandArgs, cmdArgs string) (err error) {
    splitArgs := strings.Fields(cmdArgs)
    if len(splitArgs) <= 5 {
        err := fmt.Errorf("Incorrect arguements, arguements must be in this format")
        log.Fatal(err)
        return err
    } else {
        messagesTotal := strings.Split(splitArgs[0], "-")
        ms := messagesTotal[0]
        me := messagesTotal[1]
        msint, err := strconv.Atoi(ms)
        if err != nil {
            fmt.Println(err)
        } else {
            cmds.messageStart = msint
        }
        meint, err := strconv.Atoi(me)
        if err != nil {
            fmt.Println(err)
        } else {
            cmds.messageEnd = meint
        }
        cmds.messagesPlaceHolder = splitArgs[1]
        cmds.toPlaceholder = splitArgs[2]
        cmds.channel = splitArgs[3]
        cmds.titlePlaceHolder = splitArgs[4]
        cmds.title = strings.Join(splitArgs[5:], " ")
        return err
    }

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
        cmds := moveCommandArgs{}
        err := parseMoveCommandArgs(&cmds, params.Text)
        if err != nil {
            w.WriteHeader(500)
            return
        }
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

