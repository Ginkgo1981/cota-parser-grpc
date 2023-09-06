package server

import (
	"log"

	pb "github.com/ginkgo1981/entries-parser/api/cotaparser/v1"
	"github.com/ginkgo1981/entries-parser/internal/service"
)

type Cota struct {
	pb.UnimplementedCotaServer
}

func (c Cota) GetCotaEntries(in *pb.GetCotaEntriesRequest, stream pb.Cota_GetCotaEntriesServer) error {

	parser := service.NewCotaParser()
	response, err := parser.Parse(in.Witness, in.Version)
	if err != nil {
		log.Printf("parse err: %v", err)
		return err
	}

	if err := stream.Send(response); err != nil {
		log.Printf("send response err: %v", err)
		return err
	}

	return nil
}
