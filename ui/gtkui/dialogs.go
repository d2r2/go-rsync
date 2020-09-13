//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2020 Denis Dyakov <denis.dyakov@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package gtkui

import (
	"bytes"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/gotk3/gtk"
	"github.com/d2r2/gotk3/pango"
)

// schemaSettingsErrorDialog display error related to GLIB GSettings application configuration.
func schemaSettingsErrorDialog(parent *gtk.Window, text string, extraMsg *string) error {
	//title := "<span weight='bold' size='larger'>Schema settings configuration error</span>"
	titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
			NewMarkup(MARKUP_SIZE_LARGER, 0, 0, locale.T(MsgSchemaConfigDlgTitle, nil), nil)))
	paragraphs := []*DialogParagraph{NewDialogParagraph(text).
		SetJustify(gtk.JUSTIFY_CENTER).SetHorizAlign(gtk.ALIGN_CENTER)}
	if extraMsg != nil {
		paragraphs = append(paragraphs, NewDialogParagraph(*extraMsg).
			SetJustify(gtk.JUSTIFY_CENTER).SetHorizAlign(gtk.ALIGN_CENTER))
	}

	err := ErrorMessage(parent, titleMarkup.String(), paragraphs)
	if err != nil {
		return err
	}
	return nil
}

func removeUndescore(buttonCaption string) string {
	var buf bytes.Buffer
	for _, ch := range buttonCaption {
		if ch == '_' {
			continue
		}
		buf.WriteRune(ch)
	}

	buttonCaption = buf.String()
	return buttonCaption
}

func createInterruptBackupDialog(parent *gtk.Window) (*MessageDialog, error) {
	title := locale.T(MsgAppWindowTerminateBackupDlgTitle, nil)
	titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
	terminateButtonCaption1 := locale.T(MsgDialogYesButton, nil)
	terminateButtonMarkup1 := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
		removeUndescore(terminateButtonCaption1), nil)
	continueButtonCaption1 := locale.T(MsgDialogNoButton, nil)
	continueButtonMarkup1 := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
		removeUndescore(continueButtonCaption1), nil)
	escapeKeyMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, "esc", nil))
	text := locale.T(MsgAppWindowTerminateBackupDlgText,
		struct{ TerminateButton, ContinueButton, EscapeKey string }{
			TerminateButton: terminateButtonMarkup1.String(),
			ContinueButton:  continueButtonMarkup1.String(),
			EscapeKey:       escapeKeyMarkup.String()})
	// textMarkup := NewMarkup(0, 0, 0, text, nil)

	buttons := []DialogButton{
		{terminateButtonCaption1, gtk.RESPONSE_YES, false, func(btn *gtk.Button) error {
			style, err2 := btn.GetStyleContext()
			if err2 != nil {
				return err2
			}
			// style.AddClass("suggested-action")
			style.AddClass("destructive-action")
			return nil
		}},
		{continueButtonCaption1, gtk.RESPONSE_NO, true, func(btn *gtk.Button) error {
			style, err2 := btn.GetStyleContext()
			if err2 != nil {
				return err2
			}
			style.AddClass("suggested-action")
			// style.AddClass("destructive-action")
			return nil
		}},
	}
	dialog, err := SetupMessageDialog(parent, titleMarkup.String(), "",
		[]*DialogParagraph{NewDialogParagraph(text).SetMarkup(true)}, buttons, nil)
	return dialog, err
}

// interruptBackupDialog shows dialog and query for active process termination.
func interruptBackupDialog(parent *gtk.Window) (bool, error) {
	dialog, err := createInterruptBackupDialog(parent)
	if err != nil {
		return false, err
	}

	response := dialog.Run(false)
	PrintDialogResponse(response)
	return IsResponseYes(response), nil
}

// OutOfSpaceResponse denote response from RSYNC out of space dialog query.
type OutOfSpaceResponse int

// 3 response type from RSYNC out of space dialog query:
// 1) retry RSYNC failed call;
// 2) ignore RSYNC filed call, but continue backup process;
// 3) immediately terminate backup process.
const (
	OutOfSpaceRetry OutOfSpaceResponse = iota
	OutOfSpaceIgnore
	OutOfSpaceTerminate
)

// outOfSpaceDialogAsync show dialog once RSYNC out of space issue happens.
func outOfSpaceDialogAsync(parent *gtk.Window, paths core.SrcDstPath, freeSpace uint64) (OutOfSpaceResponse, error) {
	title := locale.T(MsgAppWindowOutOfSpaceDlgTitle, nil)
	titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
	terminateButtonCaption := locale.T(MsgAppWindowOutOfSpaceDlgTerminateButton, nil)
	terminateButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
		removeUndescore(terminateButtonCaption), nil)
	ignoreButtonCaption := locale.T(MsgAppWindowOutOfSpaceDlgIgnoreButton, nil)
	ignoreButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
		removeUndescore(ignoreButtonCaption), nil)
	retryButtonCaption := locale.T(MsgAppWindowOutOfSpaceDlgRetryButton, nil)
	retryButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
		removeUndescore(retryButtonCaption), nil)
	escapeKeyMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, "esc", nil))
	buttons := []DialogButton{
		{retryButtonCaption, gtk.RESPONSE_YES, true, func(btn *gtk.Button) error {
			style, err2 := btn.GetStyleContext()
			if err2 != nil {
				return err2
			}
			style.AddClass("suggested-action")
			// style.AddClass("destructive-action")
			return nil
		}},
		{ignoreButtonCaption, gtk.RESPONSE_CANCEL, !true, nil},
		{terminateButtonCaption, gtk.RESPONSE_NO, !true, func(btn *gtk.Button) error {
			style, err2 := btn.GetStyleContext()
			if err2 != nil {
				return err2
			}
			//style.AddClass("suggested-action")
			style.AddClass("destructive-action")
			return nil
		}},
	}
	text := locale.T(MsgAppWindowOutOfSpaceDlgText1,
		struct{ Path, FreeSpace string }{Path: paths.DestPath,
			FreeSpace: core.FormatSize(freeSpace, true)})
	paragraphs := []*DialogParagraph{NewDialogParagraph(text).SetEllipsize(pango.ELLIPSIZE_MIDDLE).SetMaxWidthChars(10)}
	text = locale.T(MsgAppWindowOutOfSpaceDlgText2,
		struct{ EscapeKey, RetryButton, IgnoreButton, TerminateButton string }{EscapeKey: escapeKeyMarkup.String(),
			RetryButton: retryButtonMarkup.String(), IgnoreButton: ignoreButtonMarkup.String(),
			TerminateButton: terminateButtonMarkup.String()})
	paragraphs = append(paragraphs, NewDialogParagraph(text).SetMarkup(true).SetHorizAlign(gtk.ALIGN_CENTER))

	ch := make(chan gtk.ResponseType)
	defer close(ch)

	MustIdleAdd(func() {
		dialog, err2 := SetupMessageDialog(parent, titleMarkup.String(), "", paragraphs, buttons, nil)
		if err2 != nil {
			lg.Fatal(err2)
		}
		ch <- dialog.Run(false)
	})

	response, _ := <-ch
	PrintDialogResponse(response)

	if IsResponseYes(response) {
		return OutOfSpaceRetry, nil
	} else if IsResponseNo(response) {
		return OutOfSpaceTerminate, nil
	} else {
		return OutOfSpaceIgnore, nil
	}
}

// questionDialog shows standard question dialog with localizable YES/NO selection.
func questionDialog(parent *gtk.Window, titleMarkup string, textMarkup string,
	defaultNo bool, yesDestructive bool, noSuggested bool) (bool, error) {
	yesButtonCaption := locale.T(MsgDialogYesButton, nil)
	noButtonCaption := locale.T(MsgDialogNoButton, nil)
	// escapeKeyMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
	// 	NewMarkup(MARKUP_SIZE_LARGER, 0, 0, "esc", nil))
	buttons := []DialogButton{
		{yesButtonCaption, gtk.RESPONSE_YES, false, func(btn *gtk.Button) error {
			if yesDestructive {
				style, err2 := btn.GetStyleContext()
				if err2 != nil {
					return err2
				}
				// style.AddClass("suggested-action")
				style.AddClass("destructive-action")
			}
			return nil
		}},
		{noButtonCaption, gtk.RESPONSE_NO, defaultNo, func(btn *gtk.Button) error {
			if noSuggested {
				style, err2 := btn.GetStyleContext()
				if err2 != nil {
					return err2
				}
				style.AddClass("suggested-action")
				// style.AddClass("destructive-action")
			}
			return nil
		}},
	}
	dialog, err := SetupMessageDialog(parent, titleMarkup, "",
		[]*DialogParagraph{NewDialogParagraph(textMarkup).SetMarkup(true)}, buttons, nil)
	for {
		response := dialog.Run(false)
		if err != nil {
			return false, err
		}
		if IsResponseDeleteEvent(response) {
			if defaultNo {
				return false, nil
			}
		} else if IsResponseYes(response) {
			return true, nil
		} else if IsResponseNo(response) {
			return false, nil
		}
	}
}
