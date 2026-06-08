package api

import "net/http"

func (api *API) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	cacheData, err := readCacheFromJSON()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed_to_read_cache"})
	}

	tags := cacheData.Tags

	WriteJSON(w, http.StatusOK, tags)
}
