package main

import (
    "log"
    "net/http"

    "github.com/AdeptTravel/adept-framework/internal/config"
    "github.com/AdeptTravel/adept-framework/internal/handlers"
)

func main() {
    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/", handlers.Info(cfg))

    addr := ":8080"
    log.Printf("⇢ listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}
