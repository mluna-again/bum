package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	tea "charm.land/bubbletea/v2"
	"github.com/go-chi/chi/v5"
)

type serverNewPaneMsg struct {
	pane Pane
}

func startServer(t *tea.Program) {
	router := chi.NewRouter()
	router.Post("/new", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		info := Pane{}
		d := json.NewDecoder(r.Body)
		err := d.Decode(&info)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if info.TmuxPaneID == "" {
			http.Error(w, "pane_id required", http.StatusBadRequest)
			return
		}

		if info.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}

		t.Send(serverNewPaneMsg{pane: info})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s added", info.Title)))
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
