package qdrant

import (
	"fmt"

	"vector-ai/constants"
	"vector-ai/util"

	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
)

func (qdr Qdr) Upload(orgId string, workspaceId string, documentId string, floats [][]float32, chunks []string) (string, error) {

	points := []*pb.PointStruct{}

	// Upload points
	for i, vector := range floats {
		chunk := chunks[i]
		uuid := uuid.New()
		pointId := uuid.String()
		//fmt.Println(pointId, chunk)

		point := pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: pointId},
			},
			Vectors: &pb.Vectors{VectorsOptions: &pb.Vectors_Vector{Vector: &pb.Vector{Data: vector}}},
			Payload: map[string]*pb.Value{
				"workspaceId": {
					Kind: &pb.Value_StringValue{StringValue: workspaceId},
				},
				"documentId": {
					Kind: &pb.Value_StringValue{StringValue: documentId},
				},
				"tags": {
					Kind: &pb.Value_ListValue{},
				},
				"chunk": {
					Kind: &pb.Value_StringValue{StringValue: chunk},
				},
				"index": {
					Kind: &pb.Value_IntegerValue{IntegerValue: int64(i)},
				},
				"embedder": {
					Kind: &pb.Value_StringValue{StringValue: constants.Embedder},
				},
			},
		}

		points = append(points, &point)
	}

	// Create points grpc client
	pointsClient := pb.NewPointsClient(qdr.Connection)

	ctx, cancel := util.GetContextWithDuration(len(points))
	defer cancel()

	waitUpsert := true

	_, err := pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: orgId,
		Wait:           &waitUpsert,
		Points:         points,
	})

	// updateResult := operationResponse.GetResult()
	// updateStatus := updateResult.GetStatus()
	// status := updateStatus.String() // completely useless

	if err == nil {
		upsert := fmt.Sprintf("Upserted %d points \n", len(points))
		return upsert, err
	} else {
		failure := fmt.Sprintf("Could not upsert points: %v", err)
		return failure, err
	}
}
