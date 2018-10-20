package scripts

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSaveFileToPath(t *testing.T) {
	type args struct {
		filepath string
		data     []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Write to file",
			args:    args{filepath: "../tests/testfile", data: []byte("abcdef")},
			wantErr: false,
		},
		{
			name:    "Override file",
			args:    args{filepath: "../tests/testfile", data: []byte("ijklmn")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveFileToPath(tt.args.filepath, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SaveFileToPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			dat, err := ioutil.ReadFile(tt.args.filepath)
			if err != nil {
				t.Errorf("ReadFile() error")
			}

			if string(dat) != string(tt.args.data) {
				t.Errorf("Input and file content mismatch, want %s, got %s", string(tt.args.data), string(dat))
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Ensure directory exist",
			args:    args{filepath: "."},
			wantErr: true,
		},
		{
			name:    "Ensure directory exist",
			args:    args{filepath: "../tests/createdDir/"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureDir(tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err := os.Remove("../tests/createdDir/")
	if err != nil {
		t.Errorf("Could not remove the test directory (../tests/createdDir/)")
	}
}
