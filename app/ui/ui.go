package ui

import (
	"fmt"

	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	banking "github.com/tomwheeler/demo-bank/app/bank"
)

type myTheme struct{}

var (
	sClient                                 *banking.BankClient
	rClient                                 *banking.BankClient
	window                                  fyne.Window
	senderBankLabel, recipientBankLabel     *widget.Label
	senderBankBalance, recipientBankBalance *widget.Label
	senderBankStatus, recipientBankStatus   *widget.Label
)

// BuildUI creates the Banking UI, showing (and constantly updating)
// details of the sender's and recipient's banks.
func BuildUI(senderClient *banking.BankClient, recipientClient *banking.BankClient) {
	sClient = senderClient
	rClient = recipientClient

	app := app.New()
	app.Settings().SetTheme(&myTheme{})

	window = app.NewWindow("Money Transfer App - UI")

	sName, err := sClient.GetName()
	if err != nil {
		sName = "UNKNOWN"
	}

	rName, err := rClient.GetName()
	if err != nil {
		rName = "UNKNOWN"
	}

	senderBankLabel = widget.NewLabel(fmt.Sprintf("%s bank: ", sName))
	senderBankLabel.TextStyle = fyne.TextStyle{Bold: true}
	senderBankBalance = widget.NewLabel("Balance: $????.??")
	senderBankBalance.TextStyle = fyne.TextStyle{Bold: true}
	senderBankStatus = widget.NewLabel("Service Status: Unknown")
	senderBankStatus.Importance = widget.MediumImportance
	senderBankStatus.TextStyle = fyne.TextStyle{Bold: true}

	recipientBankLabel = widget.NewLabel(fmt.Sprintf("%s bank: ", rName))
	recipientBankLabel.TextStyle = fyne.TextStyle{Bold: true}
	recipientBankBalance = widget.NewLabel("Balance: $????.??")
	recipientBankBalance.TextStyle = fyne.TextStyle{Bold: true}
	recipientBankStatus = widget.NewLabel("Service Status: Unknown")
	recipientBankStatus.Importance = widget.MediumImportance
	recipientBankStatus.TextStyle = fyne.TextStyle{Bold: true}

	grid := container.New(layout.NewGridLayout(3),
		senderBankLabel, senderBankBalance, senderBankStatus,
		recipientBankLabel, recipientBankBalance, recipientBankStatus,
	)

	go func() {
		tick := time.Tick(500 * time.Millisecond) // increase rate?
		for range tick {

			sName, err := sClient.GetName()
			if err == nil {
				updateSenderName(sName)
			}

			rName, err := rClient.GetName()
			if err == nil {
				updateRecipientName(rName)
			}

			sBalance, err := sClient.GetBalance()
			if err == nil {
				updateSenderBalance(sBalance)
			}

			rBalance, err := rClient.GetBalance()
			if err == nil {
				updateRecipientBalance(rBalance)
			}

			if sClient.IsServiceRunning() {
				markSenderBankOnline()
			} else {
				markSenderBankOffline()
			}

			if rClient.IsServiceRunning() {
				markRecipientBankOnline()
			} else {
				markRecipientBankOffline()
			}
		}
	}()

	window.SetContent(grid)
	window.ShowAndRun()
}

func updateSenderName(name string) {
	senderBankLabel.SetText(fmt.Sprintf("%s bank: ", name))
	window.Content().Refresh()
}

func updateRecipientName(name string) {
	recipientBankLabel.SetText(fmt.Sprintf("%s bank: ", name))
	window.Content().Refresh()
}

func updateSenderBalance(newBalance int) {
	senderBankBalance.SetText(fmt.Sprintf("Balance: $%d.00", newBalance))
	window.Content().Refresh()
}

func updateRecipientBalance(newBalance int) {
	recipientBankBalance.SetText(fmt.Sprintf("Balance: $%d.00", newBalance))
	window.Content().Refresh()
}

func markSenderBankOffline() {
	senderBankStatus.SetText("Service Status: Offline")
	senderBankStatus.Importance = widget.DangerImportance
	window.Content().Refresh()
}

func markRecipientBankOffline() {
	recipientBankStatus.SetText("Service Status: Offline")
	recipientBankStatus.Importance = widget.DangerImportance
	window.Content().Refresh()
}

func markSenderBankOnline() {
	senderBankStatus.SetText("Service Status: Online")
	senderBankStatus.Importance = widget.SuccessImportance
	window.Content().Refresh()
}

func markRecipientBankOnline() {
	recipientBankStatus.SetText("Service Status: Online")
	recipientBankStatus.Importance = widget.SuccessImportance
	window.Content().Refresh()
}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameSuccess {
		return color.NRGBA{R: 0, G: 165, B: 40, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (myTheme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNameCaptionText:
		return 18
	case theme.SizeNameInlineIcon:
		return 28
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameScrollBar:
		return 18
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 28
	case theme.SizeNameInputBorder:
		return 2
	default:
		return theme.DefaultTheme().Size(s)
	}
}
