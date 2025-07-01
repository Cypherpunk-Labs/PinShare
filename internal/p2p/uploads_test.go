package p2p

import "testing"

func TestProcessUploads(t *testing.T) {
	type args struct {
		folderPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test01",
			args: args{folderPath: "../../test/test01"},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ProcessUploads(tt.args.folderPath)
		})
	}
}
