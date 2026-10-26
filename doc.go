// Package openwa provides a synchronous HTTP client for the OpenWA API.
//
// A successful send confirms gateway acceptance, not delivery to the recipient.
// The client never automatically retries requests. Callers must reconcile a send
// returning ErrOutcomeUnknown before deciding whether to repeat it.
//
// Session creation and listing are explicit operations; the client never
// provisions automatically. It does not manage credentials, normalize regional
// phone numbers, or persist application state.
package openwa
