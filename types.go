package caddy_profiling

// Re-export types from the types package for backwards compatibility.
// Existing code importing caddy_profiling.Parameters etc. will continue to work.
import "go.lumeweb.com/caddy_profiling/types"

type Parameters = types.Parameters
type ProfileType = types.ProfileType

const (
	CPU          = types.CPU
	Goroutine    = types.Goroutine
	Heap         = types.Heap
	Allocs       = types.Allocs
	Threadcreate = types.Threadcreate
	Block        = types.Block
	Mutex        = types.Mutex
)

type ProfilingParameterSetter = types.ProfilingParameterSetter
type Profiler = types.Profiler
