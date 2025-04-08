package main

import (
	"testing"

	"github.com/edgelaboratories/eve/pkg/asset"
	"github.com/edgelaboratories/eve/pkg/marketdata"
	"github.com/edgelaboratories/go-libraries/date"
	"github.com/stretchr/testify/require"
)

func Test_inputToAsset(t *testing.T) {
	for name, tc := range map[string]struct {
		input    CDSInput
		expected []CDSAsset
	}{
		"single input": {
			input: CDSInput{
				ID: "issuer",
				UpfrontPayments: map[Tenor]float64{
					"M12": 0.2,
				},
				InterestCurve: marketdata.TermStructure{
					"M12": 0.01,
				},
				RecoveryRate: 0.4,
				CouponRate:   0.01,
				Date:         date.New(2024, 9, 10),
			},
			expected: []CDSAsset{
				{
					ID:       "issuer",
					Maturity: date.New(2025, 9, 20),
					Coupons: []asset.FixedCoupon{
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 6, 20),
							PaymentDate:   date.New(2024, 9, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 9, 20),
							PaymentDate:   date.New(2024, 12, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 12, 20),
							PaymentDate:   date.New(2025, 3, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 3, 20),
							PaymentDate:   date.New(2025, 6, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 6, 20),
							PaymentDate:   date.New(2025, 9, 20),
						},
					},
					Date:         date.New(2024, 9, 10),
					RecoveryRate: 0.4,
					Upfront:      0.002,
					InterestCurve: &InterestRateCurveRepresentation{
						Data: marketdata.TermStructure{
							"M12": 0.01,
						},
					},
				},
			},
		},
		"single input/Y2": {
			input: CDSInput{
				ID: "issuer",
				UpfrontPayments: map[Tenor]float64{
					"Y2": 0.5,
				},
				InterestCurve: marketdata.TermStructure{
					"M12": 0.01,
				},
				RecoveryRate: 0.4,
				CouponRate:   0.01,
				Date:         date.New(2024, 9, 10),
			},
			expected: []CDSAsset{
				{
					ID:       "issuer",
					Maturity: date.New(2026, 9, 20),
					Coupons: []asset.FixedCoupon{
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 6, 20),
							PaymentDate:   date.New(2024, 9, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 9, 20),
							PaymentDate:   date.New(2024, 12, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2024, 12, 20),
							PaymentDate:   date.New(2025, 3, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 3, 20),
							PaymentDate:   date.New(2025, 6, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 6, 20),
							PaymentDate:   date.New(2025, 9, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 9, 20),
							PaymentDate:   date.New(2025, 12, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2025, 12, 20),
							PaymentDate:   date.New(2026, 3, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2026, 3, 20),
							PaymentDate:   date.New(2026, 6, 20),
						},
						{
							Type:          asset.FixedType,
							FixedRate:     0.01,
							InitialFixing: date.New(2026, 6, 20),
							PaymentDate:   date.New(2026, 9, 20),
						},
					},
					Date:         date.New(2024, 9, 10),
					RecoveryRate: 0.4,
					Upfront:      0.005,
					InterestCurve: &InterestRateCurveRepresentation{
						Data: marketdata.TermStructure{
							"M12": 0.01,
						},
					},
				},
			},
		},
	} {
		tc := tc

		t.Run(name, func(t *testing.T) {
			got, err := inputToAsset(tc.input)
			require.NoError(t, err)

			for i := range tc.expected {
				require.Equal(t, tc.expected[i].ID, got[i].ID)
				require.Equal(t, tc.expected[i].Maturity, got[i].Maturity)
				require.Equal(t, tc.expected[i].Coupons, got[i].Coupons)
				require.Equal(t, tc.expected[i].Date, got[i].Date)
				require.Equal(t, tc.expected[i].RecoveryRate, got[i].RecoveryRate)
				require.Equal(t, tc.expected[i].Upfront, got[i].Upfront)
			}
		})
	}
}
