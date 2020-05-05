package slack

import (
    "time"
    "net/http"
    "bytes"
    "fmt"
    "strings"
    "strconv"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io/ioutil"
    )

const (
    VersionNumber = "v0"
    hTimestamp = "X-Slack-Request-Timestamp"
    hSignature = "X-Slack-Signature"
)

type AuthInfo struct {
    Signature []byte
    Hash      []byte
}

func (a *AuthInfo) SignatureIsValid() bool {
    if hmac.Equal(a.Signature, a.Hash)  {
        return true
    } else {
        return false
    }
}

func NewAuthInfo(r *http.Request, secret string) (a AuthInfo, err error) {
    // Grabbing timestamp and signature from header
    t := r.Header.Get(hTimestamp)
    s := r.Header.Get(hSignature)
    ti, err := strconv.ParseInt(t, 10, 64)
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
    // Decoding signature from header
    hs, err := hex.DecodeString(strings.TrimPrefix(s, fmt.Sprintf("%s=", VersionNumber)))
    if err != nil {
        fmt.Println(err)
        return a, err
    }
    // Getting contents of the body and then putting them back 
    // Doing this so after it authorizes I can still parse the body
    body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
    r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
    // Creating hmac to compare with signature
    hash := hmac.New(sha256.New, []byte(secret))
    if _, err = hash.Write([]byte(fmt.Sprintf("%s:%s:%s", VersionNumber, t, body))); err != nil {
        fmt.Println(err)
        return a, err
    }
    sig := hash.Sum(nil)
    a = AuthInfo{hs, sig}
    return a, err
}

func Authentication(r *http.Request, secret string) (a AuthInfo, err error) {
    ai, err := NewAuthInfo(r, secret)
    return ai, err

}
