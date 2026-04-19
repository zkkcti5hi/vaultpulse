// Package filter provides lease filtering, sorting, grouping, and rendering utilities.
//
// RenderTemplate renders a slice of SecretLease values using a Go text/template
// string. If the template string is empty, a sensible default is used.
// Custom templates have access to all exported fields of SecretLease as well
// as the helper functions "upper" and "default".
package filter
