package utils

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"twitchgo/types"
)

type InMemoryTriviaDB struct {
	questions []types.TriviaQuestion
	rng       *rand.Rand
}

func NewInMemoryTriviaDB() *InMemoryTriviaDB {
	db := &InMemoryTriviaDB{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	db.loadDefaultQuestions()
	return db
}

func (db *InMemoryTriviaDB) GetRandomQuestion() *types.TriviaQuestion {
	enabledQuestions := db.getEnabledQuestions()
	if len(enabledQuestions) == 0 {
		return nil
	}
	return &enabledQuestions[db.rng.Intn(len(enabledQuestions))]
}

func (db *InMemoryTriviaDB) GetQuestionByID(id string) *types.TriviaQuestion {
	for _, q := range db.questions {
		if q.ID == id {
			return &q
		}
	}
	return nil
}

func (db *InMemoryTriviaDB) AddQuestion(question types.TriviaQuestion) {
	db.questions = append(db.questions, question)
}

func (db *InMemoryTriviaDB) EnableQuestion(id string) bool {
	for i, q := range db.questions {
		if q.ID == id {
			db.questions[i].Enabled = true
			return true
		}
	}
	return false
}

func (db *InMemoryTriviaDB) DisableQuestion(id string) bool {
	for i, q := range db.questions {
		if q.ID == id {
			db.questions[i].Enabled = false
			return true
		}
	}
	return false
}

func (db *InMemoryTriviaDB) GetQuestionCount() int {
	return len(db.questions)
}

func (db *InMemoryTriviaDB) GetEnabledQuestionCount() int {
	return len(db.getEnabledQuestions())
}

func (db *InMemoryTriviaDB) SaveToJSONFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(db.questions); err != nil {
		return err
	}

	log.Printf("Successfully saved %d trivia questions to %s", len(db.questions), filename)
	return nil
}

func (db *InMemoryTriviaDB) ReloadFromFile() error {
	return db.loadFromJSONFile("data/trivia_questions.json")
}

func (db *InMemoryTriviaDB) getEnabledQuestions() []types.TriviaQuestion {
	var enabled []types.TriviaQuestion
	for _, q := range db.questions {
		if q.Enabled {
			enabled = append(enabled, q)
		}
	}
	return enabled
}

func (db *InMemoryTriviaDB) loadDefaultQuestions() {
	if err := db.loadFromJSONFile("data/trivia_questions.json"); err != nil {
		log.Printf("Failed to load trivia questions from JSON file: %v", err)
		log.Println("Loading fallback default questions...")
		db.loadFallbackQuestions()
	}
}

func (db *InMemoryTriviaDB) loadFromJSONFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var questions []types.TriviaQuestion
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&questions); err != nil {
		return err
	}

	db.questions = questions
	log.Printf("Successfully loaded %d trivia questions from %s", len(questions), filename)
	return nil
}

func (db *InMemoryTriviaDB) loadFallbackQuestions() {
	db.questions = []types.TriviaQuestion{
		{ID: "t0000000001", Question: "Qual reality show brasileiro confina participantes em uma casa e é conhecido pela sigla BBB?", Answer: "Big Brother Brasil", Enabled: true},
		{ID: "t0000000002", Question: "Qual o nome do streamer brasileiro famoso pelo bordão 'Aí pai, para!'?", Answer: "Casimiro", Enabled: true},
		{ID: "t0000000003", Question: "Qual o nome da personagem de novela, interpretada por Adriana Esteves, que se tornou um meme mundial com a frase 'Me serve, vadia'?", Answer: "Carminha", Enabled: true},
		{ID: "t0000000004", Question: "Qual o nome do cantor que viralizou com a música 'Caneta Azul'?", Answer: "Manoel Gomes", Enabled: true},
		{ID: "t0000000005", Question: "Qual o nome do podcast apresentado por Igão e Mítico, um dos mais populares do Brasil?", Answer: "Podpah", Enabled: true},
		{ID: "t0000000006", Question: "De qual grupo de humor se originou o meme 'Nheco nheco no potinho'?", Answer: "Hermes e Renato", Enabled: true},
		{ID: "t0000000007", Question: "Qual o nome da personagem de um vídeo viral que diz 'Meu nome é Júlia, e eu gosto de pular'?", Answer: "Júlia", Enabled: true},
		{ID: "t0000000008", Question: "Qual o nome da cantora que viralizou com o hit 'Que Tiro Foi Esse'?", Answer: "Jojo Todynho", Enabled: true},
		{ID: "t0000000009", Question: "O meme 'Nazaré Confusa', com a personagem olhando para os lados, veio de qual novela?", Answer: "Senhora do Destino", Enabled: true},
		{ID: "t0000000010", Question: "Complete o bordão do vídeo viral da banda New Dingo: 'Acorda, Pedrinho, que hoje tem...'", Answer: "campeonato", Enabled: true},
		{ID: "t0000000011", Question: "Qual o nome do youtuber que ficou famoso pelo quadro '5inco Minutos'?", Answer: "Kéfera", Enabled: true},
		{ID: "t0000000012", Question: "Qual o nome do grupo de humor do YouTube famoso por esquetes como 'Reunião de Condomínio'?", Answer: "Porta dos Fundos", Enabled: true},
		{ID: "t0000000013", Question: "Qual o nome do personagem de um vídeo viral que ficou conhecido como a 'Grávida de Taubaté'?", Answer: "Maria Verônica", Enabled: true},
		{ID: "t0000000014", Question: "Qual o nome do streamer brasileiro de jogos conhecido como 'O Pai' e pelo bordão 'Respeita o F'?", Answer: "Gaules", Enabled: true},
		{ID: "t0000000015", Question: "Qual o nome do personagem criado por um humorista baiano que ficou famoso por seus vídeos no WhatsApp e pelo bordão 'Ô, psit'?", Answer: "Dum Ice", Enabled: true},
	}
}
