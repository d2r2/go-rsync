package gtkui

import (
	"bytes"
	"html"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Markup keeps data to create and generate Pango Markup strings.
// Simplify construct of all corresponding format tags to translate
// calls to Pango Markup subsystem.
type Markup struct {
	font       MarkupFont
	foreground MarkupColor
	background MarkupColor
	left       interface{}
	right      interface{}
	middle     []*Markup
}

// NewMarkup create new markup object. Allows to create nested markup tags.
func NewMarkup(font MarkupFont, foreground MarkupColor, backround MarkupColor,
	left, right interface{}, spans ...*Markup) *Markup {

	/*
		if left != nil {
			if str, ok := left.(string); ok {
				left = escapeString(str)
			}
		}
		if right != nil {
			if str, ok := right.(string); ok {
				right = escapeString(str)
			}
		}
	*/

	sp := &Markup{font: font, foreground: foreground, background: backround, left: left, right: right, middle: spans}
	return sp
}

// String provide Stringer interface.
func (v *Markup) String() string {
	var buf bytes.Buffer
	formatMarkup(v, &buf)
	return buf.String()
}

// formatMarkup write Pango Markup string stored in bytes.Buffer object.
func formatMarkup(span *Markup, buf *bytes.Buffer) {
	buf.WriteString("<span")
	if span.font != 0 {
		buf.WriteString(spew.Sprintf(" %s", span.font))
	}
	if span.foreground != 0 {
		buf.WriteString(spew.Sprintf(" fgcolor=%s", span.foreground))
	}
	if span.background != 0 {
		buf.WriteString(spew.Sprintf(" bgcolor=%s", span.background))
	}
	buf.WriteString(">")
	if span.left != nil {
		str := spew.Sprintf("%v", span.left)

		if _, ok := span.left.(string); ok {
			str = html.EscapeString(str)
		}

		buf.WriteString(str)
	}
	for _, item := range span.middle {
		formatMarkup(item, buf)
	}
	if span.right != nil {
		str := spew.Sprintf("%v", span.right)

		if _, ok := span.right.(string); ok {
			str = html.EscapeString(str)
		}

		buf.WriteString(str)
	}
	buf.WriteString("</span>")
}

/*
var markupEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	`"`, "&#34;", // "&#34;" is shorter than "&quot;".
	`<`, "&lt;",
	`>`, "&gt;",
)

// EscapeString escapes special characters like "<" to become "&lt;". It
// escapes only five such characters: <, >, &, ' and ".
// UnescapeString(EscapeString(s)) == s always holds, but the converse isn't
// always true.
func escapeString(s string) string {
	return markupEscaper.Replace(s)
}
*/

// MarkupFont is a bitmask to generate Pango Markup font attributes.
// Read: https://developer.gnome.org/pango/stable/PangoMarkupFormat.html
type MarkupFont int64

// Read: https://developer.gnome.org/pango/stable/PangoMarkupFormat.html
const (
	// Pango Font Size
	MARKUP_SIZE_XX_SMALL MarkupFont = 1 << iota
	MARKUP_SIZE_X_SMALL
	MARKUP_SIZE_SMALL
	MARKUP_SIZE_MEDIUM
	MARKUP_SIZE_LARGE
	MARKUP_SIZE_X_LARGE
	MARKUP_SIZE_XX_LARGE
	MARKUP_SIZE_SMALLER
	MARKUP_SIZE_LARGER
	// Pango Font Style
	MARKUP_STYLE_NORMAL
	MARKUP_STYLE_OBLIQUE
	MARKUP_STYLE_ITALIC
	// Pango Font Weight
	MARKUP_WEIGHT_ULTRALIGHT
	MARKUP_WEIGHT_LIGHT
	MARKUP_WEIGHT_NORMAL
	MARKUP_WEIGHT_BOLD
	MARKUP_WEIGHT_ULTRABOLD
	MARKUP_WEIGHT_HEAVY
	// Pango Font Variant
	MARKUP_VARIANT_NORMAL
	MARKUP_VARIANT_SMALLCAPS
	// Pango Font Stretch
	MARKUP_STRETCH_ULTRACONDENSED
	MARKUP_STRETCH_EXTRACONDENSED
	MARKUP_STRETCH_CONDENSED
	MARKUP_STRETCH_SEMICONDENSED
	MARKUP_STRETCH_NORMAL
	MARKUP_STRETCH_SEMIEXPANDED
	MARKUP_STRETCH_EXPANDED
	MARKUP_STRETCH_EXTRAEXPANDED
	MARKUP_STRETCH_ULTRAEXPANDED
)

// String provide Stringer interface.
func (v MarkupFont) String() string {
	var buf bytes.Buffer
	const sizeAttr = "font_size"
	const styleAttr = "font_style"
	const weightAttr = "font_weight"
	const variantAttr = "font_variant"
	const stretchAttr = "font_stretch"
	const template = "%s='%s' "
	// Pango Font Size
	if v&MARKUP_SIZE_XX_SMALL != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "xx-small"))
	}
	if v&MARKUP_SIZE_X_SMALL != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "x-small"))
	}
	if v&MARKUP_SIZE_SMALL != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "small"))
	}
	if v&MARKUP_SIZE_MEDIUM != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "medium"))
	}
	if v&MARKUP_SIZE_LARGE != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "large"))
	}
	if v&MARKUP_SIZE_X_LARGE != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "x-large"))
	}
	if v&MARKUP_SIZE_XX_LARGE != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "xx-large"))
	}
	if v&MARKUP_SIZE_SMALLER != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "smaller"))
	}
	if v&MARKUP_SIZE_LARGER != 0 {
		buf.WriteString(spew.Sprintf(template, sizeAttr, "larger"))
	}
	// Pango Font Style
	if v&MARKUP_STYLE_NORMAL != 0 {
		buf.WriteString(spew.Sprintf(template, styleAttr, "normal"))
	}
	if v&MARKUP_STYLE_OBLIQUE != 0 {
		buf.WriteString(spew.Sprintf(template, styleAttr, "oblique"))
	}
	if v&MARKUP_STYLE_ITALIC != 0 {
		buf.WriteString(spew.Sprintf(template, styleAttr, "italic"))
	}
	// Pango Font Weight
	if v&MARKUP_WEIGHT_ULTRALIGHT != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "ultralight"))
	}
	if v&MARKUP_WEIGHT_LIGHT != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "light"))
	}
	if v&MARKUP_WEIGHT_NORMAL != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "normal"))
	}
	if v&MARKUP_WEIGHT_BOLD != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "bold"))
	}
	if v&MARKUP_WEIGHT_ULTRABOLD != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "ultrabold"))
	}
	if v&MARKUP_WEIGHT_HEAVY != 0 {
		buf.WriteString(spew.Sprintf(template, weightAttr, "heavy"))
	}
	// Pango Font Variant
	if v&MARKUP_VARIANT_NORMAL != 0 {
		buf.WriteString(spew.Sprintf(template, variantAttr, "normal"))
	}
	if v&MARKUP_VARIANT_SMALLCAPS != 0 {
		buf.WriteString(spew.Sprintf(template, variantAttr, "smallcaps"))
	}
	// Pango Font Stretch
	if v&MARKUP_STRETCH_ULTRACONDENSED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "ultracondensed"))
	}
	if v&MARKUP_STRETCH_EXTRACONDENSED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "extracondensed"))
	}
	if v&MARKUP_STRETCH_CONDENSED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "condensed"))
	}
	if v&MARKUP_STRETCH_SEMICONDENSED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "semicondensed"))
	}
	if v&MARKUP_STRETCH_NORMAL != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "normal"))
	}
	if v&MARKUP_STRETCH_SEMIEXPANDED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "semiexpanded"))
	}
	if v&MARKUP_STRETCH_EXPANDED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "expanded"))
	}
	if v&MARKUP_STRETCH_EXTRAEXPANDED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "extraexpanded"))
	}
	if v&MARKUP_STRETCH_ULTRAEXPANDED != 0 {
		buf.WriteString(spew.Sprintf(template, stretchAttr, "ultraexpanded"))
	}

	str := strings.TrimSpace(buf.String())
	return str
}

// MarkupColor is a flag to generate Pango Markup color attributes.
type MarkupColor int

// Colors taken from: https://en.wikipedia.org/wiki/X11_color_names
const (
	MARKUP_COLOR_ALICE_BLUE MarkupColor = iota
	MARKUP_COLOR_ANTIQUE_WHITE
	MARKUP_COLOR_AQUE
	MARKUP_COLOR_AQUAMARINE
	MARKUP_COLOR_AZURE
	MARKUP_COLOR_BEIGE
	MARKUP_COLOR_BISQUE
	MARKUP_COLOR_BLACK
	MARKUP_COLOR_BLANCHED_ALMOND
	MARKUP_COLOR_BLUE
	MARKUP_COLOR_BLUE_VIOLET
	MARKUP_COLOR_BROWN
	MARKUP_COLOR_BURLYWOOD
	MARKUP_COLOR_CADET_BLUE
	MARKUP_COLOR_CHARTREUSE
	MARKUP_COLOR_CHOCOLATE
	MARKUP_COLOR_CORAL
	MARKUP_COLOR_CORNFLOWER
	MARKUP_COLOR_CORNSILK
	MARKUP_COLOR_CRIMSON
	MARKUP_COLOR_CYAN
	MARKUP_COLOR_DARK_BLUE
	MARKUP_COLOR_DARK_CYAN
	MARKUP_COLOR_DARK_GOLDENROD
	MARKUP_COLOR_DARK_GRAY
	MARKUP_COLOR_DARK_GREEN
	MARKUP_COLOR_DARK_KHAKI
	MARKUP_COLOR_DARK_MAGENTA
	MARKUP_COLOR_DARK_OLIVE_GREEN
	MARKUP_COLOR_DARK_ORANGE
	MARKUP_COLOR_DARK_ORCHID
	MARKUP_COLOR_DARK_RED
	MARKUP_COLOR_DARK_SALMON
	MARKUP_COLOR_DARK_SEA_GREEN
	MARKUP_COLOR_DARK_SLATE_BLUE
	MARKUP_COLOR_DARK_SLATE_GRAY
	MARKUP_COLOR_DARK_TURQUOISE
	MARKUP_COLOR_DARK_VIOLET
	MARKUP_COLOR_DEEP_PINK
	MARKUP_COLOR_DEEP_SKY_BLUE
	MARKUP_COLOR_DIM_GRAY
	MARKUP_COLOR_DODGER_BLUE
	MARKUP_COLOR_FIREBRICK
	MARKUP_COLOR_FLORAL_WHITE
	MARKUP_COLOR_FOREST_GREEN
	MARKUP_COLOR_FUCHSIA
	MARKUP_COLOR_GAINSBORO
	MARKUP_COLOR_GHOST_WHITE
	MARKUP_COLOR_GOLD
	MARKUP_COLOR_GOLDENROD
	MARKUP_COLOR_GRAY
	MARKUP_COLOR_WEB_GRAY
	MARKUP_COLOR_GREEN
	MARKUP_COLOR_WEB_GREEN
	MARKUP_COLOR_GREEN_YELLOW
	MARKUP_COLOR_HONEYDEW
	MARKUP_COLOR_HOT_PINK
	MARKUP_COLOR_INDIAN_RED
	MARKUP_COLOR_INDIGO
	MARKUP_COLOR_IVORY
	MARKUP_COLOR_KHAKI
	MARKUP_COLOR_LAVENDER
	MARKUP_COLOR_LAVENDER_BLUSH
	MARKUP_COLOR_LAWN_GREEN
	MARKUP_COLOR_LEMON_CHIFFON
	MARKUP_COLOR_LIGHT_BLUE
	MARKUP_COLOR_LIGHT_CORAL
	MARKUP_COLOR_LIGHT_CYAN
	MARKUP_COLOR_LIGHT_GOLDENROD
	MARKUP_COLOR_LIGHT_GRAY
	MARKUP_COLOR_LIGHT_GREEN
	MARKUP_COLOR_LIGHT_PINK
	MARKUP_COLOR_LIGHT_SALMON
	MARKUP_COLOR_LIGHT_SEA_GREEN
	MARKUP_COLOR_LIGHT_SKY_BLUE
	MARKUP_COLOR_LIGHT_SLATE_GRAY
	MARKUP_COLOR_LIGHT_STEEL_BLUE
	MARKUP_COLOR_LIGHT_YELLOW
	MARKUP_COLOR_LIME
	MARKUP_COLOR_LIME_GREEN
	MARKUP_COLOR_LINEN
	MARKUP_COLOR_MAGENTA
	MARKUP_COLOR_MAROON
	MARKUP_COLOR_WEB_MAROON
	MARKUP_COLOR_MEDIUM_AQUAMARINE
	MARKUP_COLOR_MEDIUM_BLUE
	MARKUP_COLOR_MEDIUM_ORCHID
	MARKUP_COLOR_MEDIUM_PURPLE
	MARKUP_COLOR_MEDIUM_SEA_GREEN
	MARKUP_COLOR_MEDIUM_SLATE_BLUE
	MARKUP_COLOR_MEDIUM_SPRING_GREEN
	MARKUP_COLOR_MEDIUM_TURQUOISE
	MARKUP_COLOR_MEDIUM_VIOLET_RED
	MARKUP_COLOR_MIDNIGHT_BLUE
	MARKUP_COLOR_MINT_CREAM
	MARKUP_COLOR_MISTY_ROSE
	MARKUP_COLOR_MOCCASIN
	MARKUP_COLOR_NAVAJO_WHITE
	MARKUP_COLOR_NAVY_BLUE
	MARKUP_COLOR_OLD_LACE
	MARKUP_COLOR_OLIVE
	MARKUP_COLOR_OLIVE_DRAB
	MARKUP_COLOR_ORANGE
	MARKUP_COLOR_ORANGE_RED
	MARKUP_COLOR_ORCHID
	MARKUP_COLOR_PALE_GOLDENROD
	MARKUP_COLOR_PALE_GREEN
	MARKUP_COLOR_PALE_TURQUOISE
	MARKUP_COLOR_PALE_VIOLET_RED
	MARKUP_COLOR_PAPAYA_WHIP
	MARKUP_COLOR_PEACH_PUFF
	MARKUP_COLOR_PERU
	MARKUP_COLOR_PINK
	MARKUP_COLOR_PLUM
	MARKUP_COLOR_POWDER_BLUE
	MARKUP_COLOR_PURPLE
	MARKUP_COLOR_WEB_PURPLE
	MARKUP_COLOR_REBECCA_PURPLE
	MARKUP_COLOR_RED
	MARKUP_COLOR_ROSY_BROWN
	MARKUP_COLOR_ROYAL_BLUE
	MARKUP_COLOR_SADDLE_BROWN
	MARKUP_COLOR_SALMON
	MARKUP_COLOR_SANDY_BROWN
	MARKUP_COLOR_SEA_GREEN
	MARKUP_COLOR_SEASHELL
	MARKUP_COLOR_SIENNA
	MARKUP_COLOR_SILVER
	MARKUP_COLOR_SKY_BLUE
	MARKUP_COLOR_SLATE_BLUE
	MARKUP_COLOR_SLATE_GRAY
	MARKUP_COLOR_SNOW
	MARKUP_COLOR_SPRING_GREEN
	MARKUP_COLOR_STEEL_BLUE
	MARKUP_COLOR_TAN
	MARKUP_COLOR_TEAL
	MARKUP_COLOR_THISTLE
	MARKUP_COLOR_TOMATO
	MARKUP_COLOR_TURQUOISE
	MARKUP_COLOR_VIOLET
	MARKUP_COLOR_WHEAT
	MARKUP_COLOR_WHITE
	MARKUP_COLOR_WHITE_SMOKE
	MARKUP_COLOR_YELLOW
	MARKUP_COLOR_YELLOW_GREEN
)

// String provide Stringer interface.
func (v MarkupColor) String() string {
	var m = map[MarkupColor]string{
		MARKUP_COLOR_ALICE_BLUE:          "Alice Blue",
		MARKUP_COLOR_ANTIQUE_WHITE:       "Antique White",
		MARKUP_COLOR_AQUE:                "Aqua",
		MARKUP_COLOR_AQUAMARINE:          "Aquamarine",
		MARKUP_COLOR_AZURE:               "Azure",
		MARKUP_COLOR_BEIGE:               "Beige",
		MARKUP_COLOR_BISQUE:              "Bisque",
		MARKUP_COLOR_BLACK:               "Black",
		MARKUP_COLOR_BLANCHED_ALMOND:     "Blanched Almond",
		MARKUP_COLOR_BLUE:                "Blue",
		MARKUP_COLOR_BLUE_VIOLET:         "Blue Violet",
		MARKUP_COLOR_BROWN:               "Brown",
		MARKUP_COLOR_BURLYWOOD:           "Burlywood",
		MARKUP_COLOR_CADET_BLUE:          "Cadet Blue",
		MARKUP_COLOR_CHARTREUSE:          "Chartreuse",
		MARKUP_COLOR_CHOCOLATE:           "Chocolate",
		MARKUP_COLOR_CORAL:               "Coral",
		MARKUP_COLOR_CORNFLOWER:          "Cornflower",
		MARKUP_COLOR_CORNSILK:            "Cornsilk",
		MARKUP_COLOR_CRIMSON:             "Crimson",
		MARKUP_COLOR_CYAN:                "Cyan",
		MARKUP_COLOR_DARK_BLUE:           "Dark Blue",
		MARKUP_COLOR_DARK_CYAN:           "Dark Cyan",
		MARKUP_COLOR_DARK_GOLDENROD:      "Dark Goldenrod",
		MARKUP_COLOR_DARK_GRAY:           "Dark Gray",
		MARKUP_COLOR_DARK_GREEN:          "Dark Green",
		MARKUP_COLOR_DARK_KHAKI:          "Dark Khaki",
		MARKUP_COLOR_DARK_MAGENTA:        "Dark Magenta",
		MARKUP_COLOR_DARK_OLIVE_GREEN:    "Dark Olive Green",
		MARKUP_COLOR_DARK_ORANGE:         "Dark Orange",
		MARKUP_COLOR_DARK_ORCHID:         "Dark Orchid",
		MARKUP_COLOR_DARK_RED:            "Dark Red",
		MARKUP_COLOR_DARK_SALMON:         "Dark Salmon",
		MARKUP_COLOR_DARK_SEA_GREEN:      "Dark Sea Green",
		MARKUP_COLOR_DARK_SLATE_BLUE:     "Dark Slate Blue",
		MARKUP_COLOR_DARK_SLATE_GRAY:     "Dark Slate Gray",
		MARKUP_COLOR_DARK_TURQUOISE:      "Dark Turquoise",
		MARKUP_COLOR_DARK_VIOLET:         "Dark Violet",
		MARKUP_COLOR_DEEP_PINK:           "Deep Pink",
		MARKUP_COLOR_DEEP_SKY_BLUE:       "Deep Sky Blue",
		MARKUP_COLOR_DIM_GRAY:            "Dim Gray",
		MARKUP_COLOR_DODGER_BLUE:         "Dodger Blue",
		MARKUP_COLOR_FIREBRICK:           "Firebrick",
		MARKUP_COLOR_FLORAL_WHITE:        "Floral White",
		MARKUP_COLOR_FOREST_GREEN:        "Forest Green",
		MARKUP_COLOR_FUCHSIA:             "Fuchsia",
		MARKUP_COLOR_GAINSBORO:           "Gainsboro",
		MARKUP_COLOR_GHOST_WHITE:         "Ghost White",
		MARKUP_COLOR_GOLD:                "Gold",
		MARKUP_COLOR_GOLDENROD:           "Goldenrod",
		MARKUP_COLOR_GRAY:                "Gray",
		MARKUP_COLOR_WEB_GRAY:            "Web Gray",
		MARKUP_COLOR_GREEN:               "Green",
		MARKUP_COLOR_WEB_GREEN:           "Web Green",
		MARKUP_COLOR_GREEN_YELLOW:        "Green Yellow",
		MARKUP_COLOR_HONEYDEW:            "Honeydew",
		MARKUP_COLOR_HOT_PINK:            "Hot Pink",
		MARKUP_COLOR_INDIAN_RED:          "Indian Red",
		MARKUP_COLOR_INDIGO:              "Indigo",
		MARKUP_COLOR_IVORY:               "Ivory",
		MARKUP_COLOR_KHAKI:               "Khaki",
		MARKUP_COLOR_LAVENDER:            "Lavender",
		MARKUP_COLOR_LAVENDER_BLUSH:      "Lavender Blush",
		MARKUP_COLOR_LAWN_GREEN:          "Lawn Green",
		MARKUP_COLOR_LEMON_CHIFFON:       "Lemon Chiffon",
		MARKUP_COLOR_LIGHT_BLUE:          "Light Blue",
		MARKUP_COLOR_LIGHT_CORAL:         "Light Coral",
		MARKUP_COLOR_LIGHT_CYAN:          "Light Cyan",
		MARKUP_COLOR_LIGHT_GOLDENROD:     "Light Goldenrod",
		MARKUP_COLOR_LIGHT_GRAY:          "Light Gray",
		MARKUP_COLOR_LIGHT_GREEN:         "Light Green",
		MARKUP_COLOR_LIGHT_PINK:          "Light Pink",
		MARKUP_COLOR_LIGHT_SALMON:        "Light Salmon",
		MARKUP_COLOR_LIGHT_SEA_GREEN:     "Light Sea Green",
		MARKUP_COLOR_LIGHT_SKY_BLUE:      "Light Sky Blue",
		MARKUP_COLOR_LIGHT_SLATE_GRAY:    "Light Slate Gray",
		MARKUP_COLOR_LIGHT_STEEL_BLUE:    "Light Steel Blue",
		MARKUP_COLOR_LIGHT_YELLOW:        "Light Yellow",
		MARKUP_COLOR_LIME:                "Lime",
		MARKUP_COLOR_LIME_GREEN:          "Lime Green",
		MARKUP_COLOR_LINEN:               "Linen",
		MARKUP_COLOR_MAGENTA:             "Magenta",
		MARKUP_COLOR_MAROON:              "Maroon",
		MARKUP_COLOR_WEB_MAROON:          "Web Maroon",
		MARKUP_COLOR_MEDIUM_AQUAMARINE:   "Medium Aquamarine",
		MARKUP_COLOR_MEDIUM_BLUE:         "Medium Blue",
		MARKUP_COLOR_MEDIUM_ORCHID:       "Medium Orchid",
		MARKUP_COLOR_MEDIUM_PURPLE:       "Medium Purple",
		MARKUP_COLOR_MEDIUM_SEA_GREEN:    "Medium Sea Green",
		MARKUP_COLOR_MEDIUM_SLATE_BLUE:   "Medium Slate Blue",
		MARKUP_COLOR_MEDIUM_SPRING_GREEN: "Medium Spring Green",
		MARKUP_COLOR_MEDIUM_TURQUOISE:    "Medium Turquoise",
		MARKUP_COLOR_MEDIUM_VIOLET_RED:   "Medium Violet Red",
		MARKUP_COLOR_MIDNIGHT_BLUE:       "Midnight Blue",
		MARKUP_COLOR_MINT_CREAM:          "Mint Cream",
		MARKUP_COLOR_MISTY_ROSE:          "Misty Rose",
		MARKUP_COLOR_MOCCASIN:            "Moccasin",
		MARKUP_COLOR_NAVAJO_WHITE:        "Navajo White",
		MARKUP_COLOR_NAVY_BLUE:           "Navy Blue",
		MARKUP_COLOR_OLD_LACE:            "Old Lace",
		MARKUP_COLOR_OLIVE:               "Olive",
		MARKUP_COLOR_OLIVE_DRAB:          "Olive Drab",
		MARKUP_COLOR_ORANGE:              "Orange",
		MARKUP_COLOR_ORANGE_RED:          "Orange Red",
		MARKUP_COLOR_ORCHID:              "Orchid",
		MARKUP_COLOR_PALE_GOLDENROD:      "Pale Goldenrod",
		MARKUP_COLOR_PALE_GREEN:          "Pale Green",
		MARKUP_COLOR_PALE_TURQUOISE:      "Pale Turquoise",
		MARKUP_COLOR_PALE_VIOLET_RED:     "Pale Violet Red",
		MARKUP_COLOR_PAPAYA_WHIP:         "Papaya Whip",
		MARKUP_COLOR_PEACH_PUFF:          "Peach Puff",
		MARKUP_COLOR_PERU:                "Peru",
		MARKUP_COLOR_PINK:                "Pink",
		MARKUP_COLOR_PLUM:                "Plum",
		MARKUP_COLOR_POWDER_BLUE:         "Powder Blue",
		MARKUP_COLOR_PURPLE:              "Purple",
		MARKUP_COLOR_WEB_PURPLE:          "Web Purple",
		MARKUP_COLOR_REBECCA_PURPLE:      "Rebecca Purple",
		MARKUP_COLOR_RED:                 "Red",
		MARKUP_COLOR_ROSY_BROWN:          "Rosy Brown",
		MARKUP_COLOR_ROYAL_BLUE:          "Royal Blue",
		MARKUP_COLOR_SADDLE_BROWN:        "Saddle Brown",
		MARKUP_COLOR_SALMON:              "Salmon",
		MARKUP_COLOR_SANDY_BROWN:         "Sandy Brown",
		MARKUP_COLOR_SEA_GREEN:           "Sea Green",
		MARKUP_COLOR_SEASHELL:            "Seashell",
		MARKUP_COLOR_SIENNA:              "Sienna",
		MARKUP_COLOR_SILVER:              "Silver",
		MARKUP_COLOR_SKY_BLUE:            "Sky Blue",
		MARKUP_COLOR_SLATE_BLUE:          "Slate Blue",
		MARKUP_COLOR_SLATE_GRAY:          "Slate Gray",
		MARKUP_COLOR_SNOW:                "Snow",
		MARKUP_COLOR_SPRING_GREEN:        "Spring Green",
		MARKUP_COLOR_STEEL_BLUE:          "Steel Blue",
		MARKUP_COLOR_TAN:                 "Tan",
		MARKUP_COLOR_TEAL:                "Teal",
		MARKUP_COLOR_THISTLE:             "Thistle",
		MARKUP_COLOR_TOMATO:              "Tomato",
		MARKUP_COLOR_TURQUOISE:           "Turquoise",
		MARKUP_COLOR_VIOLET:              "Violet",
		MARKUP_COLOR_WHEAT:               "Wheat",
		MARKUP_COLOR_WHITE:               "White",
		MARKUP_COLOR_WHITE_SMOKE:         "White Smoke",
		MARKUP_COLOR_YELLOW:              "Yellow",
		MARKUP_COLOR_YELLOW_GREEN:        "Yellow Green",
	}
	const template = "'%s'"
	if val, ok := m[v]; ok {
		return spew.Sprintf(template, val)
	}
	return ""
}
