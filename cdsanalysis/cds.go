package main

import (
	"cdsanalysis/integration"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/edgelaboratories/eve/pkg/asset"
	"github.com/edgelaboratories/eve/pkg/marketdata"
	"github.com/edgelaboratories/go-libraries/date"
	"github.com/edgelaboratories/go-libraries/daycount"
	log "github.com/sirupsen/logrus"
)

type CDSInput struct {
	ID              string                   `json:"issuer"`
	UpfrontPayments map[Tenor]float64        `json:"spreads"`
	InterestCurve   marketdata.TermStructure `json:"interestCurve"`
	RecoveryRate    float64                  `json:"recoveryRate"`
	CouponRate      float64                  `json:"couponRate"`
	Name            string                   `json:"name"`
	Date            date.Date                `json:"date"`
	Frequency       string                   `json:"frequency"`
}

type CDSAsset struct {
	ID           string              `json:"issuer"`
	Maturity     date.Date           `json:"maturity"`
	Frequency    string              `json:"frequency"`
	Coupons      []asset.FixedCoupon `json:"coupons"`
	Date         date.Date           `json:"date"`
	RecoveryRate float64             `json:"recoveryRate"`
	Upfront      float64             `json:"upfront"`

	InterestCurve *InterestRateCurveRepresentation
}

func loadCDSData(issuers []string) map[string]CDSInput {
	cdsData := make(map[string]CDSInput, len(issuers))
	// Load CDS data
	for _, issuer := range issuers {
		// Load CDS data for issuer
		path := "./data/" + issuer + ".json"
		content, err := os.ReadFile(path)
		if err != nil {
			log.Errorf("could not read %s: %v", path, err)

			continue
		}

		var cdsInput CDSInput
		err = json.Unmarshal(content, &cdsInput)
		if err != nil {
			log.Errorf("could not unmarshal %s: %v", path, err)

			continue
		}

		cdsData[issuer] = cdsInput
	}

	return cdsData
}

func inputToAsset(cdsInput CDSInput) ([]CDSAsset, error) {
	// Convert CDS input to CDS asset
	assets := make([]CDSAsset, 0, len(cdsInput.UpfrontPayments))
	for tenor, uf := range cdsInput.UpfrontPayments {
		// Create CDS asset
		yf, err := tenor.ToYearFraction()
		if err != nil {
			return nil, err
		}
		if yf < 0.25 {
			// to update if we wish to use shorter maturity.

			continue
		}

		firstCouponDate := lastIMMDate(cdsInput.Date)
		// firstCouponDate := lastIMMDate(cdsInput.Date)
		nextCouponDate := nextIMMDate(cdsInput.Date)
		maturity, err := tenor.ShiftDateByTenor(nextIMMDate(cdsInput.Date))
		if err != nil {
			return nil, err
		}

		// Build coupons.
		coupons, err := generateCoupons(cdsInput.CouponRate, firstCouponDate, nextCouponDate, maturity)
		if err != nil {
			return nil, err
		}

		// Build interest rate curve.
		interestCurve := InterestRateCurveRepresentation{
			Data:  cdsInput.InterestCurve,
			Model: &PiecewiseLinearCurveModel{},
		}
		err = interestCurve.Build()
		if err != nil {
			return nil, err
		}

		asset := CDSAsset{
			ID:           cdsInput.ID,
			Maturity:     maturity,
			Frequency:    cdsInput.Frequency,
			Coupons:      coupons,
			Date:         cdsInput.Date,
			RecoveryRate: cdsInput.RecoveryRate,
			Upfront:      uf / 100.0,

			InterestCurve: &interestCurve,
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func calibrateCreditCurves(issuers []string) (map[string]map[string]float64, error) {
	cdsData := loadCDSData(issuers)

	creditCurves := make(map[string]map[string]float64, len(cdsData))
	for _, cdsInput := range cdsData {
		assets, err := inputToAsset(cdsInput)
		if err != nil {
			log.Errorf("could not convert input to asset: %v", err)

			continue
		}

		creditCurve, err := calibrateCreditTermStructure(assets)
		if err != nil {
			log.Errorf("could not calibrate credit term structure: %v", err)

			continue
		}

		creditCurves[cdsInput.ID] = creditCurve
	}

	return creditCurves, nil
}

func calibrateCreditTermStructure(cdsAssets []CDSAsset) (map[string]float64, error) {
	e := extractor{configuration: DefaultConfiguration()}
	// e.configuration.Parametrization = ParametrizedLongShortNS{}
	e.configuration.Parametrization = ParametrizedLongShortNS{}

	curve, err := e.extractCurve(e.configuration.Parametrization, cdsAssets)
	if err != nil {
		return nil, err
	}

	curveObj := *curve
	fmt.Println(curveObj)
	return map[string]float64{
		"M12": curveObj.Value(1.0),
		"Y7":  curveObj.Value(7.0),
		"Y20": curveObj.Value(20.0),
		"Y50": curveObj.Value(50.0),
	}, nil
}

// generateCoupons generates the fixed coupons for a CDS asset.
// Each of them are paid on IMM dates.
func generateCoupons(spread float64, firstCouponDate, nextCouponDate, maturity date.Date) ([]asset.FixedCoupon, error) {
	// Generate coupons
	coupons := make([]asset.FixedCoupon, 0)

	initialFixingDate := firstCouponDate
	paymentDate := nextCouponDate
	for !paymentDate.After(maturity) {
		coupon := asset.FixedCoupon{
			Type:          asset.FixedType,
			FixedRate:     spread,
			InitialFixing: initialFixingDate,
			PaymentDate:   paymentDate,
		}

		coupons = append(coupons, coupon)

		initialFixingDate = paymentDate
		paymentDate = nextIMMDate(paymentDate)
	}

	return coupons, nil
}

func priceCDSSum(cdsAssets []CDSAsset, creditTS TermStructure) float64 {
	// Calculate the sum of the CDS prices
	price := 0.0
	for _, cds := range cdsAssets {
		cdsPrice := priceCDS(cds, creditTS)
		price += cdsPrice * cdsPrice
	}

	return price
}

func priceCDS(cds CDSAsset, creditTS TermStructure) float64 {
	// Calculate premium and protection leg
	premium := premiumLeg(cds, creditTS)
	protection := protectionLeg(cds, creditTS)

	return protection - premium - cds.Upfront
}

func premiumLeg(cds CDSAsset, creditTS TermStructure) float64 {
	// Calculate premium leg
	referenceDate := cds.Date

	premium := 0.0
	for _, coupon := range cds.Coupons {
		couponPaymentDate := coupon.CouponPaymentDate()
		if !couponPaymentDate.After(referenceDate) {
			continue
		}

		yfCouponInitialFixing := daycount.YearFraction(referenceDate, coupon.CouponInitialFixing(), daycount.ActualThreeSixty)
		if yfCouponInitialFixing < 0 {
			yfCouponInitialFixing = 0
		}
		yfCouponPayment := daycount.YearFraction(referenceDate, couponPaymentDate, daycount.ActualThreeSixty)

		couponValue := coupon.FixedRate * daycount.YearFraction(coupon.CouponInitialFixing(), coupon.CouponPaymentDate(), daycount.ActualThreeSixty)

		survivalProbabilityAverage := 0.5 * (survivalProbability(creditTS, yfCouponPayment) + survivalProbability(creditTS, yfCouponInitialFixing))
		discountFactor := survivalProbabilityAverage * cds.InterestCurve.DiscountFactor(yfCouponPayment)
		premium += couponValue * discountFactor
	}

	return premium
}

func protectionLeg(cds CDSAsset, creditTS TermStructure) float64 {
	// Calculate protection leg
	referenceDate := cds.Date

	yfReference := daycount.YearFraction(referenceDate, referenceDate, daycount.ActualThreeSixty)
	yfMaturity := daycount.YearFraction(referenceDate, cds.Maturity, daycount.ActualThreeSixty)

	variate := func(yf float64) float64 {
		return cds.InterestCurve.DiscountFactor(yf) * math.Exp(-creditTS.Value(yf)*yf)
	}

	integrationTerm := integration.Integrate(func(yf float64) float64 {
		return variate(yf) * survivalProbabilityDensity(creditTS, yf)
	}, yfReference, yfMaturity)

	return (1.0 - cds.RecoveryRate) * integrationTerm
}
