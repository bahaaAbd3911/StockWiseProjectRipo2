package models

// StoreSettings is what an admin edits in Store Settings: the store's name,
// its logo and the theme colors for light and dark mode. The keys are the
// frontend's own (camelCase), because the frontend owns this format; the
// backend only checks that the values are safe and stores them.
type StoreSettings struct {
	StoreName string `json:"storeName"`
	// Logo is "" (none), an http(s) URL, or an image embedded as a data URL
	// (data:image/png;base64,...), at most about 1 MB.
	Logo string `json:"logo"`
	// Colors maps "light" / "dark" to token -> "#RRGGBB", e.g.
	// {"light": {"text": "#17231D", "button": "#22C55E"}, "dark": {...}}.
	Colors map[string]map[string]string `json:"colors"`
}
