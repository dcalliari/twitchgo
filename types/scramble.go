package types

type ScrambleWord struct {
	ID      string `json:"id"`
	Word    string `json:"word"`
	Enabled bool   `json:"enabled"`
}

type ScrambleDatabase interface {
	GetRandomWord() *ScrambleWord
	ReloadWords() error
}
