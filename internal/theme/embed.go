package theme

import _ "embed"

//go:embed theme.css
var embeddedCSS string

func init() {
	Init(embeddedCSS)
}
