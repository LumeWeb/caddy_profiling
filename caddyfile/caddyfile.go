// Package caddyfile provides Caddyfile syntax support for the caddy_profiling
// module. It registers global options for the pyroscope, profefe, and profiling
// apps, allowing configuration via Caddyfile global options block.
package caddyfile

import (
	"encoding/json"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	caddyfileadapter "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"

	"go.lumeweb.com/caddy_profiling/profefe"
	"go.lumeweb.com/caddy_profiling/profiling"
	"go.lumeweb.com/caddy_profiling/pyroscope"
	"go.lumeweb.com/caddy_profiling/types"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("pyroscope", parsePyroscopeOption)
	httpcaddyfile.RegisterGlobalOption("profefe", parseProfefeOption)
	httpcaddyfile.RegisterGlobalOption("profiling", parseProfilingOption)
}

// parsePyroscopeOption configures the pyroscope app from the Caddyfile
// global options block. Syntax:
//
//	pyroscope {
//	    server_address      <url>
//	    application_name    <name>
//	    auth_token           <token>
//	    basic_auth_user     <user>
//	    basic_auth_password <pass>
//	    tenant_id            <id>
//	    disable_gc_runs
//	    upload_rate          <duration>
//	    profile_types        <types...>
//	}
func parsePyroscopeOption(d *caddyfileadapter.Dispenser, _ any) (any, error) {
	d.Next() // consume option name

	app := new(pyroscope.App)
	params := new(types.Parameters)

	for d.NextBlock(0) {
		switch d.Val() {
		case "server_address":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.ServerAddress = d.Val()
		case "application_name":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.ApplicationName = d.Val()
		case "auth_token":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.AuthToken = d.Val()
		case "basic_auth_user":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.BasicAuthUser = d.Val()
		case "basic_auth_password":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.BasicAuthPassword = d.Val()
		case "tenant_id":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.TenantID = d.Val()
		case "disable_gc_runs":
			app.DisableGCRuns = true
		case "upload_rate":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return nil, d.Errf("parsing upload_rate: %v", err)
			}
			app.UploadRate = caddy.Duration(dur)
		case "profile_types":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return nil, d.ArgErr()
			}
			for _, arg := range args {
				params.ProfileTypes = append(params.ProfileTypes, types.ProfileType(arg))
			}
			app.Parameters = params
		default:
			return nil, d.Errf("unrecognized pyroscope option '%s'", d.Val())
		}
	}

	return httpcaddyfile.App{
		Name:  "pyroscope",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

// parseProfefeOption configures the profefe app from the Caddyfile
// global options block. Syntax:
//
//	profefe {
//	    address        <url>
//	    service         <name>
//	    timeout         <duration>
//	    profile_types   <types...>
//	}
func parseProfefeOption(d *caddyfileadapter.Dispenser, _ any) (any, error) {
	d.Next() // consume option name

	app := new(profefe.App)
	params := new(types.Parameters)

	for d.NextBlock(0) {
		switch d.Val() {
		case "address":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.Address = d.Val()
		case "service":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.Service = d.Val()
		case "timeout":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return nil, d.Errf("parsing timeout: %v", err)
			}
			app.Timeout = caddy.Duration(dur)
		case "profile_types":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return nil, d.ArgErr()
			}
			for _, arg := range args {
				params.ProfileTypes = append(params.ProfileTypes, types.ProfileType(arg))
			}
			app.Parameters = params
		default:
			return nil, d.Errf("unrecognized profefe option '%s'", d.Val())
		}
	}

	return httpcaddyfile.App{
		Name:  "profefe",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

// parseProfilingOption configures the profiling parent app from the Caddyfile
// global options block. The profiling app hosts child profilers (pyroscope,
// profefe) with shared profiling parameters. Syntax:
//
//	profiling {
//	    cpu_profile_rate       <int>
//	    block_profile_rate     <int>
//	    mutex_profile_fraction <int>
//	    profile_types          <types...>
//
//	    pyroscope {
//	        server_address      <url>
//	        application_name    <name>
//	        auth_token           <token>
//	        basic_auth_user     <user>
//	        basic_auth_password <pass>
//	        tenant_id            <id>
//	        disable_gc_runs
//	        upload_rate          <duration>
//	        profile_types        <types...>
//	    }
//
//	    profefe {
//	        address        <url>
//	        service         <name>
//	        timeout         <duration>
//	        profile_types   <types...>
//	    }
//	}
func parseProfilingOption(d *caddyfileadapter.Dispenser, _ any) (any, error) {
	d.Next() // consume option name

	app := new(profiling.App)
	// params is set on the app's Parameters field, then inherited by child profilers
	params := new(types.Parameters)

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "cpu_profile_rate":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			n, err := strconv.Atoi(d.Val())
			if err != nil {
				return nil, d.Errf("invalid integer '%s': %v", d.Val(), err)
			}
			params.CPUProfileRate = n
		case "block_profile_rate":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			n, err := strconv.Atoi(d.Val())
			if err != nil {
				return nil, d.Errf("invalid integer '%s': %v", d.Val(), err)
			}
			params.BlockProfileRate = n
		case "mutex_profile_fraction":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			n, err := strconv.Atoi(d.Val())
			if err != nil {
				return nil, d.Errf("invalid integer '%s': %v", d.Val(), err)
			}
			params.MutexProfileFraction = n
		case "profile_types":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return nil, d.ArgErr()
			}
			for _, arg := range args {
				params.ProfileTypes = append(params.ProfileTypes, types.ProfileType(arg))
			}
		case "pyroscope":
			pyrApp, err := parseNestedPyroscope(d, params)
			if err != nil {
				return nil, err
			}
			raw, err := wrapProfiler("pyroscope", pyrApp)
			if err != nil {
				return nil, err
			}
			app.ProfilersRaw = append(app.ProfilersRaw, raw)
		case "profefe":
			pfApp, err := parseNestedProfefe(d, params)
			if err != nil {
				return nil, err
			}
			raw, err := wrapProfiler("profefe", pfApp)
			if err != nil {
				return nil, err
			}
			app.ProfilersRaw = append(app.ProfilersRaw, raw)
		default:
			return nil, d.Errf("unrecognized profiling option '%s'", d.Val())
		}
	}

	app.Parameters = *params

	return httpcaddyfile.App{
		Name:  "profiling",
		Value: caddyconfig.JSON(app, nil),
	}, nil
}

// wrapProfiler wraps a profiler app config as a json.RawMessage suitable for
// the profiling app's ProfilersRaw field, using the inline_key=profiler format.
// The profiler struct fields must be flattened at the top level alongside the
// "profiler" key — not nested under an "App" key — because Caddy's inline_key
// unmarshaling expects them as siblings.
func wrapProfiler(name string, appConfig any) (json.RawMessage, error) {
	raw, err := json.Marshal(appConfig)
	if err != nil {
		return nil, err
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	m["profiler"] = caddyconfig.JSON(name, nil)
	return json.Marshal(m)
}

// parseNestedPyroscope parses a pyroscope sub-block within the profiling app.
// It inherits shared parameters from the parent profiling block.
func parseNestedPyroscope(d *caddyfileadapter.Dispenser, parentParams *types.Parameters) (*pyroscope.App, error) {
	app := new(pyroscope.App)
	// inherit parent parameters
	inherited := *parentParams
	app.Parameters = &inherited

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		directive := d.Val()
		switch directive {
		case "server_address":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.ServerAddress = d.Val()
		case "application_name":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.ApplicationName = d.Val()
		case "auth_token":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.AuthToken = d.Val()
		case "basic_auth_user":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.BasicAuthUser = d.Val()
		case "basic_auth_password":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.BasicAuthPassword = d.Val()
		case "tenant_id":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.TenantID = d.Val()
		case "disable_gc_runs":
			app.DisableGCRuns = true
		case "upload_rate":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return nil, d.Errf("parsing upload_rate: %v", err)
			}
			app.UploadRate = caddy.Duration(dur)
		case "profile_types":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return nil, d.ArgErr()
			}
			// override inherited profile types
			localParams := *parentParams
			localParams.ProfileTypes = nil
			for _, arg := range args {
				localParams.ProfileTypes = append(localParams.ProfileTypes, types.ProfileType(arg))
			}
			app.Parameters = &localParams
		default:
			return nil, d.Errf("unrecognized pyroscope option '%s'", directive)
		}
	}

	return app, nil
}

// parseNestedProfefe parses a profefe sub-block within the profiling app.
// It inherits shared parameters from the parent profiling block.
func parseNestedProfefe(d *caddyfileadapter.Dispenser, parentParams *types.Parameters) (*profefe.App, error) {
	app := new(profefe.App)
	// inherit parent parameters
	inherited := *parentParams
	app.Parameters = &inherited

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		directive := d.Val()
		switch directive {
		case "address":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.Address = d.Val()
		case "service":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			app.Service = d.Val()
		case "timeout":
			if !d.NextArg() {
				return nil, d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return nil, d.Errf("parsing timeout: %v", err)
			}
			app.Timeout = caddy.Duration(dur)
		case "profile_types":
			args := d.RemainingArgs()
			if len(args) == 0 {
				return nil, d.ArgErr()
			}
			localParams := *parentParams
			localParams.ProfileTypes = nil
			for _, arg := range args {
				localParams.ProfileTypes = append(localParams.ProfileTypes, types.ProfileType(arg))
			}
			app.Parameters = &localParams
		default:
			return nil, d.Errf("unrecognized profefe option '%s'", directive)
		}
	}

	return app, nil
}
