package keygen

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseValidateLicenseKeyResponseBody(t *testing.T) {
	tests := []struct {
		name            string
		responseBody    string
		expectedLicense *LicenseID
		expectedError   error
	}{
		{
			name: "license pending activation",
			responseBody: `{
    "data": {
        "id": "0e63accc-2903-4e74-b703-eb8a47b9d6c8",
        "type": "licenses",
        "attributes": {
            "name": null,
            "key": "3EE66B-626606-AE7999-C023F0-767194-V3",
            "expiry": null,
            "status": "ACTIVE",
            "uses": 0,
            "suspended": false,
            "scheme": null,
            "encrypted": false,
            "strict": true,
            "floating": false,
            "protected": true,
            "version": null,
            "maxMachines": 1,
            "maxProcesses": null,
            "maxUsers": null,
            "maxCores": null,
            "maxUses": null,
            "requireHeartbeat": false,
            "requireCheckIn": false,
            "lastValidated": "2025-05-06T04:09:54.529Z",
            "lastCheckIn": null,
            "nextCheckIn": null,
            "lastCheckOut": null,
            "metadata": {},
            "created": "2025-05-06T04:09:45.128Z",
            "updated": "2025-05-06T04:09:45.128Z"
        },
        "relationships": {
            "account": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151"
                },
                "data": {
                    "type": "accounts",
                    "id": "87c9078c-6ce3-4c29-9fa2-e092a594b151"
                }
            },
            "product": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/product"
                },
                "data": {
                    "type": "products",
                    "id": "c0fbaab6-034e-48f3-8a78-70bb8c59e2cf"
                }
            },
            "policy": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/policy"
                },
                "data": {
                    "type": "policies",
                    "id": "4986ef97-bf44-4d90-8bc0-b47c2439d00f"
                }
            },
            "group": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/group"
                },
                "data": null
            },
            "owner": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/owner"
                },
                "data": null
            },
            "users": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/users"
                },
                "meta": {
                    "count": 0
                }
            },
            "machines": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/machines"
                },
                "meta": {
                    "cores": 0,
                    "count": 0
                }
            },
            "tokens": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/tokens"
                }
            },
            "entitlements": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8/entitlements"
                }
            }
        },
        "links": {
            "self": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/0e63accc-2903-4e74-b703-eb8a47b9d6c8"
        }
    },
    "meta": {
        "ts": "2025-05-06T04:09:54.531Z",
        "valid": false,
        "detail": "fingerprint is not activated (has no associated machine)",
        "code": "NO_MACHINE",
        "scope": {
            "fingerprint": "Y0AXq9TstLyDQcGPhhYlhesF_Mg1Hj4FF7fErY69is8X"
        }
    }
}`,
			expectedLicense: &LicenseID{
				ID:          "0e63accc-2903-4e74-b703-eb8a47b9d6c8",
				ExpireAt:    nil,
				IsActivated: false,
				IsExpired:   false,
			},
			expectedError: nil,
		},

		{
			name: "valid license",
			responseBody: `{
    "data": {
        "id": "9d1e8df9-229f-4b5d-a207-945dcfa1e996",
        "type": "licenses",
        "attributes": {
            "name": null,
            "key": "8ECE46-C5CB99-263245-93E5CC-AD0361-V3",
            "expiry": "2025-05-29T07:02:09.922Z",
            "status": "ACTIVE",
            "uses": 0,
            "suspended": false,
            "scheme": null,
            "encrypted": false,
            "strict": true,
            "floating": false,
            "protected": true,
            "version": null,
            "maxMachines": 1,
            "maxProcesses": null,
            "maxUsers": null,
            "maxCores": null,
            "maxUses": null,
            "requireHeartbeat": false,
            "requireCheckIn": false,
            "lastValidated": "2025-05-06T03:38:47.579Z",
            "lastCheckIn": null,
            "nextCheckIn": null,
            "lastCheckOut": null,
            "metadata": {
                "stripeCustomerId": "cus_SC8kvfXLrODZlq",
                "stripeCheckoutSessionId": "cs_test_a12FEQu82usfxGayomGKYubHZTA6NwjnFwdgTg1rIYKNdKh421wEQGhVXn"
            },
            "created": "2025-04-25T11:30:23.076Z",
            "updated": "2025-05-06T03:38:10.304Z"
        },
        "relationships": {
            "account": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151"
                },
                "data": {
                    "type": "accounts",
                    "id": "87c9078c-6ce3-4c29-9fa2-e092a594b151"
                }
            },
            "product": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/product"
                },
                "data": {
                    "type": "products",
                    "id": "c0fbaab6-034e-48f3-8a78-70bb8c59e2cf"
                }
            },
            "policy": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/policy"
                },
                "data": {
                    "type": "policies",
                    "id": "4986ef97-bf44-4d90-8bc0-b47c2439d00f"
                }
            },
            "group": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/group"
                },
                "data": null
            },
            "owner": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/owner"
                },
                "data": null
            },
            "users": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/users"
                },
                "meta": {
                    "count": 0
                }
            },
            "machines": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/machines"
                },
                "meta": {
                    "cores": 0,
                    "count": 1
                }
            },
            "tokens": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/tokens"
                }
            },
            "entitlements": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/entitlements"
                }
            }
        },
        "links": {
            "self": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996"
        }
    },
    "meta": {
        "ts": "2025-05-06T03:38:47.606Z",
        "valid": true,
        "detail": "is valid",
        "code": "VALID",
        "scope": {
            "fingerprint": "Y0AXq9TstLyDQcGPhhYlhesF_Mg1Hj4FF7fErY69is8X"
        }
    }
}`,
			expectedLicense: &LicenseID{
				ID:          "9d1e8df9-229f-4b5d-a207-945dcfa1e996",
				ExpireAt:    timeDate(2025, 5, 29, 7, 2, 9, int(922*time.Millisecond), time.UTC),
				IsExpired:   false,
				IsActivated: true,
			},
			expectedError: nil,
		},

		{
			name: "expired license",
			responseBody: `{
    "data": {
        "id": "9d1e8df9-229f-4b5d-a207-945dcfa1e996",
        "type": "licenses",
        "attributes": {
            "name": null,
            "key": "8ECE46-C5CB99-263245-93E5CC-AD0361-V3",
            "expiry": "2025-04-29T07:02:09.922Z",
            "status": "EXPIRED",
            "uses": 0,
            "suspended": false,
            "scheme": null,
            "encrypted": false,
            "strict": true,
            "floating": false,
            "protected": true,
            "version": null,
            "maxMachines": 1,
            "maxProcesses": null,
            "maxUsers": null,
            "maxCores": null,
            "maxUses": null,
            "requireHeartbeat": false,
            "requireCheckIn": false,
            "lastValidated": "2025-05-06T03:44:37.727Z",
            "lastCheckIn": null,
            "nextCheckIn": null,
            "lastCheckOut": null,
            "metadata": {
                "stripeCustomerId": "cus_SC8kvfXLrODZlq",
                "stripeCheckoutSessionId": "cs_test_a12FEQu82usfxGayomGKYubHZTA6NwjnFwdgTg1rIYKNdKh421wEQGhVXn"
            },
            "created": "2025-04-25T11:30:23.076Z",
            "updated": "2025-05-06T03:44:28.039Z"
        },
        "relationships": {
            "account": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151"
                },
                "data": {
                    "type": "accounts",
                    "id": "87c9078c-6ce3-4c29-9fa2-e092a594b151"
                }
            },
            "product": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/product"
                },
                "data": {
                    "type": "products",
                    "id": "c0fbaab6-034e-48f3-8a78-70bb8c59e2cf"
                }
            },
            "policy": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/policy"
                },
                "data": {
                    "type": "policies",
                    "id": "4986ef97-bf44-4d90-8bc0-b47c2439d00f"
                }
            },
            "group": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/group"
                },
                "data": null
            },
            "owner": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/owner"
                },
                "data": null
            },
            "users": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/users"
                },
                "meta": {
                    "count": 0
                }
            },
            "machines": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/machines"
                },
                "meta": {
                    "cores": 0,
                    "count": 1
                }
            },
            "tokens": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/tokens"
                }
            },
            "entitlements": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/entitlements"
                }
            }
        },
        "links": {
            "self": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996"
        }
    },
    "meta": {
        "ts": "2025-05-06T03:44:37.740Z",
        "valid": true,
        "detail": "is expired",
        "code": "EXPIRED",
        "scope": {
            "fingerprint": "Y0AXq9TstLyDQcGPhhYlhesF_Mg1Hj4FF7fErY69is8X"
        }
    }
}`,
			expectedLicense: &LicenseID{
				ID:          "9d1e8df9-229f-4b5d-a207-945dcfa1e996",
				ExpireAt:    timeDate(2025, 4, 29, 7, 2, 9, int(922*time.Millisecond), time.UTC),
				IsActivated: true,
				IsExpired:   true,
			},
			expectedError: nil,
		},

		{
			name: "license already activated",
			responseBody: `{
    "data": {
        "id": "9d1e8df9-229f-4b5d-a207-945dcfa1e996",
        "type": "licenses",
        "attributes": {
            "name": null,
            "key": "8ECE46-C5CB99-263245-93E5CC-AD0361-V3",
            "expiry": "2025-04-29T07:02:09.922Z",
            "status": "EXPIRED",
            "uses": 0,
            "suspended": false,
            "scheme": null,
            "encrypted": false,
            "strict": true,
            "floating": false,
            "protected": true,
            "version": null,
            "maxMachines": 1,
            "maxProcesses": null,
            "maxUsers": null,
            "maxCores": null,
            "maxUses": null,
            "requireHeartbeat": false,
            "requireCheckIn": false,
            "lastValidated": "2025-05-06T04:14:43.290Z",
            "lastCheckIn": null,
            "nextCheckIn": null,
            "lastCheckOut": null,
            "metadata": {
                "stripeCustomerId": "cus_SC8kvfXLrODZlq",
                "stripeCheckoutSessionId": "cs_test_a12FEQu82usfxGayomGKYubHZTA6NwjnFwdgTg1rIYKNdKh421wEQGhVXn"
            },
            "created": "2025-04-25T11:30:23.076Z",
            "updated": "2025-05-06T04:11:28.074Z"
        },
        "relationships": {
            "account": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151"
                },
                "data": {
                    "type": "accounts",
                    "id": "87c9078c-6ce3-4c29-9fa2-e092a594b151"
                }
            },
            "product": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/product"
                },
                "data": {
                    "type": "products",
                    "id": "c0fbaab6-034e-48f3-8a78-70bb8c59e2cf"
                }
            },
            "policy": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/policy"
                },
                "data": {
                    "type": "policies",
                    "id": "4986ef97-bf44-4d90-8bc0-b47c2439d00f"
                }
            },
            "group": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/group"
                },
                "data": null
            },
            "owner": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/owner"
                },
                "data": null
            },
            "users": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/users"
                },
                "meta": {
                    "count": 0
                }
            },
            "machines": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/machines"
                },
                "meta": {
                    "cores": 0,
                    "count": 1
                }
            },
            "tokens": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/tokens"
                }
            },
            "entitlements": {
                "links": {
                    "related": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996/entitlements"
                }
            }
        },
        "links": {
            "self": "/v1/accounts/87c9078c-6ce3-4c29-9fa2-e092a594b151/licenses/9d1e8df9-229f-4b5d-a207-945dcfa1e996"
        }
    },
    "meta": {
        "ts": "2025-05-06T04:14:43.299Z",
        "valid": false,
        "detail": "fingerprint is not activated (does not match any associated machines)",
        "code": "FINGERPRINT_SCOPE_MISMATCH",
        "scope": {
            "fingerprint": "Y0AXq9TstLyDQcGPhhYlhesF_Mg1Hj4FF7fErY69is8"
        }
    }
}
			`,
			expectedLicense: nil,
			expectedError:   ErrLicenseKeyAlreadyActivated,
		},

		{
			name: "license not found",
			responseBody: `{
    "data": null,
    "meta": {
        "ts": "2025-05-06T03:43:27.496Z",
        "valid": false,
        "detail": "does not exist",
        "code": "NOT_FOUND",
        "scope": {
            "fingerprint": "Y0AXq9TstLyDQcGPhhYlhesF_Mg1Hj4FF7fErY69is8X"
        }
    }
}`,
			expectedLicense: nil,
			expectedError:   ErrLicenseKeyNotFound,
		},

		{
			name:            "invalid json",
			responseBody:    `invalid json`,
			expectedLicense: nil,
			expectedError:   errors.New("invalid character 'i' looking for beginning of value"),
		},

		{
			name: "valid license without data",
			responseBody: `{
				"meta": {
					"valid": true,
					"code": "VALID"
				}
			}`,
			expectedLicense: nil,
			expectedError:   ErrUnexpectedResponse,
		},
		{
			name: "valid license with null data",
			responseBody: `{
				"meta": {
					"valid": true,
					"code": "VALID"
				},
				"data": null
			}`,
			expectedLicense: nil,
			expectedError:   ErrUnexpectedResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the JSON string to a reader
			reader := strings.NewReader(tt.responseBody)

			// Call the function being tested
			licenseID, err := parseValidateLicenseKeyResponseBody(reader)

			// Check the error
			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) && err.Error() != tt.expectedError.Error() {
					// Special case for JSON decoding errors which don't support errors.Is
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !reflect.DeepEqual(licenseID, tt.expectedLicense) {
				t.Errorf("expected %v == %v", licenseID, tt.expectedLicense)
			}
		})
	}
}

func timeDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) *time.Time {
	t := time.Date(year, month, day, hour, min, sec, nsec, loc)
	return &t
}
