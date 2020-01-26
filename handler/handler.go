package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Sigafoos/pvpservice/pvp"
)

type Handler struct {
	pvp *pvp.PVP
}

func New(pvp *pvp.PVP) *Handler {
	return &Handler{
		pvp: pvp,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading register body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user pvp.Player
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("error unmarshalling register body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.ID == "" || user.Username == "" || user.Server == "" || user.IGN == "" || user.FriendCode == "" {
		// TODO a helpful response
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.pvp.Register(user)
	if err != nil {
		log.Printf("error registering user: %s", err)
		// TODO 419 or 500
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	server := r.FormValue("server")
	if server == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO an error message
		return
	}
	players := h.pvp.ListPlayers(server)
	b, err := json.Marshal(players)
	if err != nil {
		log.Printf("error marshalling player list: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(b)
}
