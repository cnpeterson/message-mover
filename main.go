package main

import (
    "time"
    "net/http"
    "bytes"
    "fmt"
    "log"
    "os"
    "strings"
    "strconv"
    "github.com/joho/godotenv"
    "io/ioutil"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    )

type SlackCommandRequest struct {
    Token        string `json:"token"`
    Command      string `json:"command"`
    Text         string `json:"text"`
    ResponseUrl  string `json:"response_url"`
    TriggerId    string `json:"trigger_id"`
    UserId       string `json:"user_id"`
    UserName     string `json:"user_name"`
    TeamId       string `json:"team_id"`
    EnterpriseId string `json:"enterprise_id"`
    ChannelId    string `json:"channel_id"`
}

type MoveCommandArgs struct {
    MessageStart        int    `json:"message_start"`
    MessageEnd          int    `json:"message_end"`
    MessagesPlaceholder string `json:"messages_placeholder"`
    ToPlaceholder       string `json:"to_placeholder"`
    Channel             string `json:"channel"`
    TitlePlaceholder    string `json:"title_place_holder"`
    Title               string `json:"title"`
}

type SlackCovoHistoryRequest struct {
    Token     string   `json:"token"`
    Channel   string   `json:"channel"`
    Cursor    *string  `json:"cursor,omitempty"`
    Inclusive *bool    `json:"inclusive,omitempty"`
    Latest    *float64 `json:"latest,omitempty"`
    Limit     *int     `json:"limit,omitempty"`
    Oldest    *float64 `json:"oldest,omitempty"`
}

type SlackConvHistoryMessage struct {
    Type string `json:"type"`
    User string `json:"user"`
    Text string `json:"text"`
    Ts   string `json:"ts"`
}

type SlackConvoHistoryResponse struct {
    Ok               bool                      `json:"ok"`
    Messages         []SlackConvHistoryMessage `json:"messages"`
    HasMore          bool                      `json:"has_more"`
    PinCount         int                       `json:"pin_count"`
    ResponseMetadata ResponseMetadata          `json:"response_metadata"`
}

type ResponseMetadata struct {
    NextCursor string `json:"next_cursor"`
}

type Message struct {
    Text string `json:"text"`
}

type SlackCommandAuthHeaders struct {
    XSlackRequestTimestamp int64
    XSlackSignature        string
    VersionNumber          string
}

func NewAuthHeaders(r *http.Request) (a SlackCommandAuthHeaders, err error) {
    // VersionNumber is always v0 with current slack API
    t := r.Header.Get("X-Slack-Request-Timestamp")
    s := r.Header.Get("X-Slack-Signature")
    ti, err := strconv.ParseInt(t, 0, 64)
    if err != nil {
        fmt.Println(err)
        return a, err
    }
    now := time.Now()
    sec := now.Unix()
    // Checking to see if request is more than five minutes from local time
    if (ti - sec) > 60 * 5 {
        err = fmt.Errorf("Request older than 5 minutes")
        return a, err
    }
    a = SlackCommandAuthHeaders{ti, s, "v0"}
    return a, err
}

func SlackCommandParse (r *http.Request) (s SlackCommandRequest, err error) {
    if err = r.ParseForm(); err != nil {
        return s, err
    }
    s.Token = r.PostForm.Get("token")
    s.Command = r.PostForm.Get("command")
    s.Text = r.PostForm.Get("text")
    s.ResponseUrl = r.PostForm.Get("response_url")
    s.TriggerId = r.PostForm.Get("trigger_id")
    s.UserId = r.PostForm.Get("user_id")
    s.UserName = r.PostForm.Get("user_name")
    s.TeamId = r.PostForm.Get("team_id")
    s.EnterpriseId = r.PostForm.Get("enterprise_id")
    s.ChannelId = r.PostForm.Get("channel_id")
    return s, err
}

func (s *SlackCommandRequest) TokenIsValid(t string) bool {
    if len(s.Token) > 0 && s.Token == t {
        return true
    } else {
        return false
    }
}

func (a *SlackCommandAuthHeaders) SignatureIsValid(sig string) bool {
    if sig == a.XSlackSignature {
        return true
    } else {
        return false
    }
}

func parseMoveCommandArgs (cmds *MoveCommandArgs, cmdArgs string) (err error) {
    // use regex to validate instead of this
    splitArgs := strings.Fields(cmdArgs)
    if len(splitArgs) <= 5 {
        err := fmt.Errorf("Incorrect arguments, arguments must be in this format: [# of messages] messages to [channel] titled [title]")
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
            cmds.MessageStart = msint
        }
        meint, err := strconv.Atoi(me)
        if err != nil {
            fmt.Println(err)
        } else {
            cmds.MessageEnd = meint
        }
        cmds.MessagesPlaceholder = splitArgs[1]
        cmds.ToPlaceholder = splitArgs[2]
        cmds.Channel = splitArgs[3]
        cmds.TitlePlaceholder = splitArgs[4]
        cmds.Title = strings.Join(splitArgs[5:], " ")
        return err
    }

}

func SlackCommandHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("received request")
    fmt.Printf("%s", r.Body)
    s, err := SlackCommandParse(r)
    if err != nil {
        fmt.Println("Error parsing slack command")
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // validate token
    if !s.TokenIsValid(os.Getenv("SLACK_VERIFICATION_TOKEN")) {
        fmt.Println("incorrect slack verification token")
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    switch s.Command {
    case "/move":
        params := &Message{Text: s.Text}
        cmds := MoveCommandArgs{}
        err := parseMoveCommandArgs(&cmds, params.Text)
        if err != nil {
            w.WriteHeader(500)
            return
        }
        // authenticate
        var b bytes.Buffer
        var as string
        var sig string

        a, err := NewAuthHeaders(r)
        if err != nil {
            fmt.Println(err)
            w.WriteHeader(500)
            return
        }

        // creating string to be compared to X-Slack-Signature 
        b.WriteString(a.VersionNumber)
        b.WriteString(":")
        ts := strconv.FormatInt(a.XSlackRequestTimestamp, 0)
        b.WriteString(ts)
        b.WriteString(":")
        body, _ := ioutil.ReadAll(r.Body)
        sb := string(body)
        b.WriteString(sb)
        as = b.String()
        b.Reset()
        // creating hex to compare with X-Slack-Signature
        b.WriteString(a.VersionNumber)
        b.WriteString("=")
        sec := os.Getenv("SLACK_SIGNING_SECRET")
        fmt.Printf("Secret: %s Data: %s\n", sec, as)
        h := hmac.New(sha256.New, []byte(sec))
        h.Write([]byte(as))
        sha := hex.EncodeToString(h.Sum(nil))
        fmt.Println("Result: " + sha)
        b.WriteString(sha)
        sig = b.String()
        b.Reset()
        // comparing X-Slack-Signature and our string
        if !a.SignatureIsValid(sig) {
            fmt.Println("sigs don't match")
            fmt.Printf("Sig: %s RequestSig: %s", sig, a.XSlackSignature)
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        // Check permissions to channels
        //   Possibly add direct messages later 

        // get chat history
        //   grab range of chat history needed

        // Post chat to channel as one message
        //    Messages included will be added as a thread

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

    http.HandleFunc("/command", SlackCommandHandler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        panic(err)
    }
}

