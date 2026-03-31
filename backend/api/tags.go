package api

import "net/http"

func (api *API) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}

	posts, err := readAllPostListItems()
	if err != nil {
		api.Logger.Printf("failed to read tags: %v", err)
		WriteError(w, http.StatusInternalServerError, "failed_to_read_tags")
		return
	}

	seen := make(map[string]struct{})
	tags := make([]string, 0)
	for _, post := range posts {
		for _, tag := range post.Tag {
			if tag == "" {
				continue
			}
			if _, exists := seen[tag]; exists {
				continue
			}
			seen[tag] = struct{}{}
			tags = append(tags, tag)
		}
	}

	WriteJSON(w, http.StatusOK, tags)
}
