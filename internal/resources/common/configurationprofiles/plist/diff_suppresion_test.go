package plist

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessConfigurationProfileForDiffSuppression(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		fieldsToRemove []string
		want           string
		wantErr        bool
	}{
		{
			name: "Remove multiple standard fields",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadUUID</key>
			<string>12345678-1234-1234-1234-123456789012</string>
			<key>PayloadIdentifier</key>
			<string>com.example.profile</string>
			<key>PayloadOrganization</key>
			<string>Example Org</string>
			<key>PayloadDisplayName</key>
			<string>Test Profile</string>
			<key>PayloadDescription</key>
			<string>Test Description</string>
	</dict>
</plist>`,
			fieldsToRemove: []string{"PayloadUUID", "PayloadIdentifier", "PayloadOrganization", "PayloadDisplayName"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadDescription</key>
			<string>Test Description</string>
	</dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "Remove fields from nested array of dictionaries",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadUUID</key>
			<string>12345678-1234-1234-1234-123456789012</string>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>PayloadUUID</key>
							<string>abcdef12-3456-7890-abcd-ef1234567890</string>
							<key>PayloadIdentifier</key>
							<string>com.example.profile.1</string>
							<key>PayloadOrganization</key>
							<string>Example Org</string>
							<key>PayloadDisplayName</key>
							<string>Test Profile 1</string>
							<key>PayloadType</key>
							<string>com.apple.security.pkcs1</string>
					</dict>
					<dict>
							<key>PayloadUUID</key>
							<string>fedcba98-7654-3210-fedc-ba9876543210</string>
							<key>PayloadIdentifier</key>
							<string>com.example.profile.2</string>
							<key>PayloadOrganization</key>
							<string>Example Org</string>
							<key>PayloadDisplayName</key>
							<string>Test Profile 2</string>
							<key>PayloadType</key>
							<string>com.apple.security.pkcs1</string>
					</dict>
			</array>
	</dict>
</plist>`,
			fieldsToRemove: []string{"PayloadUUID", "PayloadIdentifier", "PayloadOrganization", "PayloadDisplayName"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>PayloadType</key>
							<string>com.apple.security.pkcs1</string>
					</dict>
					<dict>
							<key>PayloadType</key>
							<string>com.apple.security.pkcs1</string>
					</dict>
			</array>
	</dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "Remove fields from deeply nested structure",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadUUID</key>
			<string>12345678-1234-1234-1234-123456789012</string>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>PayloadUUID</key>
							<string>abcdef12-3456-7890-abcd-ef1234567890</string>
							<key>NestedContent</key>
							<dict>
									<key>PayloadUUID</key>
									<string>11111111-2222-3333-4444-555555555555</string>
									<key>PayloadIdentifier</key>
									<string>com.example.nested</string>
									<key>PayloadOrganization</key>
									<string>Example Org Nested</string>
									<key>PayloadDisplayName</key>
									<string>Nested Profile</string>
									<key>Settings</key>
									<dict>
											<key>PayloadUUID</key>
											<string>99999999-8888-7777-6666-555555555555</string>
											<key>ConfigurationOption</key>
											<string>SomeValue</string>
									</dict>
							</dict>
					</dict>
			</array>
	</dict>
</plist>`,
			fieldsToRemove: []string{"PayloadUUID", "PayloadIdentifier", "PayloadOrganization", "PayloadDisplayName"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>NestedContent</key>
							<dict>
									<key>Settings</key>
									<dict>
											<key>ConfigurationOption</key>
											<string>SomeValue</string>
									</dict>
							</dict>
					</dict>
			</array>
	</dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "Complex profile with mixed content and multiple nesting levels",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadUUID</key>
			<string>12345678-1234-1234-1234-123456789012</string>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>PayloadUUID</key>
							<string>abcdef12-3456-7890-abcd-ef1234567890</string>
							<key>PayloadIdentifier</key>
							<string>com.example.profile.1</string>
							<key>PayloadOrganization</key>
							<string>Example Org</string>
							<key>PayloadContent</key>
							<data>
									SGVsbG8gV29ybGQ=
							</data>
					</dict>
					<dict>
							<key>PayloadDisplayName</key>
							<string>Test Profile 2</string>
							<key>SubSettings</key>
							<array>
									<dict>
											<key>PayloadUUID</key>
											<string>11111111-2222-3333-4444-555555555555</string>
											<key>PayloadIdentifier</key>
											<string>com.example.subsetting</string>
											<key>Setting</key>
											<true/>
									</dict>
							</array>
					</dict>
			</array>
			<key>PayloadOrganization</key>
			<string>Example Org</string>
	</dict>
</plist>`,
			fieldsToRemove: []string{"PayloadUUID", "PayloadIdentifier", "PayloadOrganization", "PayloadDisplayName"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
			<key>PayloadContent</key>
			<array>
					<dict>
							<key>PayloadContent</key>
							<data>SGVsbG8gV29ybGQ=</data>
					</dict>
					<dict>
							<key>SubSettings</key>
							<array>
									<dict>
											<key>Setting</key>
											<true/>
									</dict>
							</array>
					</dict>
			</array>
	</dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "XML Tag Normalization Test",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Test1</key>
    <true />
    <key>Test2</key>
    <true  />
    <key>Test3</key>
    <true   />
    <key>Test4</key>
    <false />
    <key>Test5</key>
    <false    />
  </dict>
</plist>`,
			fieldsToRemove: []string{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Test1</key>
    <true/>
    <key>Test2</key>
    <true/>
    <key>Test3</key>
    <true/>
    <key>Test4</key>
    <false/>
    <key>Test5</key>
    <false/>
  </dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "Base64 Data Normalization",
			input: `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
		<dict>
			<key>Certificate1</key>
			<data>
						MIIFYjCCBEqgAwIBAgIQd70NbNs2
						+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF
			</data>
			<key>Certificate2</key>
			<data>
						MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF
			</data>
		</dict>
	</plist>`,
			fieldsToRemove: []string{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
		<dict>
			<key>Certificate1</key>
			<data>MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF</data>
			<key>Certificate2</key>
			<data>MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF</data>
		</dict>
	</plist>`,
			wantErr: false,
		},
		{
			name: "Integer Values Test",
			input: `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
		<dict>
			<key>PayloadContent</key>
			<array>
				<dict>
					<key>PayloadVersion</key>
					<integer>1</integer>
					<key>PayloadVersion2</key>
					<integer>1</integer>
				</dict>
			</array>
		</dict>
	</plist>`,
			fieldsToRemove: []string{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
		<dict>
			<key>PayloadContent</key>
			<array>
				<dict>
					<key>PayloadVersion</key>
					<integer>1</integer>
					<key>PayloadVersion2</key>
					<integer>1</integer>
				</dict>
			</array>
		</dict>
	</plist>`,
			wantErr: false,
		},
		{
			name: "Complex Root Certificate Profile",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>PayloadContent</key>
    <array>
      <dict>
        <key>PayloadDisplayName</key>
        <string>Google Example Root CA</string>
        <key>PayloadCertificateFileName</key>
        <string>GTS_Root_G1.cer</string>
        <key>PayloadContent</key>
        <data>
          MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsFADBX
          MQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQMA4GA1UE
          CxMHUm9vdCBDQTEbMBkGA1UEAxMSR2xvYmFsU2lnbiBSb290IENBMB4XDTIwMDYx
          OTAwMDA0MloXDTI4MDEyODAwMDA0MlowRzELMAkGA1UEBhMCVVMxIjAgBgNVBAoT
          GUdvb2dsZSBUcnVzdCBTZXJ2aWNlcyBMTEMxFDASBgNVBAMTC0dUUyBSb290IFIx
          MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAthECix7joXebO9y/lD63
          ladAPKH9gvl9MgaCcfb2jH/76Nu8ai6Xl6OMS/kr9rH5zoQdsfnFl97vufKj6bwS
          iV6nqlKr+CMny6SxnGPb15l+8Ape62im9MZaRw1NEDPjTrETo8gYbEvs/AmQ351k
          KSUjB6G00j0uYODP0gmHu81I8E3CwnqIiru6z1kZ1q+PsAewnjHxgsHA3y6mbWwZ
          DrXYfiYaRQM9sHmklCitD38m5agI/pboPGiUU+6DOogrFZYJsuB6jC511pzrp1Zk
          j5ZPaK49l8KEj8C8QMALXL32h7M1bKwYUH+E4EzNktMg6TO8UpmvMrUpsyUqtEj5
          cuHKZPfmghCN6J3Cioj6OGaK/GP5Afl4/Xtcd/p2h/rs37EOeZVXtL0m79YB0esW
          CruOC7XFxYpVq9Os6pFLKcwZpDIlTirxZUTQAs6qzkm06p98g7BAe+dDq6dso499
          iYH6TKX/1Y7DzkvgtdizjkXPdsDtQCv9Uw+wp9U7DbGKogPeMa3Md+pvez7W35Ei
          Eua++tgy/BBjFFFy3l3WFpO9KWgz7zpm7AeKJt8T11dleCfeXkkUAKIAf5qoIbap
          sZWwpbkNFhHax2xIPEDgfg1azVY80ZcFuctL7TlLnMQ/0lUTbiSw1nH69MG6zO0b
          9f6BQdgAmD06yK56mDcYBZUCAwEAAaOCATgwggE0MA4GA1UdDwEB/wQEAwIBhjAP
          BgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTkrysmcRorSCeFL1JmLO/wiRNxPjAf
          BgNVHSMEGDAWgBRge2YaRQ2XyolQL30EzTSo//z9SzBgBggrBgEFBQcBAQRUMFIw
          JQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnBraS5nb29nL2dzcjEwKQYIKwYBBQUH
          MAKGHWh0dHA6Ly9wa2kuZ29vZy9nc3IxL2dzcjEuY3J0MDIGA1UdHwQrMCkwJ6Al
          oCOGIWh0dHA6Ly9jcmwucGtpLmdvb2cvZ3NyMS9nc3IxLmNybDA7BgNVHSAENDAy
          MAgGBmeBDAECATAIBgZngQwBAgIwDQYLKwYBBAHWeQIFAwIwDQYLKwYBBAHWeQIF
          AwMwDQYJKoZIhvcNAQELBQADggEBADSkHrEoo9C0dhemMXoh6dFSPsjbdBZBiLg9
          NR3t5P+T4Vxfq7vqfM/b5A3Ri1fyJm9bvhdGaJQ3b2t6yMAYN/olUazsaL+yyEn9
          WprKASOshIArAoyZl+tJaox118fessmXn1hIVw41oeQa1v1vg4Fv74zPl6/AhSrw
          9U5pCZEt4Wi4wStz6dTZ/CLANx8LZh1J7QJVj2fhMtfTJr9w4z30Z209fOU0iOMy
          +qduBmpvvYuR7hZL6Dupszfnw0Skfths18dG9ZKb59UhvmaSGZRVbNQpsg3BZlvi
          d0lIKO2d1xozclOzgjXPYovJJIultzkMu34qQb9Sz/yilrbCgj8=
        </data>
        <key>PayloadDescription</key>
        <string></string>
        <key>AllowAllAppsAccess</key>
        <true />
        <key>KeyIsExtractable</key>
        <false />
        <key>PayloadEnabled</key>
        <true/>
        <key>PayloadIdentifier</key>
        <string>e0eda400-195d-4e65-9719-ab6ab33910cf</string>
        <key>PayloadOrganization</key>
        <string>Example Org</string>
        <key>PayloadType</key>
        <string>com.apple.security.pkcs1</string>
        <key>PayloadUUID</key>
        <string>e0eda400-195d-4e65-9719-ab6ab33910cf</string>
        <key>PayloadVersion</key>
        <integer>1</integer>
      </dict>
    </array>
    <key>PayloadDescription</key>
    <string>Distributes the Root Example Certificates</string>
    <key>PayloadDisplayName</key>
    <string>Example Certs</string>
    <key>PayloadEnabled</key>
    <true/>
    <key>PayloadIdentifier</key>
    <string>d0fde289-97c3-4d7c-a218-89a70f88c5aa</string>
    <key>PayloadOrganization</key>
    <string>Example Org</string>
    <key>PayloadRemovalDisallowed</key>
    <true/>
    <key>PayloadScope</key>
    <string>System</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>d0fde289-97c3-4d7c-a218-89a70f88c5aa</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
  </dict>
</plist>`,
			fieldsToRemove: []string{"PayloadUUID", "PayloadIdentifier", "PayloadOrganization"},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>PayloadContent</key>
    <array>
      <dict>
        <key>PayloadDisplayName</key>
        <string>Google Example Root CA</string>
        <key>PayloadCertificateFileName</key>
        <string>GTS_Root_G1.cer</string>
        <key>PayloadContent</key>
        <data>
          MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsFADBX
          MQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQMA4GA1UE
          CxMHUm9vdCBDQTEbMBkGA1UEAxMSR2xvYmFsU2lnbiBSb290IENBMB4XDTIwMDYx
          OTAwMDA0MloXDTI4MDEyODAwMDA0MlowRzELMAkGA1UEBhMCVVMxIjAgBgNVBAoT
          GUdvb2dsZSBUcnVzdCBTZXJ2aWNlcyBMTEMxFDASBgNVBAMTC0dUUyBSb290IFIx
          MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAthECix7joXebO9y/lD63
          ladAPKH9gvl9MgaCcfb2jH/76Nu8ai6Xl6OMS/kr9rH5zoQdsfnFl97vufKj6bwS
          iV6nqlKr+CMny6SxnGPb15l+8Ape62im9MZaRw1NEDPjTrETo8gYbEvs/AmQ351k
          KSUjB6G00j0uYODP0gmHu81I8E3CwnqIiru6z1kZ1q+PsAewnjHxgsHA3y6mbWwZ
          DrXYfiYaRQM9sHmklCitD38m5agI/pboPGiUU+6DOogrFZYJsuB6jC511pzrp1Zk
          j5ZPaK49l8KEj8C8QMALXL32h7M1bKwYUH+E4EzNktMg6TO8UpmvMrUpsyUqtEj5
          cuHKZPfmghCN6J3Cioj6OGaK/GP5Afl4/Xtcd/p2h/rs37EOeZVXtL0m79YB0esW
          CruOC7XFxYpVq9Os6pFLKcwZpDIlTirxZUTQAs6qzkm06p98g7BAe+dDq6dso499
          iYH6TKX/1Y7DzkvgtdizjkXPdsDtQCv9Uw+wp9U7DbGKogPeMa3Md+pvez7W35Ei
          Eua++tgy/BBjFFFy3l3WFpO9KWgz7zpm7AeKJt8T11dleCfeXkkUAKIAf5qoIbap
          sZWwpbkNFhHax2xIPEDgfg1azVY80ZcFuctL7TlLnMQ/0lUTbiSw1nH69MG6zO0b
          9f6BQdgAmD06yK56mDcYBZUCAwEAAaOCATgwggE0MA4GA1UdDwEB/wQEAwIBhjAP
          BgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTkrysmcRorSCeFL1JmLO/wiRNxPjAf
          BgNVHSMEGDAWgBRge2YaRQ2XyolQL30EzTSo//z9SzBgBggrBgEFBQcBAQRUMFIw
          JQYIKwYBBQUHMAGGGWh0dHA6Ly9vY3NwLnBraS5nb29nL2dzcjEwKQYIKwYBBQUH
          MAKGHWh0dHA6Ly9wa2kuZ29vZy9nc3IxL2dzcjEuY3J0MDIGA1UdHwQrMCkwJ6Al
          oCOGIWh0dHA6Ly9jcmwucGtpLmdvb2cvZ3NyMS9nc3IxLmNybDA7BgNVHSAENDAy
          MAgGBmeBDAECATAIBgZngQwBAgIwDQYLKwYBBAHWeQIFAwIwDQYLKwYBBAHWeQIF
          AwMwDQYJKoZIhvcNAQELBQADggEBADSkHrEoo9C0dhemMXoh6dFSPsjbdBZBiLg9
          NR3t5P+T4Vxfq7vqfM/b5A3Ri1fyJm9bvhdGaJQ3b2t6yMAYN/olUazsaL+yyEn9
          WprKASOshIArAoyZl+tJaox118fessmXn1hIVw41oeQa1v1vg4Fv74zPl6/AhSrw
          9U5pCZEt4Wi4wStz6dTZ/CLANx8LZh1J7QJVj2fhMtfTJr9w4z30Z209fOU0iOMy
          +qduBmpvvYuR7hZL6Dupszfnw0Skfths18dG9ZKb59UhvmaSGZRVbNQpsg3BZlvi
          d0lIKO2d1xozclOzgjXPYovJJIultzkMu34qQb9Sz/yilrbCgj8=
        </data>
        <key>PayloadDescription</key>
        <string></string>
        <key>AllowAllAppsAccess</key>
        <true />
        <key>KeyIsExtractable</key>
        <false />
        <key>PayloadEnabled</key>
        <true/>
        <key>PayloadType</key>
        <string>com.apple.security.pkcs1</string>
        <key>PayloadVersion</key>
        <integer>1</integer>
      </dict>
    </array>
    <key>PayloadDescription</key>
    <string>Distributes the Root Example Certificates</string>
    <key>PayloadDisplayName</key>
    <string>Example Certs</string>
    <key>PayloadEnabled</key>
    <true/>
    <key>PayloadRemovalDisallowed</key>
    <true/>
    <key>PayloadScope</key>
    <string>System</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
  </dict>
</plist>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessConfigurationProfileForDiffSuppression(tt.input, tt.fieldsToRemove)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Normalize both expected and actual output by removing all whitespace
			normalizedGot := normalizeWhitespace(got)
			normalizedWant := normalizeWhitespace(tt.want)

			if normalizedGot != normalizedWant {
				t.Errorf("ProcessConfigurationProfileForDiffSuppression() \nGot:  %v\nWant: %v", got, tt.want)
			}
		})
	}
}

// normalizeWhitespace removes all whitespace from a string to make comparison easier
func normalizeWhitespace(s string) string {
	// Replace all whitespace (including newlines and tabs) with a single space
	space := regexp.MustCompile(`\s+`)
	s = space.ReplaceAllString(s, " ")
	// Trim leading/trailing space
	return strings.TrimSpace(s)
}

func TestRemoveSpecifiedXMLFields(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]interface{}
		fieldsToRemove []string
		want           map[string]interface{}
	}{
		{
			name: "Remove single field",
			input: map[string]interface{}{
				"PayloadUUID": "test-uuid",
				"PayloadType": "test-type",
			},
			fieldsToRemove: []string{"PayloadUUID"},
			want: map[string]interface{}{
				"PayloadType": "test-type",
			},
		},
		{
			name: "Remove nested field",
			input: map[string]interface{}{
				"PayloadContent": map[string]interface{}{
					"PayloadUUID": "nested-uuid",
					"PayloadType": "nested-type",
				},
			},
			fieldsToRemove: []string{"PayloadUUID"},
			want: map[string]interface{}{
				"PayloadContent": map[string]interface{}{
					"PayloadType": "nested-type",
				},
			},
		},
		{
			name: "Remove field from array",
			input: map[string]interface{}{
				"PayloadContent": []interface{}{
					map[string]interface{}{
						"PayloadUUID": "array-uuid",
						"PayloadType": "array-type",
					},
				},
			},
			fieldsToRemove: []string{"PayloadUUID"},
			want: map[string]interface{}{
				"PayloadContent": []interface{}{
					map[string]interface{}{
						"PayloadType": "array-type",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeSpecifiedXMLFields(tt.input, tt.fieldsToRemove, "")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeBase64Content(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name: "Normalize deeply indented data tag content",
			input: `<data>
									MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF
									ADBXMQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQ
									MA4GA1UECxMHUm9vdCBDQTEbMBkGA1UEAxMSR2xvYmFsU2lnbiBSb290IENBMB4X
									</data>`,
			want: "<data>MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsFADBXMQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQMA4GA1UECxMHUm9vdCBDQTEbMBkGA1UEAxMSR2xvYmFsU2lnbiBSb290IENBMB4X</data>",
		},
		{
			name: "Handle nested map with deeply indented content",
			input: map[string]interface{}{
				"payload": "<data>\n\t\t\t\tMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA\n\t\t\t\t</data>",
				"other":   "regular string", // This should stay unchanged
			},
			want: map[string]interface{}{
				"payload": "<data>MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA</data>",
				"other":   "regular string",
			},
		},
		{
			name: "Handle array with mixed content",
			input: []interface{}{
				"<data>\n\t\t\t\tMIICIjAN\n\t\t\t\t</data>",
				"regular string", // This should stay unchanged
				map[string]interface{}{
					"nested": "<data> \n\t\tSGVsbG8= \n\t\t</data>",
				},
			},
			want: []interface{}{
				"<data>MIICIjAN</data>",
				"regular string",
				map[string]interface{}{
					"nested": "<data>SGVsbG8=</data>",
				},
			},
		},
		{
			name: "Handle multiple data tags in string",
			input: `<dict>
									<key>cert1</key>
									<data>
											MIICIjAN
											BgkqhkiG
									</data>
									<key>cert2</key>
									<data>
											SGVsbG8=
									</data>
							</dict>`,
			want: `<dict>
									<key>cert1</key>
									<data>MIICIjANBgkqhkiG</data>
									<key>cert2</key>
									<data>SGVsbG8=</data>
							</dict>`,
		},
		{
			name: "Non-base64 content",
			input: map[string]interface{}{
				"string": "Hello World", // This should stay unchanged
				"number": 42,
				"bool":   true,
			},
			want: map[string]interface{}{
				"string": "Hello World",
				"number": 42,
				"bool":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeBase64Content(tt.input)
			assert.Equal(t, tt.want, got, "normalizeBase64Content() = %v, want %v", got, tt.want)
		})
	}
}

func TestNormalizeBase64(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "Remove whitespace from base64",
			input: `MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsF
								 ADBXMQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQ`,
			want: "MIIFYjCCBEqgAwIBAgIQd70NbNs2+RrqIQ/E8FjTDTANBgkqhkiG9w0BAQsFADBXMQswCQYDVQQGEwJCRTEZMBcGA1UEChMQR2xvYmFsU2lnbiBudi1zYTEQ",
		},
		{
			name:  "Preserve XML content",
			input: "<dict><key>test</key><string>value</string></dict>",
			want:  "<dict><key>test</key><string>value</string></dict>",
		},
		{
			name:  "Deep indentation",
			input: "\n\t\t\t\tMIICIjAN\n\t\t\t\tBgkqhkiG9w0BAQ\n\t\t\t\t",
			want:  "MIICIjANBgkqhkiG9w0BAQ",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Only whitespace",
			input: "\n\t\r ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeBase64(tt.input)
			assert.Equal(t, tt.want, got, "NormalizeBase64() = %v, want %v", got, tt.want)
		})
	}
}

func TestNormalizeXMLTags(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "Normalize true tag",
			input: "<true    />",
			want:  "<true/>",
		},
		{
			name:  "Normalize false tag",
			input: "<false\t/>",
			want:  "<false/>",
		},
		{
			name:  "Handle normal string",
			input: "regular string",
			want:  "regular string",
		},
		{
			name: "Handle nested map",
			input: map[string]interface{}{
				"enabled": "<true   />",
			},
			want: map[string]interface{}{
				"enabled": "<true/>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeXMLTags(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnescapeHTMLEntities(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "Unescape basic entities",
			input: "Hello &amp; World",
			want:  "Hello & World",
		},
		{
			name:  "Handle no entities",
			input: "Regular string",
			want:  "Regular string",
		},
		{
			name: "Handle nested map",
			input: map[string]interface{}{
				"text": "Hello &amp; World",
			},
			want: map[string]interface{}{
				"text": "Hello & World",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeHTMLEntities(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimTrailingWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Trim spaces and tabs",
			input: "line1  \t \nline2\t  \nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "Handle no trailing whitespace",
			input: "line1\nline2\nline3",
			want:  "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimTrailingWhitespace(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmptyStringHandling(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "Empty string in PayloadDescription",
			input: `<key>PayloadDescription</key>
<string></string>`,
			want: `<key>PayloadDescription</key>
<string></string>`,
		},
		{
			name: "String with only whitespace",
			input: `<key>PayloadDescription</key>
<string>    </string>`,
			want: `<key>PayloadDescription</key>
<string>    </string>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeXMLTags(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Add new test for boolean value normalization
func TestBooleanValueNormalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Space after true",
			input: "<true />",
			want:  "<true/>",
		},
		{
			name:  "Multiple spaces after true",
			input: "<true     />",
			want:  "<true/>",
		},
		{
			name:  "Space after false",
			input: "<false />",
			want:  "<false/>",
		},
		{
			name:  "Tab after false",
			input: "<false\t/>",
			want:  "<false/>",
		},
		{
			name:  "Mixed whitespace",
			input: "<true \t  />",
			want:  "<true/>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeXMLTags(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNormalizeEmptyStrings(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  map[string]interface{}
	}{
		{
			name: "Empty and whitespace strings",
			input: map[string]interface{}{
				"empty":    "",
				"spaces":   "   ",
				"newlines": "\n    \n",
				"tabs":     "\t\t",
				"mixed":    "  \n\t  ",
				"content":  "actual content",
			},
			want: map[string]interface{}{
				"empty":    "",
				"spaces":   "",
				"newlines": "",
				"tabs":     "",
				"mixed":    "",
				"content":  "actual content",
			},
		},
		{
			name: "Nested dictionary",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner_empty":   "",
					"inner_spaces":  "   ",
					"inner_content": "test",
				},
			},
			want: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner_empty":   "",
					"inner_spaces":  "",
					"inner_content": "test",
				},
			},
		},
		{
			name: "Array with empty strings",
			input: map[string]interface{}{
				"items": []interface{}{
					"   ",
					"\n\t",
					"content",
					"",
				},
			},
			want: map[string]interface{}{
				"items": []interface{}{
					"",
					"",
					"content",
					"",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeEmptyStrings(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessConfigurationProfileEmptyStrings(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		fieldsToRemove []string
		want           string
		wantErr        bool
	}{
		{
			name: "Empty string variations",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>empty</key>
	<string></string>
	<key>spaces</key>
	<string>   </string>
	<key>newlines</key>
	<string>
	
	</string>
	<key>content</key>
	<string>actual content</string>
</dict>
</plist>`,
			fieldsToRemove: []string{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>content</key>
	<string>actual content</string>
	<key>empty</key>
	<string/>
	<key>newlines</key>
	<string/>
	<key>spaces</key>
	<string/>
</dict>
</plist>`,
			wantErr: false,
		},
		{
			name: "Mixed empty strings and content",
			input: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadDescription</key>
	<string>
	</string>
	<key>PayloadDisplayName</key>
	<string>Test Profile</string>
	<key>PayloadContent</key>
	<array>
			<dict>
					<key>EmptyField</key>
					<string>   </string>
					<key>ContentField</key>
					<string>actual content</string>
			</dict>
	</array>
</dict>
</plist>`,
			fieldsToRemove: []string{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
			<dict>
					<key>ContentField</key>
					<string>actual content</string>
					<key>EmptyField</key>
					<string/>
			</dict>
	</array>
	<key>PayloadDescription</key>
	<string/>
	<key>PayloadDisplayName</key>
	<string>Test Profile</string>
</dict>
</plist>`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessConfigurationProfileForDiffSuppression(tt.input, tt.fieldsToRemove)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Normalize both expected and actual output by removing all whitespace
			normalizedGot := normalizeWhitespace(got)
			normalizedWant := normalizeWhitespace(tt.want)

			assert.Equal(t, normalizedWant, normalizedGot)

			if normalizedGot != normalizedWant {
				t.Errorf("ProcessConfigurationProfileForDiffSuppression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
