package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"transfers-api/internal/config"
	"transfers-api/internal/logging"
	"transfers-api/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type TransfersMySQLRepo struct {
	db    *sql.DB
	table string
}

func NewTransfersMySQLRepository(cfg config.MySQL) *TransfersMySQLRepo {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.Username, cfg.Password, cfg.Hostname, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logging.Logger.Fatalf("error creating MySQL connection: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logging.Logger.Fatalf("error connecting to MySQL: %v", err)
	}

	return &TransfersMySQLRepo{
		db:    db,
		table: cfg.Table,
	}
}

func (r *TransfersMySQLRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	return "", fmt.Errorf("mysql repository create not implemented")
}

func (r *TransfersMySQLRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	return models.Transfer{}, fmt.Errorf("mysql repository get by id not implemented")
}

func (r *TransfersMySQLRepo) Update(ctx context.Context, transfer models.Transfer) error {
	return fmt.Errorf("mysql repository update not implemented")
}

func (r *TransfersMySQLRepo) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("mysql repository delete not implemented")
}
