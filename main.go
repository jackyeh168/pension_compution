package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
)

//go:embed templates/index.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/index.html"))

// CalcRequest 前端傳入的計算參數。
type CalcRequest struct {
	AnnualExpense float64 `json:"annualExpense"`
	NominalReturn float64 `json:"nominalReturn"`
	InflationRate float64 `json:"inflationRate"`
	GovtPension   float64 `json:"govtPension"`
}

// CalcResponse 回傳給前端的計算結果。
type CalcResponse struct {
	MinCapital  int              `json:"minCapital"`
	Records     []YearRecord     `json:"records"`
	Sensitivity []SensitivityRow `json:"sensitivity"`
}

// SensitivityRow 敏感度分析的單一列。
type SensitivityRow struct {
	ReturnRate float64 `json:"returnRate"`
	MinCapital int     `json:"minCapital"`
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func handleCalc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "僅支援 POST", http.StatusMethodNotAllowed)
		return
	}

	var req CalcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "參數格式錯誤", http.StatusBadRequest)
		return
	}

	const years = 45

	// 實質報酬率 = 預期報酬率 − 預期通膨率
	realReturn := req.NominalReturn - req.InflationRate

	minCap, records := FindMinimumCapital(
		req.AnnualExpense, realReturn, req.GovtPension, years,
	)

	// 敏感度分析：以使用者通膨率為基準，比較三種名目報酬率對應的實質報酬率
	nominals := []float64{0.06, req.NominalReturn, 0.10}
	rateSet := map[float64]bool{}
	for _, n := range nominals {
		rateSet[n] = true
	}
	rates := make([]float64, 0, len(rateSet))
	for rate := range rateSet {
		rates = append(rates, rate)
	}
	sort.Float64s(rates)

	sensitivity := make([]SensitivityRow, 0, len(rates))
	for _, nominal := range rates {
		rr := nominal - req.InflationRate
		cap, _ := FindMinimumCapital(
			req.AnnualExpense, rr, req.GovtPension, years,
		)
		sensitivity = append(sensitivity, SensitivityRow{ReturnRate: nominal, MinCapital: cap})
	}

	resp := CalcResponse{
		MinCapital:  minCap,
		Records:     records,
		Sensitivity: sensitivity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/calc", handleCalc)

	addr := ":" + port
	log.Printf("退休金計算器啟動於 http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
