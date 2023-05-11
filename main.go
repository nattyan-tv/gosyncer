package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Config struct {
    ProxyUrl string `json:"proxy_url"`
    ServerUrl string `json:"server_url"`
    Path string `json:"path"`
}

type ServerInfo struct {
    Id string `json:"id"`
    Files map[string]int `json:"files"`
    Removed []string `json:"removed"`
}

func load_config() Config {
    path := filepath.Join("config.json")
    jtext, err := ioutil.ReadFile(path)
    if err != nil {
        fmt.Println("エラーが発生しました。")
        fmt.Println(err)
        os.Exit(1)
    }
    var config Config
    json.Unmarshal([]byte(jtext), &config)
    return config
}

func set_proxy(proxy_url string) {
    proxyUrl, err := url.Parse(proxy_url)
    if err == nil {
        http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
    }
}

func main() {
    config := load_config()
    if config.ProxyUrl != "" {
        set_proxy(config.ProxyUrl)
    }
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    surl, _ := url.JoinPath(config.ServerUrl, "info.json")

    resp, err := client.Get(surl)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(body))

    file_check(config.Path)
}

func file_check(input_path string) {
    err := filepath.Walk(input_path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        fmt.Printf("[%s] %d\n", path, info.ModTime().UnixMilli())
        return nil
    })
    if err != nil {
        panic(err)
    }
}

func file_download(filepath string, url string, client http.Client) error {
    resp, err := client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}
