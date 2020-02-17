package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Sigafoos/pvpservice/pvp"

	"github.com/gocraft/dbr/v2"
	"github.com/lib/pq"
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

	user := h.getUserFromRequest(r)
	if user == nil || user.ID == "" || user.Server == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.pvp.RegisterUser(user)
	if err != nil {
		pqErr := err.(*pq.Error)
		switch pqErr.Code {
		case "23505":
			w.WriteHeader(http.StatusConflict)
		default:
			log.Printf("error registering user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)

}

func (h *Handler) registerPlayer(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromRequest(r)

	// egg for ultra is optional
	if user == nil || user.ID == "" || user.Username == "" || user.IGN == "" || user.FriendCode == "" {
		// TODO a helpful response
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.pvp.CreateUser(user)
	if err != nil {
		pqErr := err.(*pq.Error)
		switch pqErr.Code {
		case "23505":
			w.WriteHeader(http.StatusConflict)
		default:
			log.Printf("error creating user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
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

func (h *Handler) Player(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getPlayer(w, r)

	case http.MethodPost:
		h.registerPlayer(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Friendship(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createFriendship(w, r)

	case http.MethodGet:
		h.getFriends(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	player, err := h.pvp.GetPlayer(id)
	if err != nil {
		if err == dbr.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			pqErr, ok := err.(*pq.Error)
			if !ok {
				log.Println(err)
			} else {
				log.Printf("%s: %s", pqErr.Code, pqErr.Message)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	b, err := json.Marshal(player)
	if err != nil {
		log.Printf("error marshalling player: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(b)
}

func (h *Handler) getUserFromRequest(r *http.Request) *pvp.Player {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil
	}
	var user pvp.Player
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil
	}

	return &user
}

func (h *Handler) createFriendship(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var f pvp.Friendship
	err = json.Unmarshal(body, &f)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.pvp.AddFriend(f)
	if err != nil {
		// TODO error checking
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) getFriends(w http.ResponseWriter, r *http.Request) {
	// TODO check if the user exists, as this won't throw an error
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		// TODO an error message
		return
	}
	friends, err := h.pvp.GetFriends(id)
	if err != nil {
		log.Println("error getting friends" + err.Error())
		// TODO
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(friends)
	if err != nil {
		log.Printf("error marshalling friend list: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(b)
}
