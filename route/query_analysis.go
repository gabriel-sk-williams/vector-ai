package route

import (
	bg "context"
	"encoding/json"
	"fmt"
	"time"
	"vector-ai/model"
	"vector-ai/util"

	pb "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

const VssFromPromptTemplate = `
take the following LLM prompt:

"{{.prompt}}"

And Create a vector similarity search query from it, only extracting terms the user might be searching for.

Focus on search terms and always ignore terms that are used as directions to an LLM.

For your output, only provide the most revelant search terms, space-separated, with no other conversational verbiage of any kind. This response is intended to be used directly in a vector similarity search.
`

const AIInstructionsBasePrompt = `
{{.instructions}}

The analysis should be done on a per-document basis, and should relate back to the proper document ID, which is provided in the context

Do not use any outside or real-world information. Please format your answer in raw YAML and only raw YAML, with no outside conversational text. Your response will be directly parsed.

The YAML Schema should be as such:
{{.responseSchema}}

Relevant pieces of information:
{{.context}}

If any of this information is irrelevant, it can be discarded.

Current conversation:
Human: {{.query}}
Raw AI YAML Response:
`

func (s Session) QueryAnalysis(m model.WebSocketsMessage) model.Message {

	workspaceId := m.WorkspaceID
	conversationId := m.ConversationID
	templateId := m.TemplateID.String
	authorName := m.AuthorName
	query := m.QueryText
	forceContext := m.ForceContext
	responseSchema := m.ResponseSchema
	timestamp := time.Now().Format(time.RFC3339)

	var context string
	var results []*pb.ScoredPoint
	var groupsResult *pb.GroupsResult
	ctx := bg.Background()

	// Get conversation history for use in s.constructHistory (if constructHistory {...
	// conversationHistory, err := s.handler.PG.ListMessages(s.conversationId)
	// check(err
	// ...}

	// Save message to postgres
	pgMessage, err := s.handler.PG.CreateMessage(workspaceId, conversationId, templateId, query, "Human", authorName, timestamp)
	check(err)

	s.tracker.Broadcast(model.UserResponse(pgMessage, workspaceId, conversationId))

	if len(forceContext) > 0 {
		context = forceContext
	} else {

		s.tracker.Broadcast(model.QueryStatus("Constructing VSS with AI...", workspaceId, conversationId))
		prompt := prompts.NewPromptTemplate(VssFromPromptTemplate, []string{"prompt"})
		constructedPrompt, err := prompt.Format(map[string]any{
			"prompt": query,
		})
		check(err)

		//chatQuery := s.constructSingle(constructedPrompt)

		// Call AI including chat history
		completion, err := s.chatbot.Call(ctx, constructedPrompt) // , llms.WithFunctions(functions))
		check(err)

		s.tracker.Broadcast(model.QueryStatus("Performing Vector Similarity Search...", workspaceId, conversationId))
		// Vectorizing query body...
		floats, err := s.embedder.EmbedQuery(ctx, completion)
		check(err)

		_, err = s.handler.QD.GetCollection(s.orgId)
		if err == nil {

			// Get configs and Create a struct from them
			configs, err := s.handler.PG.ListWorkspaceConfigs(workspaceId)
			check(err)

			options := util.MarshalVssOptions(configs)

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

					// fmt.Println("chunk", val)

					scoredChunk := model.ScoredChunk{ID: pointId, Value: val, Score: score}

					// fmt.Println("scores", payloadMap, score, documentId)

					_, keyExists := chunksByDocId[documentId]
					if !keyExists { // Create key
						chunksByDocId[documentId] = &model.ChunkValues{}
					}

					chunksByDocId[documentId].AddChunk(scoredChunk)

					if err != nil {
						fmt.Printf("loader error: %s", err)
					}
				}
			}

			var ch model.ContextHolder
			// Create a ContextLoader and add to ContextHolder
			for docId, cv := range chunksByDocId {
				document, _ := s.handler.PG.GetDocument(docId)
				ch.AddLoader(model.ContextLoaderScored{DocumentId: docId, DocumentName: document.Name, Context: cv})
			}
			ch.Query = completion
			s.tracker.Broadcast(model.VssResponse(ch, workspaceId, conversationId))

			// convert to json
			jsonBytes, err := json.Marshal(ch)
			check(err)
			context = string(jsonBytes)
		}
	}

	s.tracker.Broadcast(model.QueryStatus("Building prompt...", workspaceId, conversationId))

	// Building prompt...
	template, err := s.handler.PG.GetTemplate(templateId)
	check(err)

	prompt := prompts.NewPromptTemplate(AIInstructionsBasePrompt, []string{"context", "query"})
	constructedPrompt, err := prompt.Format(map[string]any{
		"instructions":   template.Text,
		"context":        context,
		"query":          query,
		"responseSchema": responseSchema,
	})
	check(err)

	// fmt.Println(constructedPrompt)

	s.tracker.Broadcast(model.QueryStatus("Querying AI...", workspaceId, conversationId))

	// chatHistory := s.construct(conversationHistory, constructedPrompt)
	//chatQuery := s.constructSingle(constructedPrompt)

	// Call AI including chat history
	//completion, err := s.chatbot.Call(ctx, chatQuery) // , llms.WithFunctions(functions))

	completion, err := s.chatbot.Call(
		ctx,
		constructedPrompt,
		llms.WithStreamingFunc(func(ctx bg.Context, chunk []byte) error {
			s.tracker.Broadcast(model.AiStreamChunk(chunk, s.workspaceId, s.conversationId))
			return nil
		}),
	)

	check(err)

	reply := completion

	timestamp = time.Now().Format(time.RFC3339) // new timestamp
	message, err := s.handler.PG.CreateAIMessage(workspaceId, conversationId, reply, "AI", "gpt-4-1106-preview", timestamp)
	check(err)

	// set context
	for _, point := range results {
		pointId := point.GetId().GetUuid()
		payloadMap := point.GetPayload()
		payloadDocId := payloadMap["documentId"]
		documentId := payloadDocId.GetStringValue()

		//for documentId := range payloadMap {
		_, err = s.handler.PG.CreateContext(documentId, message.ID, pointId)
		check(err)
		//}
	}

	return message

}
