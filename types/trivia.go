package types

type TriviaQuestion struct {
	ID       string
	Question string
	Answer   string
	Enabled  bool
}

type TriviaDatabase interface {
	GetRandomQuestion() *TriviaQuestion
	GetQuestionByID(id string) *TriviaQuestion
	AddQuestion(question TriviaQuestion)
	EnableQuestion(id string) bool
	DisableQuestion(id string) bool
	GetQuestionCount() int
	GetEnabledQuestionCount() int
	SaveToJSONFile(filename string) error
	ReloadFromFile() error
}
