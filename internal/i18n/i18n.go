package i18n

// HelpI18n contains help text for all commands
type HelpI18n struct {
	Summary  string
	Help     string
	New      string
	Start    string
	Img      string
	Version  string
	Setenv   string
	Setenvs  string
	Delenv   string
	Clearenv string
	System   string
	Redo     string
	Models   string
	Echo     string
}

// I18n contains all internationalized strings
type I18n struct {
	Env struct {
		SystemInitMessage string
	}
	Command struct {
		Help HelpI18n
		New  struct {
			NewChatStart string
		}
	}
	CallbackQuery struct {
		OpenModelList  string
		SelectProvider string
		SelectModel    string
		ChangeModel    string
	}
}

// LoadI18n loads the appropriate language based on the language code
func LoadI18n(lang string) *I18n {
	switch lang {
	case "cn", "zh-cn", "zh-hans":
		return zhHans()
	case "zh-tw", "zh-hk", "zh-mo", "zh-hant":
		return zhHant()
	case "pt", "pt-br":
		return pt()
	case "en", "en-us":
		return en()
	default:
		return en()
	}
}
