// Package balance parses the balance configuration (assets/balance.toml) into
// a typed Config. It has no game logic and imports nothing from this module
// except assets.
//
// Decoding is strict: a key in the file with no matching field in Config is
// an error, so a value the designer adds is never silently ignored. Every
// field added to Config needs a matching documented key in balance.toml.
package balance

import (
	"fmt"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds every tunable value. Fields are added by the gameplay engineer
// as the designer adds keys to balance.toml.
type Config struct{}

// Parse decodes TOML bytes into a Config, rejecting unknown keys.
func Parse(data []byte) (Config, error) {
	var cfg Config
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return Config{}, err
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, k := range undecoded {
			keys[i] = k.String()
		}
		slices.Sort(keys)
		return Config{}, fmt.Errorf("unknown balance keys: %s", strings.Join(keys, ", "))
	}
	return cfg, nil
}
