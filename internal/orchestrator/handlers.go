package orchestrator

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/dimakirio/calculatorv1/internal/models"
	"github.com/dimakirio/calculatorv1/pkg/config"
	"github.com/dimakirio/calculatorv1/pkg/logger"
	"github.com/Knetic/govaluate" // Импорт библиотеки для вычисления выражений
	"github.com/google/uuid"
	"github.com/dimakirio/calculatorv1/internal/auth"
)

var (
	expressions = make(map[string]models.Expression)
	mu          sync.Mutex
)

type Orchestrator struct {
	log *logger.Logger
	cfg *config.Config
}

func NewOrchestrator(log *logger.Logger, cfg *config.Config) *Orchestrator {
	return &Orchestrator{log: log, cfg: cfg}
}

func (o *Orchestrator) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}

	// Проверяем корректность выражения
	if !isValidExpression(req.Expression) {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid expression")
		return
	}

	// Вычисляем выражение
	result, err := evaluateExpression(req.Expression)
	if err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Failed to evaluate expression")
		return
	}

	id := uuid.New().String()
	mu.Lock()
	expressions[id] = models.Expression{
		ID:     id,
		Status: "completed",
		Result: result,
	}
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (o *Orchestrator) HandleGetExpressions(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var exprs []models.Expression
	for _, expr := range expressions {
		exprs = append(exprs, expr)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": exprs})
}

func (o *Orchestrator) HandleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/expressions/"):]
	mu.Lock()
	expr, exists := expressions[id]
	mu.Unlock()

	if !exists {
		writeJSONError(w, http.StatusNotFound, "Expression not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (o *Orchestrator) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	if req.Login == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Login and password required")
		return
	}
	db, err := models.NewDatabase(o.cfg.DBPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer db.Close()
	repo := models.NewUserRepository(db.DB())
	if err := repo.Create(req.Login, req.Password); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (o *Orchestrator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, "Invalid request body")
		return
	}
	db, err := models.NewDatabase(o.cfg.DBPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer db.Close()
	repo := models.NewUserRepository(db.DB())
	user, err := repo.GetByLogin(req.Login)
	if err != nil || !repo.ValidatePassword(user, req.Password) {
		writeJSONError(w, http.StatusUnauthorized, "Invalid login or password")
		return
	}
	jwtService := auth.NewJWTService(o.cfg.JWTSecret)
	token, err := jwtService.GenerateToken(user.ID, user.Login)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// evaluateExpression вычисляет значение выражения
func evaluateExpression(expression string) (float64, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return 0, err
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return 0, err
	}

	return result.(float64), nil
}

// isValidExpression проверяет корректность выражения
func isValidExpression(expression string) bool {
	// Простая проверка на наличие некорректных символов
	for _, char := range expression {
		if !isValidCharacter(char) {
			return false
		}
	}
	return true
}

// isValidCharacter проверяет, является ли символ допустимым
func isValidCharacter(char rune) bool {
	// Разрешенные символы: цифры, операторы (+,-,*,/), пробелы, скобки
	return (char >= '0' && char <= '9') ||
		char == '+' || char == '-' || char == '*' || char == '/' ||
		char == ' ' || char == '(' || char == ')'
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
