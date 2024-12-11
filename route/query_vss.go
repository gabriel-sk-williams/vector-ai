package route

import (
	bg "context"
	"fmt"
	"vector-ai/model"
	"vector-ai/util"

	pb "github.com/qdrant/go-client/qdrant"
)

//
// Query
//

func (s Session) QueryVss(m model.WebSocketsMessage) {
	// fmt.Println(m)

	workspaceId := m.WorkspaceID
	conversationId := m.ConversationID
	// templateId := m.TemplateID.String
	// authorName := m.AuthorName
	vssText := m.VssText
	// timestamp := time.Now().Format(time.RFC3339)

	// var context string
	var results []*pb.ScoredPoint
	var groupsResult *pb.GroupsResult

	ctx := bg.Background()

	// Vectorizing query body...
	floats, err := s.embedder.EmbedQuery(ctx, vssText)
	check(err)

	_, err = s.handler.QD.GetCollection(s.orgId)
	if err == nil {

		// Get configs and create a struct from them
		configs, err := s.handler.PG.ListWorkspaceConfigs(workspaceId)
		check(err)

		options := util.MarshalVssOptions(configs)

		// perform vss query
		groupsResult, err = s.handler.QD.Vss(floats, s.orgId, workspaceId, options) // [0.1, 0.6, 0.5...]
		check(err)

		pointGroups := groupsResult.GetGroups()

		chunksByDocId := map[string]*model.ChunkValues{}

		for _, pg := range pointGroups {
			results = pg.GetHits()

			// collect vss results into context for the AI
			for _, point := range results {
				pointId := point.GetId().GetUuid()
				payloadMap := point.GetPayload()
				score := point.GetScore()
				payloadDocId := payloadMap["documentId"]
				documentId := payloadDocId.GetStringValue()
				payloadChunk := payloadMap["chunk"]
				val := payloadChunk.GetStringValue()

				scoredChunk := model.ScoredChunk{ID: pointId, Value: val, Score: score}

				// fmt.Println("scores", payloadMap, score, documentId)

				_, keyExists := chunksByDocId[documentId]
				if !keyExists { // create key
					chunksByDocId[documentId] = &model.ChunkValues{}
				}

				chunksByDocId[documentId].AddChunk(scoredChunk)

				if err != nil {
					fmt.Printf("loader error: %s", err)
				}
			}
		}

		var ch model.ContextHolder
		for docId, cv := range chunksByDocId {
			// create a ContextLoader and add to ContextHolder
			document, _ := s.handler.PG.GetDocument(docId)
			ch.AddLoader(model.ContextLoaderScored{DocumentId: docId, DocumentName: document.Name, Context: cv})
		}
		ch.Query = vssText

		// broadcast json to frontend
		s.tracker.Broadcast(model.VssResponse(ch, workspaceId, conversationId))
	}
}
