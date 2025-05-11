package orchestrator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dimakirio/calculatorv1/internal/models"
	"github.com/dimakirio/calculatorv1/pkg/config"
	"github.com/dimakirio/calculatorv1/pkg/logger"
)

func TestHandleCalculate(t *testing.T) {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	orchestrator := NewOrchestrator(log, cfg)

	// Тест 1: Корректное выражение
	reqBody := `{"expression": "2 + 2 * 2"}`
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(orchestrator.HandleCalculate)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusCreated)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if _, exists := response["id"]; !exists {
		t.Errorf("Handler did not return an ID")
	}

	// Тест 2: Некорректное выражение
	reqBody = `{"expression": "2 + * 2"}`
	req, err = http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusUnprocessableEntity)
	}
}

func TestHandleGetExpressions(t *testing.T) {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	orchestrator := NewOrchestrator(log, cfg)

	// Добавляем тестовое выражение
	orchestrator.HandleCalculate(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(`{"expression": "2 + 2"}`)))

	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(orchestrator.HandleGetExpressions)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var response map[string][]models.Expression
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if len(response["expressions"]) == 0 {
		t.Errorf("Handler did not return any expressions")
	}
}

func TestHandleGetExpressionByID(t *testing.T) {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)
	orchestrator := NewOrchestrator(log, cfg)

	// Добавляем тестовое выражение
	rr := httptest.NewRecorder()
	orchestrator.HandleCalculate(rr, httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(`{"expression": "2 + 2"}`)))

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	id := response["id"]

	req, err := http.NewRequest("GET", "/api/v1/expressions/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(orchestrator.HandleGetExpressionByID)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var exprResponse map[string]models.Expression
	if err := json.Unmarshal(rr.Body.Bytes(), &exprResponse); err != nil {
		t.Fatal(err)
	}

	if exprResponse["expression"].ID != id {
		t.Errorf("Handler returned wrong expression ID: got %v, want %v", exprResponse["expression"].ID, id)
	}
}

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		hasError   bool
	}{
		{"2 + 2 * 2", 6, false},
		{"(2 + 2) * 2", 8, false},
		{"2 + * 2", 0, true}, // Некорректное выражение
	}

	for _, test := range tests {
		result, err := evaluateExpression(test.expression)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for expression: %s, but got none", test.expression)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for expression: %s, error: %v", test.expression, err)
			}
			if result != test.expected {
				t.Errorf("For expression: %s, expected: %v, got: %v", test.expression, test.expected, result)
			}
		}
	}
}

func TestRegisterAndLogin(t *testing.T) {
	// Удаляем тестовую БД перед запуском
	_ = os.Remove("test.db")
	cfg := &config.Config{
		DBPath:    "test.db",
		JWTSecret: "testsecret",
	}
	log := logger.NewLogger("info")
	o := NewOrchestrator(log, cfg)

	// Регистрация
	registerBody := `{"login":"testuser","password":"testpass"}`
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBufferString(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	o.HandleRegister(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Логин
	loginBody := `{"login":"testuser","password":"testpass"}`
	req = httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	o.HandleLogin(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["token"] == "" {
		t.Fatalf("no token in response")
	}
}
