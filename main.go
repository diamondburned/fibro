//go:generate fyne bundle -o bundled.go Icon.png

package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/diamondburned/arikawa/v2/session"
)

const prefTokenKey = "auth.token"

func main() {
	a := app.NewWithID("xyz.andy.fibro")
	a.SetIcon(resourceIconPng)
	w := a.NewWindow("Fibro: Discord")

	u := &ui{}
	w.SetContent(u.makeUI())
	w.Resize(fyne.NewSize(480, 320))
	go login(w, a.Preferences(), u)
	w.ShowAndRun()

	// after app quits
	if u.conn != nil {
		_ = u.conn.Close()
	}
}

func login(w fyne.Window, p fyne.Preferences, u *ui) {
	tok := p.String(prefTokenKey)
	if tok != "" {
		sess, err := session.New(tok)
		if err == nil {
			loadServers(sess, u)
			return
		} else {
			log.Println("Error connecting with token", err)
		}
	}

	email := widget.NewEntry()
	pass := widget.NewPasswordEntry()
	dialog.ShowForm("Log in to Discord", "Log in", "Use Token instead",
		[]*widget.FormItem{
			{Text: "Email", Widget: email},
			{Text: "Password", Widget: pass},
		}, func(ok bool) {
			if ok {
				doLogin(email.Text, pass.Text, w, p, u)
			} else {
				showTokenForm(w, p, u)
			}
		}, w)
}

func showTokenForm(w fyne.Window, p fyne.Preferences, u *ui) {
	token := widget.NewPasswordEntry()
	dialog.ShowForm("Log in to Discord using a token", "Log in", "Cancel",
		[]*widget.FormItem{
			{Text: "Token", Widget: token},
		}, func(ok bool) {
			if ok {
				s, err := session.New(token.Text)
				if err != nil {
					log.Println("Login Err", err)
					return
				}
				p.SetString(prefTokenKey, token.Text)
				loadServers(s, u)
			}
		}, w)
}

func doLogin(email, pass string, w fyne.Window, p fyne.Preferences, u *ui) {
	sess, err := session.Login(email, pass, "")
	if err == nil {
		p.SetString(prefTokenKey, sess.Token)
		loadServers(sess, u)
		return
	}

	if err != session.ErrMFA {
		log.Println("Login Err", err)
		return
	}

	mfa := widget.NewEntry()
	dialog.ShowForm("Multi-Factor required", "Confirm", "Cancel",
		[]*widget.FormItem{
			{Text: "Please enter your MFA token", Widget: mfa},
		},
		func(ok bool) {
			if !ok {
				return
			}
			sess, err := session.Login(email, pass, mfa.Text)
			if err != nil {
				log.Println("Failure in MFA verification")
				return
			}

			p.SetString(prefTokenKey, sess.Token)
			loadServers(sess, u)
		}, w)
}
