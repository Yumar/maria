package config

import (
	"log/slog"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults",
			env:  map[string]string{},
			want: Config{AppEnv: EnvDevelopment, HTTPAddr: ":8080", LogLevel: slog.LevelInfo, LogFormat: LogFormatJSON},
		},
		{
			name: "overrides",
			env: map[string]string{
				"APP_ENV":    "production",
				"HTTP_ADDR":  ":9090",
				"LOG_LEVEL":  "debug",
				"LOG_FORMAT": "text",
			},
			want: Config{AppEnv: EnvProduction, HTTPAddr: ":9090", LogLevel: slog.LevelDebug, LogFormat: LogFormatText},
		},
		{name: "invalid APP_ENV", env: map[string]string{"APP_ENV": "staging"}, wantErr: true},
		{name: "invalid LOG_LEVEL", env: map[string]string{"LOG_LEVEL": "verbose"}, wantErr: true},
		{name: "invalid LOG_FORMAT", env: map[string]string{"LOG_FORMAT": "xml"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"APP_ENV", "HTTP_ADDR", "LOG_LEVEL", "LOG_FORMAT"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
