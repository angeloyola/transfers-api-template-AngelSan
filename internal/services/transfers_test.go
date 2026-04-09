package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"
)

type transfersRepositoryStub struct {
	createFn  func(ctx context.Context, transfer models.Transfer) (string, error)
	getByIDFn func(ctx context.Context, id string) (models.Transfer, error)
	updateFn  func(ctx context.Context, transfer models.Transfer) error
	deleteFn  func(ctx context.Context, id string) error

	createCalls []models.Transfer
	getCalls    []string
	updateCalls []models.Transfer
	deleteCalls []string
}

func (s *transfersRepositoryStub) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	s.createCalls = append(s.createCalls, transfer)
	if s.createFn != nil {
		return s.createFn(ctx, transfer)
	}
	return transfer.ID, nil
}

func (s *transfersRepositoryStub) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	s.getCalls = append(s.getCalls, id)
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return models.Transfer{}, nil
}

func (s *transfersRepositoryStub) Update(ctx context.Context, transfer models.Transfer) error {
	s.updateCalls = append(s.updateCalls, transfer)
	if s.updateFn != nil {
		return s.updateFn(ctx, transfer)
	}
	return nil
}

func (s *transfersRepositoryStub) Delete(ctx context.Context, id string) error {
	s.deleteCalls = append(s.deleteCalls, id)
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func TestTransfersServiceCreate_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		transfer models.Transfer
		contains string
	}{
		{
			name:     "missing sender",
			transfer: validTransfer(),
			contains: "sender_id is required",
		},
		{
			name: "missing receiver",
			transfer: func() models.Transfer {
				transfer := validTransfer()
				transfer.ReceiverID = ""
				return transfer
			}(),
			contains: "required",
		},
		{
			name: "unknown currency",
			transfer: func() models.Transfer {
				transfer := validTransfer()
				transfer.Currency = enums.CurrencyUnknown
				return transfer
			}(),
			contains: "invalid currency",
		},
		{
			name: "non positive amount",
			transfer: func() models.Transfer {
				transfer := validTransfer()
				transfer.Amount = 0
				return transfer
			}(),
			contains: "amount should be greater than 0",
		},
		{
			name: "missing state",
			transfer: func() models.Transfer {
				transfer := validTransfer()
				transfer.State = "  "
				return transfer
			}(),
			contains: "state is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &transfersRepositoryStub{}
			service := NewTransfersService(config.BusinessConfig{}, repo, nil)

			if tt.name == "missing sender" {
				tt.transfer.SenderID = ""
			}

			_, err := service.Create(context.Background(), tt.transfer)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, known_errors.ErrBadRequest) {
				t.Fatalf("expected bad request error, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("expected error to contain %q, got %q", tt.contains, err.Error())
			}
			if len(repo.createCalls) != 0 {
				t.Fatalf("expected repository not to be called, got %d calls", len(repo.createCalls))
			}
		})
	}
}

func TestTransfersServiceCreate_SavesRepoAndCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	transfer := validTransfer()
	repo := &transfersRepositoryStub{
		createFn: func(_ context.Context, transfer models.Transfer) (string, error) {
			if transfer.ID != "" {
				t.Fatalf("expected repository create input without ID, got %q", transfer.ID)
			}
			return "trf-123", nil
		},
	}
	cache := &transfersRepositoryStub{}

	service := NewTransfersService(config.BusinessConfig{}, repo, cache)

	id, err := service.Create(ctx, transfer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "trf-123" {
		t.Fatalf("expected ID trf-123, got %q", id)
	}
	if len(repo.createCalls) != 1 {
		t.Fatalf("expected one repository call, got %d", len(repo.createCalls))
	}
	if len(cache.createCalls) != 1 {
		t.Fatalf("expected one cache call, got %d", len(cache.createCalls))
	}
	if cache.createCalls[0].ID != "trf-123" {
		t.Fatalf("expected cached transfer ID trf-123, got %q", cache.createCalls[0].ID)
	}
}

func TestTransfersServiceCreate_ReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repository unavailable")
	repo := &transfersRepositoryStub{
		createFn: func(_ context.Context, _ models.Transfer) (string, error) {
			return "", repoErr
		},
	}
	cache := &transfersRepositoryStub{}
	service := NewTransfersService(config.BusinessConfig{}, repo, cache)

	_, err := service.Create(context.Background(), validTransfer())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
	if len(cache.createCalls) != 0 {
		t.Fatalf("expected cache not to be called, got %d calls", len(cache.createCalls))
	}
}

func TestTransfersServiceGetByID_ReturnsCachedValue(t *testing.T) {
	t.Parallel()

	cachedTransfer := validTransfer()
	cachedTransfer.ID = "cached-1"
	cache := &transfersRepositoryStub{
		getByIDFn: func(_ context.Context, id string) (models.Transfer, error) {
			if id != cachedTransfer.ID {
				t.Fatalf("expected ID %q, got %q", cachedTransfer.ID, id)
			}
			return cachedTransfer, nil
		},
	}
	repo := &transfersRepositoryStub{}

	service := NewTransfersService(config.BusinessConfig{}, repo, cache)

	transfer, err := service.GetByID(context.Background(), cachedTransfer.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transfer != cachedTransfer {
		t.Fatalf("expected cached transfer %+v, got %+v", cachedTransfer, transfer)
	}
	if len(repo.getCalls) != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", len(repo.getCalls))
	}
}

func TestTransfersServiceGetByID_FallsBackToRepositoryAndWarmsCache(t *testing.T) {
	t.Parallel()

	repoTransfer := validTransfer()
	repoTransfer.ID = "repo-1"
	cache := &transfersRepositoryStub{
		getByIDFn: func(_ context.Context, _ string) (models.Transfer, error) {
			return models.Transfer{}, errors.New("cache miss")
		},
	}
	repo := &transfersRepositoryStub{
		getByIDFn: func(_ context.Context, id string) (models.Transfer, error) {
			if id != repoTransfer.ID {
				t.Fatalf("expected ID %q, got %q", repoTransfer.ID, id)
			}
			return repoTransfer, nil
		},
	}

	service := NewTransfersService(config.BusinessConfig{}, repo, cache)

	transfer, err := service.GetByID(context.Background(), repoTransfer.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transfer != repoTransfer {
		t.Fatalf("expected repository transfer %+v, got %+v", repoTransfer, transfer)
	}
	if len(repo.getCalls) != 1 {
		t.Fatalf("expected one repository call, got %d", len(repo.getCalls))
	}
	if len(cache.createCalls) != 1 {
		t.Fatalf("expected cache warm-up call, got %d", len(cache.createCalls))
	}
	if cache.createCalls[0] != repoTransfer {
		t.Fatalf("expected cached transfer %+v, got %+v", repoTransfer, cache.createCalls[0])
	}
}

func TestTransfersServiceGetByID_ReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("not found in repository")
	repo := &transfersRepositoryStub{
		getByIDFn: func(_ context.Context, _ string) (models.Transfer, error) {
			return models.Transfer{}, repoErr
		},
	}
	cache := &transfersRepositoryStub{
		getByIDFn: func(_ context.Context, _ string) (models.Transfer, error) {
			return models.Transfer{}, errors.New("cache miss")
		},
	}

	service := NewTransfersService(config.BusinessConfig{}, repo, cache)

	_, err := service.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestTransfersServiceUpdate_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		transfer models.Transfer
		contains string
	}{
		{
			name:     "missing id",
			transfer: validTransfer(),
			contains: "ID is required",
		},
		{
			name: "no fields to update",
			transfer: models.Transfer{
				ID:       "trf-1",
				Currency: enums.CurrencyUnknown,
			},
			contains: "no fields to update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &transfersRepositoryStub{}
			service := NewTransfersService(config.BusinessConfig{}, repo, nil)

			if tt.name == "missing id" {
				tt.transfer.ID = ""
			}

			err := service.Update(context.Background(), tt.transfer)
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, known_errors.ErrBadRequest) {
				t.Fatalf("expected bad request error, got %v", err)
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("expected error to contain %q, got %q", tt.contains, err.Error())
			}
			if len(repo.updateCalls) != 0 {
				t.Fatalf("expected repository not to be called, got %d calls", len(repo.updateCalls))
			}
		})
	}
}

func TestTransfersServiceUpdate_CallsRepository(t *testing.T) {
	t.Parallel()

	transfer := models.Transfer{
		ID:    "trf-1",
		State: "completed",
	}
	repo := &transfersRepositoryStub{}
	service := NewTransfersService(config.BusinessConfig{}, repo, nil)

	err := service.Update(context.Background(), transfer)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.updateCalls) != 1 {
		t.Fatalf("expected one repository update call, got %d", len(repo.updateCalls))
	}
	if repo.updateCalls[0] != transfer {
		t.Fatalf("expected update payload %+v, got %+v", transfer, repo.updateCalls[0])
	}
}

func TestTransfersServiceUpdate_ReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("update failed")
	repo := &transfersRepositoryStub{
		updateFn: func(_ context.Context, _ models.Transfer) error {
			return repoErr
		},
	}
	service := NewTransfersService(config.BusinessConfig{}, repo, nil)

	err := service.Update(context.Background(), models.Transfer{ID: "trf-1", State: "completed"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func TestTransfersServiceDelete_CallsRepository(t *testing.T) {
	t.Parallel()

	repo := &transfersRepositoryStub{}
	service := NewTransfersService(config.BusinessConfig{}, repo, nil)

	err := service.Delete(context.Background(), "trf-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.deleteCalls) != 1 {
		t.Fatalf("expected one repository delete call, got %d", len(repo.deleteCalls))
	}
	if repo.deleteCalls[0] != "trf-1" {
		t.Fatalf("expected deleted ID trf-1, got %q", repo.deleteCalls[0])
	}
}

func TestTransfersServiceDelete_ReturnsRepositoryError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("delete failed")
	repo := &transfersRepositoryStub{
		deleteFn: func(_ context.Context, _ string) error {
			return repoErr
		},
	}
	service := NewTransfersService(config.BusinessConfig{}, repo, nil)

	err := service.Delete(context.Background(), "trf-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func validTransfer() models.Transfer {
	return models.Transfer{
		SenderID:   "sender-1",
		ReceiverID: "receiver-1",
		Currency:   enums.CurrencyUSD,
		Amount:     100,
		State:      "pending",
	}
}
