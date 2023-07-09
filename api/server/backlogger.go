package server

import pb "github.com/uilki/lgc/api/server/generated"

//go:generate mockery --name Backlogger
type Backlogger interface {
	GetHistory() ([]pb.Message, error)
	Update(*pb.Message) error
	Close()
}

func NewBackLogger(pass string) (Backlogger, error) {
	if pass != "" {
		sqlBacklog, err := NewSqlBacklog(pass)
		if err != nil {
			return nil, err
		}
		return Backlogger(sqlBacklog), nil
	} else {
		return Backlogger(&backlog{}), nil
	}
}
