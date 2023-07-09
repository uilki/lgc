package server

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	pb "github.com/uilki/lgc/api/server/generated"
	"google.golang.org/protobuf/proto"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type participant struct {
	uuid uuid.UUID
	conn *websocket.Conn
	mes  chan []byte
}

func (c *participant) readMessages(ctx context.Context) {
	srv := ctx.Value(serverKey).(*Server)
	controller := ctx.Value(controllerKey).(*dispatcher)

	defer func() {
		controller.removePaticipant <- c
		c.conn.Close()
		srv.wg.Done()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				srv.log(ERROR, err.Error())
			}
			break
		}

		backlogMes := pb.Message{
			TimeStamp: time.Unix(time.Now().Unix(), 0).UTC().String(),
			Name:      srv.user[c.uuid].name,
			Message:   string(message),
		}

		if err = srv.history.Update(&backlogMes); err != nil {
			srv.log(ERROR, err.Error())
		}

		message, err = proto.Marshal(&backlogMes)
		if err != nil {
			srv.log(ERROR, err.Error())
			continue
		}

		controller.broadcast <- bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	}
}

func (c *participant) writeMessages(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 50)
	srv := ctx.Value(serverKey).(*Server)
	defer func() {
		ticker.Stop()
		srv.wg.Done()
	}()

	for {
		select {
		case message, ok := <-c.mes:
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				srv.log(ERROR, err.Error())
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				srv.log(ERROR, err.Error())
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				srv.log(ERROR, err.Error())
				return
			}
		case <-ctx.Done():
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}
