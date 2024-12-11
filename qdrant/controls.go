package qdrant

import (
	"fmt"
	"vector-ai/model"

	"github.com/go-errors/errors"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

type Qdr struct {
	Driver     pb.CollectionsClient
	Connection *grpc.ClientConn
}

type Controls interface {
	GetStatus() (string, error)
	ListCollections() ([]string, error)
	GetCollection(string) (*pb.CollectionInfo, error) // GetQdrantCollection
	CreateCollection(string, uint64) error
	DeleteVectorsByDocumentId(string, string, string) (uint64, error)
	DeleteVectorsByWorkspaceId(string, string) (uint64, error)
	ClearCollection(string) (int, error)
	DeleteCollection(string) error
	GetPoint(string, string) (*pb.RetrievedPoint, error)
	GetPointsByUuid(string, []string) ([]*pb.RetrievedPoint, error)
	GetPointsByIndeces(string, string, []int64) ([]*pb.RetrievedPoint, error)

	// main functions
	Vss([]float32, string, string, model.VssOptions) (*pb.GroupsResult, error)
	Query([]float32, string, string) ([]*pb.ScoredPoint, error)
	Upload(string, string, string, [][]float32, []string) (string, error)

	// helper functions
	GetPointCount(string) (uint32, error) // not in use
}

func check(err error) error {
	if err != nil {
		x := errors.New(err)
		fmt.Println(x.ErrorStack())
		return err
	}
	return nil
}
