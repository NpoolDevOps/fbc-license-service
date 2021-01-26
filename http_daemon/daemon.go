package http_daemon

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
)

type HttpHandler func(w http.ResponseWriter, req *http.Request) (interface{}, error, int)

type HttpRouter struct {
    Location string
    Handler  HttpHandler
}

var routerTable []HttpRouter = make([]HttpRouter, 0)

type ApiResp struct {
    Code  int         `json:"code"`
    Error string      `json:"error"`
    Body  interface{} `json:"body"`
}

func response(w http.ResponseWriter, resp interface{}, err error, code int) error {
    errStr := ""
    if nil != err {
        errStr = err.Error()
    }
    apiResp := ApiResp{
        Code:  code,
        Error: errStr,
        Body:  resp,
    }
    jsonStr, err := json.Marshal(&apiResp)
    if nil != err {
        return err
    }
    w.Write(jsonStr)
    return nil
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
    log.Printf("request %v -> %v [%v]", req.RemoteAddr, req.URL, req.URL.Path)
    if err := req.ParseForm(); nil != err {
        log.Printf("fail to parse form %v", req.URL)
        response(w, struct{}{}, err, -1)
        return
    }
    for _, r := range routerTable {
        if r.Location == req.URL.Path {
            resp, err, code := r.Handler(w, req)
            err = response(w, resp, err, code)
            if nil != err {
                log.Printf("fail to response %v", req.URL)
            }
            return
        }
    }
    response(w, struct{}{}, fmt.Errorf("unknown request %v", req.URL), -3)
}

func Run(port int) error {
    http.HandleFunc("/", rootHandler)

    go func(port int) {
        portStr := fmt.Sprintf(":%v", port)
        log.Printf("start http daemon [%v]", portStr)
        for {
            http.ListenAndServe(portStr, nil)
        }
    }(port)

    return nil
}

func RegisterRouter(router HttpRouter) error {
    for _, r := range routerTable {
        if r.Location == router.Location {
            return errors.New("router already exist")
        }
    }
    log.Printf("add router: %v", router.Location)
    routerTable = append(routerTable, router)
    return nil
}


func RegisterSpecialHandler(wsPath string, handler func(w http.ResponseWriter, req *http.Request)) error {
    http.HandleFunc(wsPath, handler)
    return nil
}
