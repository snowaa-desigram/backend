package service

// Service — use-cases ping. Заглушка: отвечает pong.
type Service struct{}

func New() *Service { return &Service{} }

// Ping возвращает ответ на сообщение.
func (s *Service) Ping(message string) string {
	return "pong: " + message
}
