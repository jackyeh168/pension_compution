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
	AnnualExpense  float64 `json:"annualExpense"`
	RealReturnRate float64 `json:"realReturnRate"`
	GovtPension    float64 `json:"govtPension"`
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
	minCap, records := FindMinimumCapital(
		req.AnnualExpense, req.RealReturnRate, req.GovtPension, years,
	)

	// 敏感度分析：3%、使用者選擇的報酬率、7% 三種情境
	rateSet := map[float64]bool{0.03: true, req.RealReturnRate: true, 0.07: true}
	rates := make([]float64, 0, len(rateSet))
	for rate := range rateSet {
		rates = append(rates, rate)
	}
	sort.Float64s(rates)

	sensitivity := make([]SensitivityRow, 0, len(rates))
	for _, rate := range rates {
		cap, _ := FindMinimumCapital(
			req.AnnualExpense, rate, req.GovtPension, years,
		)
		sensitivity = append(sensitivity, SensitivityRow{ReturnRate: rate, MinCapital: cap})
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
