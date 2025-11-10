package gclplugin

// Settings control the docnamecheck analyzer when loaded via golangci-lint's module plugin system.
type Settings struct {
	MaxDist                 *int    `json:"maxdist,omitempty"`
	IncludeExported         *bool   `json:"include-exported,omitempty"`
	IncludeUnexported       *bool   `json:"include-unexported,omitempty"`
	IncludeTypes            *bool   `json:"include-types,omitempty"`
	IncludeGenerated        *bool   `json:"include-generated,omitempty"`
	IncludeInterfaceMethods *bool   `json:"include-interface-methods,omitempty"`
	AllowedLeadingWords     *string `json:"allowed-leading-words,omitempty"`
	AllowedPrefixes         *string `json:"allowed-prefixes,omitempty"`
}
