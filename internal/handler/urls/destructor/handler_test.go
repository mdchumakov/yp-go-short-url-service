package destructor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestructorRequestBodyDTOIn_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dto     DestructorRequestBodyDTOIn
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid URLs",
			dto:     DestructorRequestBodyDTOIn{"6qxTVvsy", "RTfd56hn", "Jlfd67ds"},
			wantErr: false,
		},
		{
			name:    "empty array",
			dto:     DestructorRequestBodyDTOIn{},
			wantErr: true,
			errMsg:  "urls array cannot be empty",
		},
		{
			name:    "empty URL",
			dto:     DestructorRequestBodyDTOIn{"6qxTVvsy", "", "Jlfd67ds"},
			wantErr: true,
			errMsg:  "short URL cannot be empty",
		},
		{
			name:    "URL too short",
			dto:     DestructorRequestBodyDTOIn{"6qxTVvsy", "ab", "Jlfd67ds"},
			wantErr: true,
			errMsg:  "short URL is too short",
		},
		{
			name:    "URL too long",
			dto:     DestructorRequestBodyDTOIn{"6qxTVvsy", "verylongurlthatiswaytoolongandshouldfailvalidationandmore", "Jlfd67ds"},
			wantErr: true,
			errMsg:  "short URL is too long",
		},
		{
			name:    "invalid characters",
			dto:     DestructorRequestBodyDTOIn{"6qxTVvsy", "RTfd56@n", "Jlfd67ds"},
			wantErr: true,
			errMsg:  "short URL contains invalid characters",
		},
		{
			name:    "URLs with spaces",
			dto:     DestructorRequestBodyDTOIn{" 6qxTVvsy ", " RTfd56hn ", " Jlfd67ds "},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
