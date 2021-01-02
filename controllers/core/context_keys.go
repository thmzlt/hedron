/*
Unlicensed
*/

package controllers

type contextKey string

var (
	contextKeyRequest  = contextKey("Request")
	contextKeyProject  = contextKey("Project")
	contextKeyRevision = contextKey("Revision")
)
