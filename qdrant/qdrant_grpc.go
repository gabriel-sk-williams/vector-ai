package qdrant

import (
	"cmp"
	"errors"
	"fmt"
	"log"
	"slices"

	"vector-ai/util"

	pb "github.com/qdrant/go-client/qdrant"
)

// not in use
func (qdr Qdr) GetStatus() (string, error) {

	ctx, cancel := util.GetContext()
	defer cancel()

	// Check Qdrant version
	qdrantClient := pb.NewQdrantClient(qdr.Connection)
	healthCheckResult, err := qdrantClient.HealthCheck(ctx, &pb.HealthCheckRequest{})

	if err != nil {
		log.Fatalf("Could not Get health: %v", err)
	}

	return healthCheckResult.GetVersion(), err
}

func (qdr Qdr) ListCollections() ([]string, error) {

	ctx, cancel := util.GetContext()
	defer cancel()

	r, err := qdr.Driver.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		log.Fatalf("could not Get collections: %v", err)
	}

	descriptions := r.GetCollections()

	collections := make([]string, len(descriptions))
	for i, v := range descriptions {
		collections[i] = v.GetName()
	}

	return collections, err
}

func (qdr Qdr) GetCollection(qId string) (*pb.CollectionInfo, error) {
	ctx, cancel := util.GetContext()
	defer cancel()

	r, err := qdr.Driver.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: qId,
	})
	if err != nil {
		return nil, err
	}

	info := r.GetResult()

	return info, err
}

// used by upload_drive and upload_manual
func (qdr Qdr) CreateCollection(orgId string, vectorSize uint64) error {

	ctx, cancel := util.GetContext()
	defer cancel()

	// Create new collection
	var defaultSegmentNumber uint64 = 2
	_, err := qdr.Driver.Create(ctx, &pb.CreateCollection{
		CollectionName: orgId,
		VectorsConfig: &pb.VectorsConfig{Config: &pb.VectorsConfig_Params{
			Params: &pb.VectorParams{
				Size:     vectorSize,
				Distance: pb.Distance_Dot,
			},
		}},
		OptimizersConfig: &pb.OptimizersConfigDiff{
			DefaultSegmentNumber: &defaultSegmentNumber,
		},
	})

	if err != nil {
		log.Fatalf("\nCould not create collection: %v", err)
	} else {
		log.Println("\nCollection", orgId, "created")
	}

	return fmt.Errorf("create error")
}

func (qdr Qdr) DeleteVectorsByWorkspaceId(orgId string, workspaceId string) (uint64, error) {
	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	fmt.Println("DeleteVectorsByWorkspaceId", orgId, workspaceId)

	pointsClient := pb.NewPointsClient(qdr.Connection)

	countResponse, err := pointsClient.Count(ctx, &pb.CountPoints{
		CollectionName: orgId,
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
	check(err)

	countResult := countResponse.GetResult()
	pointCount := countResult.GetCount()

	waitDelete := true

	Delete, err := pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: orgId,
		Wait:           &waitDelete,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
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
			},
		},
	})

	if err != nil {
		log.Fatalf("Could not clear points: %v", err)
	} else {
		log.Println(Delete.GetResult())
		log.Println("Deleted", pointCount, "points")
	}

	return pointCount, err

}

func (qdr Qdr) DeleteVectorsByDocumentId(orgId string, workspaceId string, documentId string) (uint64, error) {
	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	pointsClient := pb.NewPointsClient(qdr.Connection)

	countResponse, err := pointsClient.Count(ctx, &pb.CountPoints{
		CollectionName: orgId,
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
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "documentId",
							Match: &pb.Match{
								MatchValue: &pb.Match_Text{
									Text: documentId,
								},
							},
						},
					},
				},
			},
		},
	})
	check(err)

	countResult := countResponse.GetResult()
	pointCount := countResult.GetCount()

	waitDelete := true

	Delete, err := pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: orgId,
		Wait:           &waitDelete,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
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
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "documentId",
									Match: &pb.Match{
										MatchValue: &pb.Match_Text{
											Text: documentId,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	if err != nil {
		log.Fatalf("Could not clear points: %v", err)
	} else {
		log.Println(Delete.GetResult())
		log.Println("Deleted", pointCount, "points")
	}

	return pointCount, err

}

func (qdr Qdr) ClearCollection(orgId string) (int, error) {
	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	// Get collection info -> number of points in collection
	r, err := qdr.Driver.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: orgId,
	})
	check(err)

	info := r.GetResult()
	pointCount := uint32(info.PointsCount)

	pointsClient := pb.NewPointsClient(qdr.Connection)

	if pointCount < 1 {
		return 0, errors.New("collection has no points")
	}

	// scroll collection to Get all ids
	scroll, err := pointsClient.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: orgId,
		Limit:          &pointCount,
	})
	check(err)

	scrollResult := scroll.GetResult()

	// create slice of ids from result
	ids := []*pb.PointId{}

	for _, nb := range scrollResult {
		vim := nb.GetPayload()
		for index := range vim {
			fmt.Println(index)
		}

		id := nb.GetId()
		ids = append(ids, id)
	}

	waitDelete := true

	// Delete points using list of ids
	response, err := pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: orgId,
		Wait:           &waitDelete,
		Points:         &pb.PointsSelector{PointsSelectorOneOf: &pb.PointsSelector_Points{Points: &pb.PointsIdsList{Ids: ids}}},
	})

	if err != nil {
		log.Fatalf("Could not clear points: %v", err)
	} else {
		log.Println(response.GetResult())
		log.Println("Deleted", len(ids), "points")
	}

	return len(ids), err
}

func (qdr Qdr) DeleteCollection(collectionId string) error {

	ctx, cancel := util.GetContext()
	defer cancel()

	// Delete collection
	_, err := qdr.Driver.Delete(ctx, &pb.DeleteCollection{
		CollectionName: collectionId,
	})

	if err != nil {
		log.Fatalf("Could not Delete collection: %v", err)
	} else {
		log.Println("Collection", collectionId, "Deleted")
	}

	return err
}

func (qdr Qdr) GetPoint(orgId string, pointId string) (*pb.RetrievedPoint, error) {
	ctx, cancel := util.GetContext()
	defer cancel()

	pointsClient := pb.NewPointsClient(qdr.Connection)

	GetResponse, err := pointsClient.Get(ctx, &pb.GetPoints{
		CollectionName: orgId,
		Ids: []*pb.PointId{
			{PointIdOptions: &pb.PointId_Uuid{Uuid: pointId}},
		},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})
	check(err)

	return GetResponse.GetResult()[0], err
}

func (qdr Qdr) GetPointsByUuid(orgId string, pointIds []string) ([]*pb.RetrievedPoint, error) {
	pointsClient := pb.NewPointsClient(qdr.Connection)
	ids := []*pb.PointId{}

	for _, id := range pointIds {
		ids = append(ids, &pb.PointId{
			PointIdOptions: &pb.PointId_Uuid{Uuid: id},
		})
	}

	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	response, err := pointsClient.Get(ctx, &pb.GetPoints{
		CollectionName: orgId,
		Ids:            ids,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})

	check(err)

	return response.GetResult(), err
}

func (qdr Qdr) GetPointsByIndeces(orgId string, documentId string, indeces []int64) ([]*pb.RetrievedPoint, error) {
	ctx, cancel := util.GetContext()
	defer cancel()

	pointsClient := pb.NewPointsClient(qdr.Connection)

	// filter
	searchResponse, err := pointsClient.Scroll(ctx, &pb.ScrollPoints{
		CollectionName: orgId,
		Filter: &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "documentId",
							Match: &pb.Match{
								MatchValue: &pb.Match_Text{
									Text: documentId,
								},
							},
						},
					},
				},
			},
			Should: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "index",
							Match: &pb.Match{
								MatchValue: &pb.Match_Integers{
									Integers: &pb.RepeatedIntegers{
										Integers: indeces,
									},
								},
							},
						},
					},
				},
			},
		},
		WithPayload: &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})

	searchResults := searchResponse.GetResult()

	slices.SortFunc(searchResults,
		func(a *pb.RetrievedPoint, b *pb.RetrievedPoint) int {
			return cmp.Compare(a.GetPayload()["index"].GetIntegerValue(), b.GetPayload()["index"].GetIntegerValue())
		})

	return searchResults, err
}

// currently not in use
func (qdr Qdr) GetPointCount(collectionId string) (uint32, error) {
	ctx, cancel := util.GetContextWithDuration(30)
	defer cancel()

	// Get collection info -> number of points in collection
	r, err := qdr.Driver.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: collectionId,
	})
	check(err)

	info := r.GetResult()
	pointCount := uint32(info.PointsCount)

	return pointCount, err
}
