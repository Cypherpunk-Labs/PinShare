package localfs

import "testing"

func Test_chunker_test(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		// TODO: Add test cases.
		{
			name: "test01",
			file: "../../test/test02/gt256kb.txt",
		},
		{
			name: "test02",
			file: "../../test/test02/small.txt",
		},
		{
			name: "test03",
			file: "../../../../../cypherpunk-labs/test-data-ai-etl/docs/pdf/Nagel IE163 Direct Electrical Production from LENR.pdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if addFileIPFS(tt.file) != Cid(tt.file) {
				// if addFileIPFS(tt.file) != chunker_test(tt.file) {
				t.Fail()
			}
		})
	}
}
