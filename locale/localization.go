package locale

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/d2r2/go-rsync/data"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// global variable to keep localizer object used to translate everything in application
var Localizer *i18n.Localizer

// One of 2 main methods to translate message ID text, using format
// functionality based on template interface.
var T = func(messageID string, template interface{}) string {
	// if Localizer isn't initialized, set up with system language
	if Localizer == nil {
		SetLanguage("")
	}
	// get localized message
	msg := Localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: template})
	return msg
}

// One of 2 main methods to translate message ID text, using format
// functionality based on template interface. Extra functionality
// allow to control plural form behavior.
var TP = func(messageID string, template interface{}, pluralCount interface{}) string {
	// if Localizer isn't initialized, set up with system language
	if Localizer == nil {
		SetLanguage("")
	}
	// get localized message
	msg := Localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: template,
		PluralCount:  pluralCount})
	return msg
}

func mustParseMessageFile(bundle *i18n.Bundle, assetIconName string) {
	file, err := data.Assets.Open(assetIconName)
	if err != nil {
		lg.Fatal(err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		lg.Fatal(err)
	}

	bundle.MustParseMessageFileBytes(buf, assetIconName)
}

// SetLanguage set up language globally for application localization.
func SetLanguage(lang string) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	mustParseMessageFile(bundle, "translate.en.toml")
	mustParseMessageFile(bundle, "translate.ru.toml")

	if lang == "" {
		lang = os.Getenv("LANG")
		// remove ".UTF-8" suffix from language if found, as "en-US.UTF-8"
		if i := strings.Index(lang, ".UTF-8"); i != -1 {
			lang = lang[:i]
		}
	}
	//Localizer = i18n.NewLocalizer(bundle, "en-US")
	Localizer = i18n.NewLocalizer(bundle, lang)
	// Test translation
	// fmt.Println(Localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "HelloWorld"}))
}
