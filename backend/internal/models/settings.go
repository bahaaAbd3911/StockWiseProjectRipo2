package models

// StoreSettings is what a store's admin edits in Store Settings: the store's
// name, its logo, the theme colors, and the content of its pages. The keys are
// the frontend's own (camelCase), because the frontend owns this format; the
// backend checks that the values are safe and stores them (one JSON row per
// store, store_settings).
type StoreSettings struct {
	StoreName string `json:"storeName"`
	// Logo is "" (none), an http(s) URL, or an image embedded as a data URL
	// (data:image/png;base64,...), at most about 1 MB.
	Logo string `json:"logo"`
	// Colors maps "light" / "dark" to token -> "#RRGGBB", e.g.
	// {"light": {"text": "#17231D", "button": "#22C55E"}, "dark": {...}}.
	Colors map[string]map[string]string `json:"colors"`
	// Content: texts and images of the store's pages. Anything left out uses
	// the frontend's defaults, so a new store looks complete from the start.
	Content *StoreContent `json:"content,omitempty"`
}

// StoreContent is the customisable content of the shop's pages (English and
// Arabic for every text).
type StoreContent struct {
	Hero           *Banner        `json:"hero,omitempty"`           // home page banner
	ProductsBanner *Banner        `json:"productsBanner,omitempty"` // products page banner
	Features       []Feature      `json:"features,omitempty"`       // the tiles under the banner (max 8)
	Footer         *FooterContent `json:"footer,omitempty"`
}

// Banner: a title, a two-part subtitle (the second part is highlighted), a
// button and an image (an uploaded image URL).
type Banner struct {
	Title        string `json:"title"`
	TitleAr      string `json:"titleAr"`
	Subtitle     string `json:"subtitle"`
	SubtitleAr   string `json:"subtitleAr"`
	Highlight    string `json:"highlight"`
	HighlightAr  string `json:"highlightAr"`
	ButtonText   string `json:"buttonText"`
	ButtonTextAr string `json:"buttonTextAr"`
	Image        string `json:"image"`
}

// Feature is one tile: a Font Awesome icon, a name and a short line.
type Feature struct {
	Icon      string `json:"icon"` // e.g. "fa-solid fa-truck"
	Name      string `json:"name"`
	NameAr    string `json:"nameAr"`
	Comment   string `json:"comment"`
	CommentAr string `json:"commentAr"`
}

type FooterContent struct {
	Tagline     string `json:"tagline"`
	TaglineAr   string `json:"taglineAr"`
	Copyright   string `json:"copyright"`
	CopyrightAr string `json:"copyrightAr"`
	// Social: network -> profile URL (facebook, instagram, x, youtube, tiktok, whatsapp).
	Social map[string]string `json:"social,omitempty"`
}

// Images returns every uploaded-image URL the content refers to (used to
// delete images that a save no longer uses).
func (c *StoreContent) Images() []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, b := range []*Banner{c.Hero, c.ProductsBanner} {
		if b != nil && b.Image != "" {
			out = append(out, b.Image)
		}
	}
	return out
}
