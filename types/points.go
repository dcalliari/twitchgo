package types

type UserData struct {
	Username   string `json:"username"`
	Points     int    `json:"points"`
	GambleLoss int    `json:"gamble_loss"`
}

type PointsDatabase interface {
	GetPoints(username string) int
	AddPoints(username string, amount int) error
	SubtractPoints(username string, amount int) error
	GetGambleLoss(username string) int
	AddGambleLoss(username string, amount int) error
	TransferPoints(sender, receiver string, amount int) error
	Gamble(username string, wager int, format string, winOdds float64) (string, int, int, error)
	GetTopPoints(limit int) ([]string, []int)
	GetTopGambleLoss(limit int) ([]string, []int)
	GetRank(username string) (int, int)
	ValidateUser(username string) error
	SaveToFile() error
	LoadFromFile() error
}
