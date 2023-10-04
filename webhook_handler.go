package github

import (
	"encoding/json"
	"net/http"
)

type WebhookHandler func(w *Webhook) error

func (h WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wh Webhook
	json.NewDecoder(r.Body).Decode(&wh)
	err := h(&wh)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
