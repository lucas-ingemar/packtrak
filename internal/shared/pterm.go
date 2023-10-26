package shared

import "github.com/pterm/pterm"

var (
	PtermWarning = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "",
		},
	}
	PtermFound = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "",
		},
	}
	PtermNotFound = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "",
		},
	}
)
