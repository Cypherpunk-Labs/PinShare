package psfs

import (
	"testing"
)

func Test_getVirusTotalReportByHash(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "test01-ok",
			args:    args{hash: "7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "test02-bad",
			args:    args{hash: "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "test03-noreport",
			args:    args{hash: "7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9068"},
			want:    false,
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVirusTotalVerdictByHash(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVirusTotalReportByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getVirusTotalReportByHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateCID(t *testing.T) {
	type args struct {
		cidString string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "test01",
			args:    args{cidString: "QmV7FsdBKuLvg41YMY6ee3yqjTfRxh92oVZ2K2eRHDB9cu"},
			want:    true,
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateCID(tt.args.cidString)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateCID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCID(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test01",
			args:    args{filePath: "/Users/mkemp/repos/cypherpunk-labs/test-data-ai-etl/docs/pdf/Nagel IE163 Direct Electrical Production from LENR.pdf"},
			want:    "QmV7FsdBKuLvg41YMY6ee3yqjTfRxh92oVZ2K2eRHDB9cu",
			wantErr: false,
		},
		{
			name:    "test02",
			args:    args{filePath: "/Users/mkemp/repos/cypherpunk-labs/test-data-ai-etl/text/HelloWorld.txt"},
			want:    "bafkreid7qoywk77r7rj3slobqfekdvs57qwuwh5d2z3sqsw52iabe3mqne",
			wantErr: false,
		},
		// ipfs add
		// sha256: 979562e2c31b9ea28d30f6b87dbcc950f47a02c175d1fcbe6d2ce209528837ec
		// ipfs add --cid-version=1 --raw-leaves Nagel\ IE163\ Direct\ Electrical\ Production\ from\ LENR.pdf
		// CID: bafybeiaecc4byj5gpadpfjnosh6y5zzqsolkgj5vdu26cqwnntodohmza4
		// in code
		// Mhash: 12203417fc7ba569b15bd2ccf67de4382f46ef4c5514e245b996a411844c974c37c4
		// CID: bafybeibuc76hxjljwfn5fthwpxsdql2g55gfkfhciw4znjarqrgjotbxyq
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCID(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sendFileToVirusTotal(t *testing.T) {
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
			args:    args{filepath: "/Users/mkemp/repos/cypherpunk-labs/test-data-ai-etl/text/HelloWorld.txt"},
			want:    true,
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SendFileToVirusTotal(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendFileToVirusTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sendFileToVirusTotal() = %v, want %v", got, tt.want)
			}
		})
	}
}
