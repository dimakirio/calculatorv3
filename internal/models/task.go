package models

type Task struct {
	ID            string
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime int
	Status        string
	Result        float64 // Добавлено поле Result
}
