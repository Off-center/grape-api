package services

import (
	"context"
	"fmt"
	"grape-api/internal/domain/instance"
)

type BootstrapService struct {
	instanceRepo instance.IInstanceRepository
}

func NewBootstrapService(instanceRepo instance.IInstanceRepository) *BootstrapService {
	return &BootstrapService{
		instanceRepo: instanceRepo,
	}
}

func (s *BootstrapService) Initialize(ctx context.Context) error {
	exists, err := s.instanceRepo.Exists(ctx)
	if err != nil {
		return fmt.Errorf("check instance existence: %w", err)
	}
	if exists {
		return nil
	}

	data := instance.NewUninitializedInstance()
	if err := s.instanceRepo.CreateIfAbsent(ctx, &data); err != nil {
		return fmt.Errorf("create uninitialized instance: %w", err)
	}

	return nil
}
