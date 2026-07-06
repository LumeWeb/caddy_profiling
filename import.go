package caddy_profiling

// Import subpackages so their init() functions register Caddy modules and
// Caddyfile global options when this module is imported via xcaddy --with.
import (
	_ "go.lumeweb.com/caddy_profiling/caddyfile"
	_ "go.lumeweb.com/caddy_profiling/profefe"
	_ "go.lumeweb.com/caddy_profiling/profiling"
	_ "go.lumeweb.com/caddy_profiling/pyroscope"
)
