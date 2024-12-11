package qdrant

import (
	"vector-ai/util"

	pb "github.com/qdrant/go-client/qdrant"
)

func (qdr Qdr) Query(vector []float32, orgId string, workspaceId string) ([]*pb.ScoredPoint, error) {

	ctx, cancel := util.GetContext()
	defer cancel()

	pointsClient := pb.NewPointsClient(qdr.Connection)

	// Unfiltered search
	unfilteredSearchResult, err := pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: orgId,
		Vector:         vector,
		Limit:          10,
		// Include payload and/or vectors in search result
		WithVectors: &pb.WithVectorsSelector{SelectorOptions: &pb.WithVectorsSelector_Enable{Enable: false}},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
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
	})

	return unfilteredSearchResult.GetResult(), err
}
