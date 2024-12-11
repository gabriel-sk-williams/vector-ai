package route

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"vector-ai/constants"
	"vector-ai/model"

	"github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"goyave.dev/goyave/v4"
)

type Tracker struct {
	sessions   map[string]map[string]map[*connection]bool
	broadcast  chan model.Envelope
	register   chan *Session
	unregister chan *Session

	ctx     context.Context
	cancel  context.CancelFunc
	handler Handler
}

// create a new tracker //pgx *pgxpool.Pool, qdr *qdrantgo.Client
func NewTracker() *Tracker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Tracker{
		broadcast:  make(chan model.Envelope),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		sessions:   make(map[string]map[string]map[*connection]bool),
		ctx:        ctx,
		cancel:     cancel,
		// handler:    h,
	}
}

func (t *Tracker) Run() {
	done := make(chan struct{}, 1)
	defer t.cancel()
	goyave.RegisterShutdownHook(func() {
		t.cancel()
		<-done
	})

	for {
		select {

		case <-t.ctx.Done():
			wg := &sync.WaitGroup{}

			for workspaceId := range t.sessions {
				wg.Add(len(t.sessions[workspaceId]))
				for conversationId := range t.sessions[workspaceId] {
					for conn := range t.sessions[workspaceId][conversationId] {
						go func(c *connection) {
							close(c.send)
							if err := c.ws.Close(); err != nil { // goyave: CloseNormal()
								goyave.ErrLogger.Println(err)
							}
							<-t.unregister // Wait for readPump to return
							wg.Done()
						}(conn)
					}
					delete(t.sessions[workspaceId], conversationId)
				}
				delete(t.sessions, workspaceId)
			}
			wg.Wait()
			done <- struct{}{}
			return

		case s := <-t.register:
			fmt.Println("registering", s.workspaceId)
			conversations := t.sessions[s.workspaceId]
			if conversations == nil {
				t.sessions[s.workspaceId] = make(map[string]map[*connection]bool)
			}

			sessions := t.sessions[s.workspaceId][s.conversationId]
			if sessions == nil {
				sessions = make(map[*connection]bool)
				t.sessions[s.workspaceId][s.conversationId] = sessions
			}
			t.sessions[s.workspaceId][s.conversationId][s.conn] = true

		case s := <-t.unregister:
			fmt.Println("unregistering", s.workspaceId)
			fmt.Println("")
			conversations := t.sessions[s.workspaceId]
			if conversations != nil {
				connections := t.sessions[s.workspaceId][s.conversationId]
				if connections != nil {
					if _, ok := connections[s.conn]; ok {
						delete(connections, s.conn)
						close(s.conn.send)
						if len(connections) == 0 {
							delete(t.sessions[s.workspaceId], s.conversationId)
						}
					}
				}
			}

		case m := <-t.broadcast:

			var pool map[*connection]bool

			if m.ConversationID != "nil" { // broadcasting to conversation
				pool = t.sessions[m.WorkspaceID][m.ConversationID]

			} else { // broadcasting to workspace
				pool = make(map[*connection]bool)
				conversations := t.sessions[m.WorkspaceID]
				for _, sessions := range conversations {
					for conn, cBool := range sessions {
						pool[conn] = cBool
					}
				}
			}

			for c := range pool {
				c.send <- m
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	//ReadBufferSize: 1024,
	//WriteBufferSize: 1024,
}

// Serve is the websocket Handler for jobs.
//func (t *Tracker) Serve(c *websocket.Conn, req *goyave.Request) error {
//res *goyave.Response,
//func (t *Tracker) Serve(w http.ResponseWriter, req *http.Request) error {
//c, err := upgrader.Upgrade(w, req, nil)
//defer conn.Close()

func (t *Tracker) Serve(res *goyave.Response, req *goyave.Request) {
	c, err := upgrader.Upgrade(res, req.Request(), nil)
	check(err)

	orgId := req.Params["orgId"]
	workspaceId := req.Params["workspaceId"]
	conversationId := req.Params["conversationId"]

	opts := []openai.Option{
		openai.WithModel(constants.LLM),
		openai.WithEmbeddingModel(constants.Embedder),
	}

	llm, err := openai.New(opts...)
	check(err)

	// embeddingOptions := []embeddings.Option{}
	embedder, err := embeddings.NewEmbedder(llm)
	check(err)

	session := &Session{
		tracker:        t,
		handler:        &t.handler,
		conn:           &connection{ws: c, send: make(chan model.Envelope)},
		orgId:          orgId,
		workspaceId:    workspaceId,
		conversationId: conversationId,
		manifest:       make(map[string]map[string]model.FileRecord),
		chatbot:        llm,
		embedder:       embedder,
		readErr:        make(chan error, 1),
		writeErr:       make(chan error, 1),
	}

	fmt.Println("websocket connection successful")
	err = session.pump()
	check(err)

}

func (t *Tracker) SetHandler(h Handler) {
	t.handler = h
}

// Broadcast send a message to all connected clients. This method is concurrently safe
// and doesn't do anything if the Hub's context is canceled.

func (t *Tracker) Broadcast(res model.Envelope) {
	select {
	case <-t.ctx.Done(): // Don't send if the hub is shutting down
	case t.broadcast <- res: // move to session
	}
}
