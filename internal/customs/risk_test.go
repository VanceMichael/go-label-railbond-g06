package customs_test

import (
	"github.com/VanceMichael/go-base-railbond-g06/internal/customs"
	"testing"
)

func TestRiskAssessmentThresholds(t *testing.T) {
	a := customs.Assess([]customs.RiskSignal{{Code: "value", Weight: 80}})
	if a.Level != customs.High || a.Score != 80 {
		t.Fatal(a)
	}
	if err := customs.ValidateAssessment(a); err != nil {
		t.Fatal(err)
	}
}
func TestDocumentChecklist(t *testing.T) {
	if err := customs.CheckDocuments([]customs.DocumentCheck{{Kind: "invoice", Present: true, Valid: true}}); err != nil {
		t.Fatal(err)
	}
	if err := customs.CheckDocuments([]customs.DocumentCheck{{Kind: "invoice", Present: false, Valid: false, Message: "missing"}}); err == nil {
		t.Fatal("missing document accepted")
	}
}
