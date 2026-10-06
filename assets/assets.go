// Package assets embeds the game's data files into the executable so the
// game ships as a single .exe. It holds bytes only: decoding images belongs in
// internal/render and parsing balance values in internal/balance. It imports
// nothing but the standard library.
package assets

import _ "embed"

// BalanceTOML is the contents of balance.toml, the balance configuration.
//
//go:embed balance.toml
var BalanceTOML []byte
