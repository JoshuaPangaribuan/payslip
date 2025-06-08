package pkguid

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUUID(t *testing.T) {
	t.Run("should create new UUID generator", func(t *testing.T) {
		// when
		gen := NewUUID()

		// then
		assert.NotNil(t, gen)
		assert.Implements(t, (*UUID)(nil), gen)
	})
}

func Test_uuidGen_Generate(t *testing.T) {
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
			u := NewUUID()
			got := u.Generate()
			if got == tt.notWant {
				t.Errorf("uuidGen.Generate() = %v, not want %v", got, tt.notWant)
			}
			// Verify it's a valid UUID
			if _, err := uuid.Parse(got); err != nil {
				t.Errorf("uuidGen.Generate() = %v, is not a valid UUID: %v", got, err)
			}
		})
	}
}

func Test_uuidGen_GenerateV4(t *testing.T) {
	tests := []struct {
		name    string
		notWant uuid.UUID
	}{
		{
			name:    "success",
			notWant: uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUUID()
			got := u.GenerateV4()
			if got == tt.notWant {
				t.Errorf("uuidGen.GenerateV4() = %v, not want %v", got, tt.notWant)
			}
			// Verify it's a valid UUID v4
			if got.Version() != 4 {
				t.Errorf("uuidGen.GenerateV4() = %v, is not a valid UUID v4", got)
			}
		})
	}
}

func Test_uuidGen_Generate_Different(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUUID()
			got1 := u.Generate()
			got2 := u.Generate()
			if got1 == got2 {
				t.Errorf("uuidGen.Generate() generated same UUID twice: %v", got1)
			}
		})
	}
}

func Test_uuidGen_GenerateV4_Different(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewUUID()
			got1 := u.GenerateV4()
			got2 := u.GenerateV4()
			if got1 == got2 {
				t.Errorf("uuidGen.GenerateV4() generated same UUID twice: %v", got1)
			}
		})
	}
}
