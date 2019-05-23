package nele

import (
	"testing"
)

func TestNewPageHandler(t *testing.T) {
	SetPostingBaseDirectory("/tmp/postings/")
	prepareTestFiles()
	tests := []struct {
		name    string
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", 15, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPageHandler()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPageHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != got.Len() {
				t.Errorf("NewPageHandler() = %v, want %v", got.Len(), tt.want)
			}
		})
	}
} // TestNewPageHandler()
