package psfs

import "testing"

func Test_validateFileType(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "test01",
			args:    args{filepath: "../../test/test01/test01.pdf"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "test02",
			args:    args{filepath: "../../test/test01/test02.exe"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test03",
			args:    args{filepath: "../../test/test01/test03.rtf"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test04",
			args:    args{filepath: "../../test/test01/test04.md"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test05",
			args:    args{filepath: "../../test/test01/test05.json"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test06",
			args:    args{filepath: "../../test/test01/test06.php"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test07",
			args:    args{filepath: "../../test/test01/test07.xml"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test08",
			args:    args{filepath: "../../test/test01/test08.txt"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test09",
			args:    args{filepath: "../../test/test01/test09.html"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test10",
			args:    args{filepath: "../../test/test01/test10_v3.epub"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test11",
			args:    args{filepath: "../../test/test01/test11.csv"},
			want:    false, // true, // not on AllowedList
			wantErr: false,
		},
		{
			name:    "test91",
			args:    args{filepath: "../../test/test01/test91.pdf"},
			want:    false,
			wantErr: false,
		},

		// Biggest concern was the plaintext type files, but the validation tool seems to be better than expected.
		// .zip .arc
		// .jpg .png
		// .avi .mp4 .mov
		// .mp3
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateFileType(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFileType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateFileType() = %v, want %v", got, tt.want)
			}
		})
	}
}
