package pyroscope

import (
	"testing"

	"github.com/grafana/pyroscope-go"
	"go.lumeweb.com/caddy_profiling/types"
)

func TestApp_SetProfilingParameter(t *testing.T) {
	type fields struct {
		Parameters   *types.Parameters
		profileTypes []pyroscope.ProfileType
	}
	type args struct {
		parameters types.Parameters
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "",
			fields: fields{
				Parameters: &types.Parameters{
					ProfileTypes: []types.ProfileType{"goroutine", "allocs", "block", "mutex"},
				},
				profileTypes: []pyroscope.ProfileType{},
			},
			args: args{
				parameters: types.Parameters{
					ProfileTypes: []types.ProfileType{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				Parameters:   tt.fields.Parameters,
				profileTypes: tt.fields.profileTypes,
			}
			a.SetProfilingParameter(tt.args.parameters)
			t.Logf("%+v", a.profileTypes)
		})
	}
}
