package caddyfile

import (
	"encoding/json"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	caddyfileadapter "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func TestParsePyroscopeOption(t *testing.T) {
	for i, tc := range []struct {
		input    string
		wantErr  bool
		validate func(t *testing.T, app httpcaddyfile.App)
	}{
		{
			input: `pyroscope {
				server_address http://pyroscope:4040
				application_name caddy-proxy
				profile_types cpu heap allocs goroutine
			}`,
			wantErr: false,
			validate: func(t *testing.T, app httpcaddyfile.App) {
				if app.Name != "pyroscope" {
					t.Errorf("expected app name 'pyroscope', got '%s'", app.Name)
				}
				var cfg map[string]json.RawMessage
				if err := json.Unmarshal(app.Value, &cfg); err != nil {
					t.Fatalf("failed to unmarshal app config: %v", err)
				}
				var serverAddr string
				json.Unmarshal(cfg["server_address"], &serverAddr)
				if serverAddr != "http://pyroscope:4040" {
					t.Errorf("expected server_address 'http://pyroscope:4040', got '%s'", serverAddr)
				}
				var appName string
				json.Unmarshal(cfg["application_name"], &appName)
				if appName != "caddy-proxy" {
					t.Errorf("expected application_name 'caddy-proxy', got '%s'", appName)
				}
			},
		},
		{
			input: `pyroscope {
				server_address https://pyroscope.example.com
				application_name my-app
				auth_token secret-token
				basic_auth_user user
				basic_auth_password pass
				tenant_id tenant-123
				disable_gc_runs
				upload_rate 30s
				profile_types cpu
			}`,
			wantErr: false,
			validate: func(t *testing.T, app httpcaddyfile.App) {
				var cfg map[string]json.RawMessage
				if err := json.Unmarshal(app.Value, &cfg); err != nil {
					t.Fatalf("failed to unmarshal app config: %v", err)
				}
				var token string
				json.Unmarshal(cfg["auth_token"], &token)
				if token != "secret-token" {
					t.Errorf("expected auth_token 'secret-token', got '%s'", token)
				}
				var disableGC bool
				json.Unmarshal(cfg["disable_gc_runs"], &disableGC)
				if !disableGC {
					t.Error("expected disable_gc_runs to be true")
				}
			},
		},
		{
			input: `pyroscope {
				unknown_option foo
			}`,
			wantErr: true,
		},
		{
			input: `pyroscope {
				server_address http://pyroscope:4040
			}`,
			wantErr: false,
			validate: func(t *testing.T, app httpcaddyfile.App) {
				var cfg map[string]json.RawMessage
				if err := json.Unmarshal(app.Value, &cfg); err != nil {
					t.Fatalf("failed to unmarshal app config: %v", err)
				}
				if _, ok := cfg["parameters"]; ok {
					t.Error("expected parameters to be absent when not set")
				}
			},
		},
	} {
		t.Run("case", func(t *testing.T) {
			d := caddyfileadapter.NewTestDispenser(tc.input)
			val, err := parsePyroscopeOption(d, nil)
			if tc.wantErr {
				if err == nil {
					t.Errorf("test %d: expected error but got none", i)
				}
				return
			}
			if err != nil {
				t.Errorf("test %d: unexpected error: %v", i, err)
				return
			}
			app, ok := val.(httpcaddyfile.App)
			if !ok {
				t.Fatalf("test %d: expected httpcaddyfile.App, got %T", i, val)
			}
			if tc.validate != nil {
				tc.validate(t, app)
			}
		})
	}
}

func TestParseProfefeOption(t *testing.T) {
	input := `profefe {
		address http://profefe:4040
		service caddy-proxy
		timeout 30s
		profile_types cpu heap
	}`
	d := caddyfileadapter.NewTestDispenser(input)
	val, err := parseProfefeOption(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app, ok := val.(httpcaddyfile.App)
	if !ok {
		t.Fatalf("expected httpcaddyfile.App, got %T", val)
	}
	if app.Name != "profefe" {
		t.Errorf("expected app name 'profefe', got '%s'", app.Name)
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(app.Value, &cfg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	var addr string
	json.Unmarshal(cfg["address"], &addr)
	if addr != "http://profefe:4040" {
		t.Errorf("expected address 'http://profefe:4040', got '%s'", addr)
	}
}

func TestPyroscopeOptionProducesValidJSON(t *testing.T) {
	input := `pyroscope {
		server_address http://pyroscope:4040
		application_name caddy-proxy
		profile_types cpu heap allocs goroutine
	}`
	d := caddyfileadapter.NewTestDispenser(input)
	val, err := parsePyroscopeOption(d, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app, ok := val.(httpcaddyfile.App)
	if !ok {
		t.Fatalf("expected httpcaddyfile.App, got %T", val)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(app.Value, &raw); err != nil {
		t.Fatalf("produced invalid JSON: %v", err)
	}
	reencoded := caddyconfig.JSON(raw, nil)
	if len(reencoded) == 0 {
		t.Fatal("re-encoded JSON is empty")
	}
}
