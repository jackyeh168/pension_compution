package main

import "math"

// YearRecord 記錄單一年度的模擬結果。
type YearRecord struct {
	Year           int     `json:"year"`
	BeginAsset     float64 `json:"beginAsset"`
	LivingCost     float64 `json:"livingCost"`
	GovtPension    float64 `json:"govtPension"`
	StockWithdrawn float64 `json:"stockWithdrawn"`
	EndAsset       float64 `json:"endAsset"`
}

// Simulate 逐年模擬退休金消耗，回傳每年記錄。
// 使用實質報酬率，生活費固定不隨通膨調整。
func Simulate(initialCapital, annualExpense, realReturnRate, govtPension float64, years int) []YearRecord {
	records := make([]YearRecord, 0, years)
	capital := initialCapital

	for yr := 1; yr <= years; yr++ {
		beginning := capital
		withdrawn := annualExpense - govtPension
		if withdrawn < 0 {
			withdrawn = 0
		}
		afterWithdrawal := capital - withdrawn
		ending := afterWithdrawal * (1 + realReturnRate)

		records = append(records, YearRecord{
			Year:           yr,
			BeginAsset:     math.Round(beginning*100) / 100,
			LivingCost:     annualExpense,
			GovtPension:    govtPension,
			StockWithdrawn: math.Round(withdrawn*100) / 100,
			EndAsset:       math.Round(ending*100) / 100,
		})
		capital = ending
	}
	return records
}

// survives 判斷初始資金是否能撐過指定年數。
func survives(capital, annualExpense, realReturnRate, govtPension float64, years int) bool {
	bal := capital
	withdrawn := annualExpense - govtPension
	if withdrawn < 0 {
		withdrawn = 0
	}
	for yr := 1; yr <= years; yr++ {
		bal = (bal - withdrawn) * (1 + realReturnRate)
		if bal < 0 {
			return false
		}
	}
	return true
}

// FindMinimumCapital 用二分法搜尋使股票資產在 years 年內不歸零的最小初始資金。
func FindMinimumCapital(annualExpense, realReturnRate, govtPension float64, years int) (int, []YearRecord) {
	upper := annualExpense * float64(years) // 無投報最差情境
	lower := 0.0

	for upper-lower > 0.5 {
		mid := (lower + upper) / 2
		if survives(mid, annualExpense, realReturnRate, govtPension, years) {
			upper = mid
		} else {
			lower = mid
		}
	}

	minCapital := int(math.Round(upper))
	records := Simulate(float64(minCapital), annualExpense, realReturnRate, govtPension, years)
	return minCapital, records
}
