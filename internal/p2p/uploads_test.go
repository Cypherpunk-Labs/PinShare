package p2p

import (
	"pinshare/internal/config"
	"testing"
)

func TestProcessUploads(t *testing.T) {
	type args struct {
		folderPath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// {
		// 	name: "test01",
		// 	args: args{folderPath: "../../test/test01"}, // don't use this too many files test timesout
		// 	want: false,
		// },
		{
			name: "test02",
			args: args{folderPath: "../../test/test02"},
			want: true,
		},
		{
			name: "test03",
			args: args{folderPath: "../../test/test03"},
			want: true,
		},
	}
	for _, tt := range tests {
		conf, _ := config.LoadConfig()
		conf.SecurityCapability = 3
		SetAppConfig(conf)
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessUploads(tt.args.folderPath); got != tt.want {
				t.Errorf("ProcessUploads() = %v, want %v", got, tt.want)
			}
		})
	}
}
