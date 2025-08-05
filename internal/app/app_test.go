package app

import (
	"net"
	"os"
	"pinshare/internal/config"
	"testing"
)

func Test_commandExists(t *testing.T) {
	type args struct {
		cmd string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test01",
			args: args{
				cmd: "ps",
			},
			want: true,
		}, {
			name: "test02",
			args: args{
				cmd: "fakecmd",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandExists(tt.args.cmd); got != tt.want {
				t.Errorf("commandExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkPort(t *testing.T) {
	type args struct {
		host string
		port int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test01",
			args: args{
				host: "8.8.8.8",
				port: 53,
			},
			want: true,
		}, {
			name: "test02",
			args: args{
				host: "127.0.0.1",
				port: 5001,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPort(tt.args.host, tt.args.port); got != tt.want {
				t.Errorf("checkPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkWebsite(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test01",
			args: args{
				url: "https://www.virustotal.com",
			},
			want: true,
		}, {
			name: "test02",
			args: args{
				url: "https://localhost",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkWebsite(tt.args.url); got != tt.want {
				t.Errorf("checkWebsite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkVTEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		// TODO: Add test cases.
		{
			name:  "test01",
			value: "REDACTED",
			want:  false,
		}, {
			name:  "test02",
			value: "",
			want:  false,
		}, {
			name:  "test03",
			value: "someotherlovelyvalue",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("VT_TOKEN", tt.value)
			if got := checkVTEnv(); got != tt.want {
				t.Errorf("checkVTEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkDependanciesAndEnableSecurityPath(t *testing.T) {
	type args struct {
		appconf     *config.AppConfig
		vttoken     string
		path        string
		nullProxy   bool
		dummyIPFS   bool
		dummyP2PSec bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test_sec=1",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "/opt/homebrew/bin:",
				nullProxy:   false,
				dummyIPFS:   true,
				dummyP2PSec: true,
			},
			want: true,
		},
		{
			name: "test_sec=1-f1",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "",
				nullProxy:   false,
				dummyIPFS:   true,
				dummyP2PSec: true,
			},
			want: false,
		},
		{
			name: "test_sec=1-f2",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "/opt/homebrew/bin:",
				nullProxy:   false,
				dummyIPFS:   false,
				dummyP2PSec: true,
			},
			want: false,
		},
		{
			name: "test_sec=2",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "testy123",
				path:        "/opt/homebrew/bin:",
				nullProxy:   false,
				dummyIPFS:   true,
				dummyP2PSec: false,
			},
			want: true,
		},
		{
			name: "test_sec=3",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "/opt/homebrew/bin:",
				nullProxy:   false,
				dummyIPFS:   true,
				dummyP2PSec: false,
			},
			want: true,
		},
		{
			name: "test_sec=4",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "/opt/test:", // ln -s /opt/homebrew/bin/ipfs
				nullProxy:   false,
				dummyIPFS:   true,
				dummyP2PSec: false,
			},
			want: true,
		},
		{
			name: "test_fail",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "",
				nullProxy:   true,
				dummyIPFS:   true,
				dummyP2PSec: false,
			},
			want: false,
		},
		{
			name: "test_sec=2-f1",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "testy123",
				path:        "/opt/homebrew/bin",
				nullProxy:   true,
				dummyIPFS:   false,
				dummyP2PSec: false,
			},
			want: false,
		},
		{
			name: "test_sec=0",
			args: args{
				appconf: &config.AppConfig{
					SecurityCapability: 0,
				},
				vttoken:     "",
				path:        "",
				nullProxy:   true,
				dummyIPFS:   false,
				dummyP2PSec: false,
			},
			want: false,
		},
	}
	oldtoken := os.Getenv("VT_TOKEN")
	oldpath := os.Getenv("PATH")
	for _, tt := range tests {
		os.Setenv("VT_TOKEN", tt.args.vttoken)
		os.Setenv("PATH", tt.args.path)
		if tt.args.nullProxy {
			os.Setenv("http_proxy", "null")
			os.Setenv("https_proxy", "null")
		} else {
			os.Unsetenv("http_proxy")
			os.Unsetenv("https_proxy")
		}
		var svr1_port = "12344"
		var svr2_port = "12345"
		if tt.args.dummyIPFS {
			svr1_port = "5001"
		}
		if tt.args.dummyP2PSec {
			svr2_port = "36939"
		}
		svr1, _ := net.Listen("tcp", ":"+svr1_port)
		defer svr1.Close()
		svr2, _ := net.Listen("tcp", ":"+svr2_port)
		defer svr2.Close()
		t.Run(tt.name, func(t *testing.T) {
			if got := checkDependanciesAndEnableSecurityPath(tt.args.appconf); got != tt.want {
				t.Errorf("checkDependanciesAndEnableSecurityPath() = %v, want %v", got, tt.want)
			}
		})
		svr1.Close()
		svr2.Close()
	}
	os.Setenv("VT_TOKEN", oldtoken)
	os.Setenv("PATH", oldpath)
	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
}
