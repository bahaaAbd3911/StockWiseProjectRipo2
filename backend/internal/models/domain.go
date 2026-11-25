package models

import "time"

// StoreDomain is a custom domain connected to a store.
type StoreDomain struct {
	ID          int64      `json:"id"`
	Domain      string     `json:"domain"`
	IsPrimary   bool       `json:"is_primary"` // the store's main address (emails)
	Verified    bool       `json:"verified"`
	VerifiedAt  *time.Time `json:"verified_at"`
	VerifyToken string     `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	// What the owner has to set up at their DNS provider.
	DNS DomainDNS `json:"dns"`
}

// DomainDNS lists the two DNS records that connect a domain.
type DomainDNS struct {
	// Proves ownership: a TXT record with this name and value.
	TXTName  string `json:"txt_name"`
	TXTValue string `json:"txt_value"`
	// Sends visitors to the platform: a CNAME to this host (for a root
	// domain such as example.com, an ALIAS/ANAME record or the server's IP).
	CNAMETarget string `json:"cname_target"`
}

type AddDomainRequest struct {
	Domain string `json:"domain"`
}
