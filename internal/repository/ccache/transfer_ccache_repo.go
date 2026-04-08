package ccache

import (
	"context"
	"time"

	"transfers-api/internal/enums"
	"transfers-api/internal/models"

	githubccache "github.com/karlseguin/ccache/v2"
)

type TransferCcacheRepo struct {
	cache *githubccache.Cache
	ttl   time.Duration
}

func NewTransferCcacheRepo(ttl time.Duration) *TransferCcacheRepo {
	return &TransferCcacheRepo{
		cache: githubccache.New(githubccache.Configure()),
		ttl:   ttl,
	}
}

func (r *TransferCcacheRepo) Create(_ context.Context, transfer models.Transfer) (string, error) {
	r.cache.Set(transfer.ID, transfer, r.ttl)
	return transfer.ID, nil
}

func (r *TransferCcacheRepo) GetByID(_ context.Context, id string) (models.Transfer, error) {
	item := r.cache.Get(id)
	if item == nil {
		return models.Transfer{}, nil
	}

	transfer, ok := item.Value().(models.Transfer)
	if !ok {
		return models.Transfer{}, nil
	}

	return transfer, nil
}

func (r *TransferCcacheRepo) Update(ctx context.Context, transfer models.Transfer) error {
	existing, err := r.GetByID(ctx, transfer.ID)
	if err != nil {
		return err
	}

	if existing.ID == "" {
		return nil
	}

	if transfer.SenderID != "" {
		existing.SenderID = transfer.SenderID
	}
	if transfer.ReceiverID != "" {
		existing.ReceiverID = transfer.ReceiverID
	}
	if transfer.Currency != enums.CurrencyUnknown {
		existing.Currency = transfer.Currency
	}
	if transfer.Amount != 0 {
		existing.Amount = transfer.Amount
	}
	if transfer.State != "" {
		existing.State = transfer.State
	}

	r.cache.Set(existing.ID, existing, r.ttl)
	return nil
}

func (r *TransferCcacheRepo) Delete(_ context.Context, id string) error {
	r.cache.Delete(id)
	return nil
}
