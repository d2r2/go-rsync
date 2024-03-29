//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2022 Denis Dyakov <denis.dyakov@gma**.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package rsync

// ------------------------------------------------------------
// File contains message identifiers for localization purpose.
// Message identifier names is self-descriptive, so ordinary
// it's easy to understand what message is made for.
// Message ID is used to call translation functions from
// "locale" package.
// ------------------------------------------------------------

const (
	MsgRsyncCallFailedError                  = "RsyncCallFailedError"
	MsgRsyncProcessTerminatedError           = "RsyncProcessTerminatedError"
	MsgRsyncCannotFindFolderSizeOutputError  = "RsyncCannotFindFolderSizeOutputError"
	MsgRsyncCannotParseFolderSizeOutputError = "RsyncCannotParseFolderSizeOutputError"
	MsgRsyncExtractVersionAndProtocolError   = "RsyncExtractVersionAndProtocolError"
)
