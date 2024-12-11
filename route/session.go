package route

import (
	"crypto/rsa"
	"encoding/json"
	"log"
	"os"
	"time"
	"vector-ai/model"

	"github.com/golang-jwt/jwt/v4"
	ws "github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

var (
	verifyKey *rsa.PublicKey
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000000
)

// Session is a middleman between the websocket connection and the tracker.
type Session struct {
	tracker        *Tracker
	handler        *Handler
	conn           *connection // The websocket connection.
	orgId          string
	workspaceId    string
	conversationId string
	token          *jwt.Token
	manifest       map[string]map[string]model.FileRecord
	chatbot        *openai.LLM
	embedder       *embeddings.EmbedderImpl
	readErr        chan error
	writeErr       chan error
}

type connection struct {
	ws   *ws.Conn // formerly goyave.websocket
	send chan model.Envelope
}

func (c *Session) pump() error {
	c.tracker.register <- c
	go c.writePump()
	go c.readPump()

	var err error
	select {
	case e := <-c.readErr:
		err = e
	case e := <-c.writeErr:
		err = e
		if err == nil {
			// tracker closing, wait for readPump to return
			<-c.readErr

		}
	}

	return err
}

func (s *Session) readPump() {
	c := s.conn
	defer func() {
		s.tracker.unregister <- s
		c.ws.Close() // previously (1, "unregistered")
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		var envelope model.UserWebsocketEnvelope
		// var query model.Message

		// Convert incoming object to []byte
		messageType, msg, err := c.ws.ReadMessage()

		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway) {
				log.Printf("error on c.ws.ReadMessage(): %v", err)
			} else {
				log.Printf("error on c.ws.ReadMessage(): %v", err)
			}
			break
		}

		if messageType == ws.BinaryMessage && msg[0] == 0xB {
			// Read the header length
			headerLength := int(msg[1])<<8 + int(msg[2])
			log.Println("header length:", headerLength)

			// Parse the header
			var header model.CoreDocumentProps
			err := json.Unmarshal(msg[3:3+headerLength], &header)
			if err != nil {
				log.Println("json unmarshal:", err)
				continue
			}

			// Process the binary data
			binaryData := msg[3+headerLength:]

			// Save binary data and process
			s.UploadFile(header, binaryData)

		} else {

			// []byte -> model.Message
			if err := json.Unmarshal([]byte(msg), &envelope); err != nil {
				log.Fatal(err)
			}

			// user is sending vss or analysis query
			if envelope.Message != nil {
				var message = *envelope.Message
				if len(message.VssText) > 0 {
					// Perform Vector Similarity Search and broadcast response
					go func(message model.WebSocketsMessage) {
						s.QueryVss(message)
					}(message)
				} else {
					// Query AI with userMessage and history and broadcast AI response
					go func(message model.WebSocketsMessage) {
						aiMessage := s.QueryAnalysis(message)
						s.tracker.Broadcast(model.AiResponse(aiMessage, s.workspaceId, s.conversationId))
					}(message)
				}
			}

			if envelope.Token != nil {
				tokenString := *envelope.Token

				// TODO: duplicate code
				// Parse the token
				token, err := jwt.ParseWithClaims(tokenString, &model.ClerkClaims{}, func(token *jwt.Token) (interface{}, error) {

					var rsa string
					if env := os.Getenv("GOYAVE_ENV"); env == "production" {
						rsa = "resources/rsa/public.pem"
					} else {
						rsa = "resources/rsa/local.pem"
					}

					// Use public key to verify
					data, err := os.ReadFile(rsa)
					check(err)

					verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(data)
					check(err)

					return verifyKey, nil
				})
				check(err)

				//assign token in session
				s.token = token
			}

			// TODO: add authorization

			// user is sending drive folders for synchronization
			if envelope.DriveFolders != nil {
				var folders = *envelope.DriveFolders

				go func(folders model.DriveFolders) {
					s.SyncDrive(folders.Folders) // UploadDrive
				}(folders)
			}

			if envelope.SyncFolders != nil {
				var folders = *envelope.SyncFolders

				go func(folders model.SyncFolders) {
					s.SyncDrive(folders.Folders)
				}(folders)
			}
		}
	}
}

// write writes a message with the given message type and payload.
// func (c *connection) write(mt int, payload []byte) error {
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeMessage(mt int, payload model.Envelope) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	// c.ws.SetWriteDeadline(time.0)
	buffer, err := json.Marshal(payload)
	check(err)
	return c.ws.WriteMessage(mt, buffer)
}

func (s *Session) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		//c.ws.Close(1, "ticker stopped")
		c.ws.Close()
	}()

	for {
		select {

		case message, ok := <-c.send:
			if !ok {
				c.write(ws.CloseMessage, []byte{})
				return
			}
			if err := c.writeMessage(ws.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(ws.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
