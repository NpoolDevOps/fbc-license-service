package utils


import (
    "encoding/json"
    "io/ioutil"
    "net/http"
)


func ReadRequestBody(req *http.Request) (interface{}, error) {

    body, _:= ioutil.ReadAll(req.Body)
    var msg interface{}
    err := json.Unmarshal(body, &msg)
    if err != nil {
        return nil, err
    }

    return msg, nil
}