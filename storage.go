package main

import (
	"context"

	"cloud.google.com/go/storage"
)

type Storage struct {
	Project string
	SAFile  string
	Active  bool
	Client  *storage.Client
	Bucket  string
}

func NewStorage(ctx context.Context, bucket string) (*Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Client: client,
		Bucket: bucket,
	}, nil
}

func (s *Storage) SaveToBucket(filename string, data []byte) error {
	ctx := context.Background()
	wc := s.Client.Bucket(s.Bucket).Object(filename).NewWriter(ctx)
	defer wc.Close()
	wc.Write(data)
	return nil
}
