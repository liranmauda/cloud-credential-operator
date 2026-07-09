package ibmcloud

import (
	"testing"
)

func Test_validateAPIEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "bare domain passes unchanged",
			input: "test.cloud.ibm.com",
			want:  "test.cloud.ibm.com",
		},
		{
			name:  "https scheme is stripped",
			input: "https://test.cloud.ibm.com",
			want:  "test.cloud.ibm.com",
		},
		{
			name:  "http scheme is stripped",
			input: "http://test.cloud.ibm.com",
			want:  "test.cloud.ibm.com",
		},
		{
			name:  "trailing slash is stripped",
			input: "test.cloud.ibm.com/",
			want:  "test.cloud.ibm.com",
		},
		{
			name:  "https scheme and trailing slash are both stripped",
			input: "https://test.cloud.ibm.com/",
			want:  "test.cloud.ibm.com",
		},
		{
			name:    "unsupported scheme returns error",
			input:   "ftp://test.cloud.ibm.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAPIEndpoint(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateAPIEndpoint(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("validateAPIEndpoint(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("validateAPIEndpoint(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func Test_NewClient_URLConstruction(t *testing.T) {
	const fakeKey = "fake-api-key-for-testing"

	tests := []struct {
		name                   string
		params                 *ClientParams
		wantIAMURL             string
		wantResourceManagerURL string
		wantErr                bool
	}{
		{
			name:                   "nil params uses SDK defaults",
			params:                 nil,
			wantIAMURL:             "https://iam.cloud.ibm.com",
			wantResourceManagerURL: "https://resource-controller.cloud.ibm.com",
		},
		{
			name:                   "empty APIEndpoint uses SDK defaults",
			params:                 &ClientParams{InfraName: "my-cluster"},
			wantIAMURL:             "https://iam.cloud.ibm.com",
			wantResourceManagerURL: "https://resource-controller.cloud.ibm.com",
		},
		{
			name:                   "bare domain builds correct URLs",
			params:                 &ClientParams{APIEndpoint: "test.cloud.ibm.com"},
			wantIAMURL:             "https://iam.test.cloud.ibm.com",
			wantResourceManagerURL: "https://resource-controller.test.cloud.ibm.com",
		},
		{
			name:                   "domain with https scheme builds correct URLs",
			params:                 &ClientParams{APIEndpoint: "https://test.cloud.ibm.com"},
			wantIAMURL:             "https://iam.test.cloud.ibm.com",
			wantResourceManagerURL: "https://resource-controller.test.cloud.ibm.com",
		},
		{
			name:    "unsupported scheme returns error",
			params:  &ClientParams{APIEndpoint: "ftp://test.cloud.ibm.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(fakeKey, tt.params)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			c := client.(*ibmcloudClient)
			if got := c.identityClient.GetServiceURL(); got != tt.wantIAMURL {
				t.Errorf("identityClient URL = %q, want %q", got, tt.wantIAMURL)
			}
			if got := c.pmClient.GetServiceURL(); got != tt.wantIAMURL {
				t.Errorf("pmClient URL = %q, want %q", got, tt.wantIAMURL)
			}
			if got := c.resourceManagerV2Client.GetServiceURL(); got != tt.wantResourceManagerURL {
				t.Errorf("resourceManagerV2Client URL = %q, want %q", got, tt.wantResourceManagerURL)
			}
		})
	}
}
