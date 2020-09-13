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
	"fmt"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
)

const (
	APP_LICENSE = `                   GNU LESSER GENERAL PUBLIC LICENSE
                       Version 3, 29 June 2007

 Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.


  This version of the GNU Lesser General Public License incorporates
the terms and conditions of version 3 of the GNU General Public
License, supplemented by the additional permissions listed below.

  0. Additional Definitions.

  As used herein, "this License" refers to version 3 of the GNU Lesser
General Public License, and the "GNU GPL" refers to version 3 of the GNU
General Public License.

  "The Library" refers to a covered work governed by this License,
other than an Application or a Combined Work as defined below.

  An "Application" is any work that makes use of an interface provided
by the Library, but which is not otherwise based on the Library.
Defining a subclass of a class defined by the Library is deemed a mode
of using an interface provided by the Library.

  A "Combined Work" is a work produced by combining or linking an
Application with the Library.  The particular version of the Library
with which the Combined Work was made is also called the "Linked
Version".

  The "Minimal Corresponding Source" for a Combined Work means the
Corresponding Source for the Combined Work, excluding any source code
for portions of the Combined Work that, considered in isolation, are
based on the Application, and not on the Linked Version.

  The "Corresponding Application Code" for a Combined Work means the
object code and/or source code for the Application, including any data
and utility programs needed for reproducing the Combined Work from the
Application, but excluding the System Libraries of the Combined Work.

  1. Exception to Section 3 of the GNU GPL.

  You may convey a covered work under sections 3 and 4 of this License
without being bound by section 3 of the GNU GPL.

  2. Conveying Modified Versions.

  If you modify a copy of the Library, and, in your modifications, a
facility refers to a function or data to be supplied by an Application
that uses the facility (other than as an argument passed when the
facility is invoked), then you may convey a copy of the modified
version:

   a) under this License, provided that you make a good faith effort to
   ensure that, in the event an Application does not supply the
   function or data, the facility still operates, and performs
   whatever part of its purpose remains meaningful, or

   b) under the GNU GPL, with none of the additional permissions of
   this License applicable to that copy.

  3. Object Code Incorporating Material from Library Header Files.

  The object code form of an Application may incorporate material from
a header file that is part of the Library.  You may convey such object
code under terms of your choice, provided that, if the incorporated
material is not limited to numerical parameters, data structure
layouts and accessors, or small macros, inline functions and templates
(ten or fewer lines in length), you do both of the following:

   a) Give prominent notice with each copy of the object code that the
   Library is used in it and that the Library and its use are
   covered by this License.

   b) Accompany the object code with a copy of the GNU GPL and this license
   document.

  4. Combined Works.

  You may convey a Combined Work under terms of your choice that,
taken together, effectively do not restrict modification of the
portions of the Library contained in the Combined Work and reverse
engineering for debugging such modifications, if you also do each of
the following:

   a) Give prominent notice with each copy of the Combined Work that
   the Library is used in it and that the Library and its use are
   covered by this License.

   b) Accompany the Combined Work with a copy of the GNU GPL and this license
   document.

   c) For a Combined Work that displays copyright notices during
   execution, include the copyright notice for the Library among
   these notices, as well as a reference directing the user to the
   copies of the GNU GPL and this license document.

   d) Do one of the following:

       0) Convey the Minimal Corresponding Source under the terms of this
       License, and the Corresponding Application Code in a form
       suitable for, and under terms that permit, the user to
       recombine or relink the Application with a modified version of
       the Linked Version to produce a modified Combined Work, in the
       manner specified by section 6 of the GNU GPL for conveying
       Corresponding Source.

       1) Use a suitable shared library mechanism for linking with the
       Library.  A suitable mechanism is one that (a) uses at run time
       a copy of the Library already present on the user's computer
       system, and (b) will operate properly with a modified version
       of the Library that is interface-compatible with the Linked
       Version.

   e) Provide Installation Information, but only if you would otherwise
   be required to provide such information under section 6 of the
   GNU GPL, and only to the extent that such information is
   necessary to install and execute a modified version of the
   Combined Work produced by recombining or relinking the
   Application with a modified version of the Linked Version. (If
   you use option 4d0, the Installation Information must accompany
   the Minimal Corresponding Source and Corresponding Application
   Code. If you use option 4d1, you must provide the Installation
   Information in the manner specified by section 6 of the GNU GPL
   for conveying Corresponding Source.)

  5. Combined Libraries.

  You may place library facilities that are a work based on the
Library side by side in a single library together with other library
facilities that are not Applications and are not covered by this
License, and convey such a combined library under terms of your
choice, if you do both of the following:

   a) Accompany the combined library with a copy of the same work based
   on the Library, uncombined with any other library facilities,
   conveyed under the terms of this License.

   b) Give prominent notice with the combined library that part of it
   is a work based on the Library, and explaining where to find the
   accompanying uncombined form of the same work.

  6. Revised Versions of the GNU Lesser General Public License.

  The Free Software Foundation may publish revised and/or new versions
of the GNU Lesser General Public License from time to time. Such new
versions will be similar in spirit to the present version, but may
differ in detail to address new problems or concerns.

  Each version is given a distinguishing version number. If the
Library as you received it specifies that a certain numbered version
of the GNU Lesser General Public License "or any later version"
applies to it, you have the option of following the terms and
conditions either of that published version or of any later version
published by the Free Software Foundation. If the Library as you
received it does not specify a version number of the GNU Lesser
General Public License, you may choose any version of the GNU Lesser
General Public License ever published by the Free Software Foundation.

  If the Library as you received it specifies that a proxy can decide
whether future versions of the GNU Lesser General Public License shall
apply, that proxy's public statement of acceptance of any version is
permanent authorization for you to choose that version for the
Library.`
)

// buildCommentBlock build multiline comments block to show in About Dialog.
func buildCommentBlock() (*bytes.Buffer, error) {
	version, protocol, err := rsync.GetRsyncVersion()
	if err != nil {
		if rsync.IsExtractVersionAndProtocolError(err) {
			version = "?"
			protocol = version
		} else {
			return nil, err
		}
	}

	var buf bytes.Buffer
	glibMajor, glibMinor, glibMicro := GetGlibVersion()
	glibBuildVersion := glib.GetBuildVersion()
	gtkMajor, gtkMinor, gtkMicro := GetGtkVersion()
	gtkBuildVersion := gtk.GetBuildVersion()
	buf.WriteString(fmt.Sprintln(locale.T(MsgAboutDlgAppDescriptionSection, nil)))
	buf.WriteString(fmt.Sprintln(locale.T(MsgAppEnvironmentTitle, nil)))
	buf.WriteString(fmt.Sprintln(fmt.Sprintf("%s.",
		locale.T(MsgGLIBInfo, struct{ GLIBCompiledVer, GLIBDetectedVer string }{
			GLIBCompiledVer: spew.Sprintf("%s", glibBuildVersion),
			GLIBDetectedVer: spew.Sprintf("%d.%d.%d", glibMajor, glibMinor, glibMicro)}))))
	buf.WriteString(fmt.Sprintln(fmt.Sprintf("%s.",
		locale.T(MsgGTKInfo, struct{ GTKCompiledVer, GTKDetectedVer string }{
			GTKCompiledVer: spew.Sprintf("%s", gtkBuildVersion),
			GTKDetectedVer: spew.Sprintf("%d.%d.%d", gtkMajor, gtkMinor, gtkMicro)}))))
	buf.WriteString(fmt.Sprintln(fmt.Sprintf("%s.",
		locale.T(MsgRsyncInfo, struct{ RSYNCDetectedVer, RSYNCDetectedProtocol string }{
			RSYNCDetectedVer: version, RSYNCDetectedProtocol: protocol}))))
	//buf.WriteString("<a href=\"https://maps.google.com/maps?q=62.1891,+-141.5372+(Example+text+in+here+will+be+rendered+in+the+maps+label)&amp;iwloc=A&amp;hl=en\" title=\"Map\">Click here for Map</a>")

	/*
		display, err := gdk.DisplayGetDefault()
		if err != nil {
			return nil, err
		}
		if gdk.IsWaylandDisplay(display) {
			buf.WriteString("WAYLAND display detected.")
			buf.WriteString(fmt.Sprintln())
		} else if gdk.IsX11Display(display) {
			buf.WriteString("X11 display detected.")
			buf.WriteString(fmt.Sprintln())
		}
	*/

	buf.WriteString(fmt.Sprintln(fmt.Sprintf("%s.",
		locale.T(MsgGolangInfo, struct{ GolangVersion, AppArchitecture string }{
			GolangVersion:   core.GetGolangVersion(),
			AppArchitecture: core.GetAppArchitecture()}))))
	buf.WriteString(fmt.Sprintln())
	buf.WriteString(fmt.Sprintln(locale.T(MsgAboutDlgAppFeaturesAndBenefitsTitle, nil)))
	buf.WriteString(fmt.Sprintln(locale.T(MsgAboutDlgAppFeaturesAndBenefitsSection, nil)))
	buf.WriteString(fmt.Sprintln(locale.T(MsgAboutDlgReleasedUnderLicense,
		struct{ LicenseName string }{LicenseName: "GNU LGPL v3.0"})))
	buf.WriteString(fmt.Sprintln())
	buf.WriteString(fmt.Sprintln(locale.T(MsgAboutDlgFollowMyGithubProjectTitle, nil)))
	buf.WriteString(fmt.Sprintln("https://github.com/d2r2?tab=repositories"))

	return &buf, nil
}

// CreateAboutDialog creates about dialog object.
func CreateAboutDialog(appSettings *SettingsStore) (*gtk.AboutDialog, error) {
	dlg, err := gtk.AboutDialogNew()
	if err != nil {
		return nil, err
	}

	dlg.SetProgramName(core.GetAppFullTitle())
	dlg.SetLogoIconName("media-tape-symbolic")
	dlg.SetVersion(core.GetAppVersion())
	dlg.SetCopyright(locale.T(MsgAboutDlgAppCopyright,
		struct{ AppCreationYears, AppCopyrightAuthor string }{
			AppCreationYears:   "2017-2020",
			AppCopyrightAuthor: "Denis Dyakov <denis.dyakov@gmail.com>"}))
	dlg.SetAuthors(core.SplitByEOL(locale.T(MsgAboutDlgAppAuthorsBlock, nil)))

	dlg.SetLicense(APP_LICENSE)

	bh := appSettings.NewBindingHelper()
	// Show about dialog on application startup
	cbAboutInfo, err := gtk.CheckButtonNewWithLabel(locale.T(MsgAboutDlgDoNotShowCaption, nil))
	if err != nil {
		return nil, err
	}
	bh.Bind(CFG_DONT_SHOW_ABOUT_ON_STARTUP, cbAboutInfo, "active", glib.SETTINGS_BIND_DEFAULT)

	content, err := dlg.GetContentArea()
	if err != nil {
		return nil, err
	}
	content.Add(cbAboutInfo)
	content.ShowAll()

	buf, err := buildCommentBlock()
	if err != nil {
		return nil, err
	}
	dlg.SetComments(buf.String())

	dlg.SetWebsite("https://gorsync.github.io/")
	dlg.SetWebsiteLabel(locale.T(MsgAboutDlgAppLearnMore,
		struct{ AppTitle string }{
			AppTitle: core.GetAppTitle()}))

	return dlg, nil
}
