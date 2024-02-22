package shared

import (
	"fmt"

	"github.com/pterm/pterm"
)

var (
	PtermWarning = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "",
		},
	}
	PtermInstalled = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "",
		},
	}
	PtermUpdated = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "󱍷",
		},
	}
	PtermMissing = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoMessageStyle,
			Text:  "",
		},
	}
	PtermRemoved = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "",
		},
	}
	PtermRed = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.ErrorMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.ErrorMessageStyle,
			Text:  "",
		},
	}
	PtermYellow = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.WarningMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.WarningMessageStyle,
			Text:  "",
		},
	}
	PtermGreen = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.SuccessMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.SuccessMessageStyle,
			Text:  "",
		},
	}
	PtermBlue = pterm.PrefixPrinter{
		MessageStyle: &pterm.ThemeDefault.InfoMessageStyle,
		Prefix: pterm.Prefix{
			Style: &pterm.ThemeDefault.InfoMessageStyle,
			Text:  "",
		},
	}

	PtermTablePrinter = pterm.TablePrinter{
		Style:                   &pterm.ThemeDefault.TableStyle,
		HasHeader:               false,
		HeaderStyle:             &pterm.ThemeDefault.TableHeaderStyle,
		HeaderRowSeparator:      "",
		HeaderRowSeparatorStyle: &pterm.ThemeDefault.TableSeparatorStyle,
		Separator:               "  ",
		SeparatorStyle:          &pterm.ThemeDefault.TableSeparatorStyle,
		RowSeparator:            "",
		RowSeparatorStyle:       &pterm.ThemeDefault.TableSeparatorStyle,
		Data:                    [][]string{},
		Boxed:                   false,
		LeftAlignment:           true,
		RightAlignment:          false,
		Writer:                  nil,
	}
)

type ptermMsgs struct {
	Start   string
	Success string
	Fail    string
}

type PtermSpinnerStatus string

const (
	PtermSpinnerInstall PtermSpinnerStatus = "install"
	PtermSpinnerUpdate  PtermSpinnerStatus = "update"
	PtermSpinnerRemove  PtermSpinnerStatus = "remove"
)

var PtermSpinnerStatusMsgs map[PtermSpinnerStatus]ptermMsgs = map[PtermSpinnerStatus]ptermMsgs{
	PtermSpinnerInstall: {
		Start:   "Installing %s...",
		Success: "%s installed successfully",
		Fail:    "%s failed to install: %s",
	},
	PtermSpinnerUpdate: {
		Start:   "Updating %s...",
		Success: "%s updated successfully",
		Fail:    "%s failed to update: %s",
	},
	PtermSpinnerRemove: {
		Start:   "Removing %s...",
		Success: "%s removed successfully",
		Fail:    "%s failed to be removed: %s",
	},
}

func PtermSpinner(spinnerStatus PtermSpinnerStatus, itemName string, f func() error) error {
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf(PtermSpinnerStatusMsgs[spinnerStatus].Start, itemName))
	spinner.SuccessPrinter = &PtermInstalled
	spinner.FailPrinter = &PtermRemoved
	if err := f(); err != nil {
		spinner.Fail(fmt.Sprintf(PtermSpinnerStatusMsgs[spinnerStatus].Fail, itemName, err.Error())) // Resolve spinner with error message.
		return err
	}
	spinner.Success(fmt.Sprintf(PtermSpinnerStatusMsgs[spinnerStatus].Success, itemName)) // Resolve spinner with error message.
	return nil
}
