package qdrant

import (
	"vector-ai/model"
	"vector-ai/util"

	pb "github.com/qdrant/go-client/qdrant"
)

// TODO: receive documentIds
func (qdr Qdr) Vss(vector []float32, orgId string, workspaceId string, options model.VssOptions) (*pb.GroupsResult, error) {
	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	pointsClient := pb.NewPointsClient(qdr.Connection)

	pointGroups, err := pointsClient.SearchGroups(ctx, &pb.SearchPointGroups{
		CollectionName: orgId,
		Vector:         vector,
		Limit:          options.VssDocumentLimit,

		WithPayload: &pb.WithPayloadSelector{
			SelectorOptions: &pb.WithPayloadSelector_Enable{
				Enable: true,
			},
		},
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "workspaceId",
							Match: &pb.Match{
								MatchValue: &pb.Match_Text{
									Text: workspaceId,
								},
							},
						},
					},
				},
			},
		},
		// WithPayload     *WithPayloadSelector
		// Params          *SearchParams
		// ScoreThreshold  *float32
		// VectorName      *string
		// WithVectors     *WithVectorsSelector
		GroupBy:   "documentId", // documentId
		GroupSize: options.VssChunkLimit,
		// ReadConsistency *ReadConsistency
		// WithLookup      *WithLookup
	})

	return pointGroups.GetResult(), err
}
