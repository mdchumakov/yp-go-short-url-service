package stats

// Response представляет структуру данных для ответа на запрос статистики.
type Response struct {
	URLsCount  int64 `json:"urls"`
	UsersCount int64 `json:"users"`
}
