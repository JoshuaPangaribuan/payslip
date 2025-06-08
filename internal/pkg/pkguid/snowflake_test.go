package pkguid

import (
	"reflect"
	"testing"

	sf "github.com/bwmarrin/snowflake"
)

func Test_snowflake_GenerateInt64(t *testing.T) {
	tests := []struct {
		name    string
		notWant int64
	}{
		{
			name:    "success",
			notWant: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := NewSnowflake()

			if got := s.GenerateInt64(); got == tt.notWant {
				t.Errorf("snowflake.GenerateInt64() = %v, not want %v", got, tt.notWant)
			}
		})
	}
}

func Test_snowflake_GenerateString(t *testing.T) {

	tests := []struct {
		name    string
		notWant string
	}{
		{
			name:    "success",
			notWant: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, _ := sf.NewNode(getNodeIDFromMachineIP())

			s := &snowflake{
				node: n,
			}
			if got := s.GenerateString(); got == tt.notWant {
				t.Errorf("snowflake.GenerateString() = %v, not want %v", got, tt.notWant)
			}
		})
	}
}

func Test_snowflake_GenerateSfID(t *testing.T) {
	// type fields struct {
	// 	node *sf.Node
	// }
	tests := []struct {
		name    string
		notWant ID
	}{
		{
			name:    "success",
			notWant: ID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, _ := sf.NewNode(getNodeIDFromMachineIP())

			s := &snowflake{
				node: n,
			}
			if got := s.GenerateSfID(); reflect.DeepEqual(got, tt.notWant) {
				t.Errorf("snowflake.GenerateSfID() = %v, not want %v", got, tt.notWant)
			}
		})
	}
}

func Test_snowflake_Generate(t *testing.T) {
	// type fields struct {
	// 	node *sf.Node
	// }
	tests := []struct {
		name    string
		notWant uint64
	}{
		{
			name:    "success",
			notWant: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, _ := sf.NewNode(getNodeIDFromMachineIP())

			s := &snowflake{
				node: n,
			}
			if got := s.Generate(); got == tt.notWant {
				t.Errorf("snowflake.Generate() = %v, not want %v", got, tt.notWant)
			}
		})
	}
}
