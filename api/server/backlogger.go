package server

type Backlogger interface {
	GetHistory() ([]Message, error)
	Update(Message) error
	Close()
}
