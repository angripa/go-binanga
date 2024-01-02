package database

import (
	"binanga/internal/cache"
	"binanga/internal/database"
	"binanga/internal/merchant/model"
	"binanga/pkg/logging"
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IterateMerchantCriteria struct {
	Offset uint
	Limit  uint
}

//go:generate mockery --name MerchantDB --filename merchant_mock.go
type MerchantDB interface {
	RunInTx(ctx context.Context, f func(ctx context.Context) error) error

	// SaveMerchant saves a given merchant with tags.
	SaveMerchant(ctx context.Context, merchant *model.Merchant) error

	// FindMerchantById returns a merchant with given slug
	// database.ErrNotFound error is returned if not exist
	FindMerchantById(ctx context.Context, id string) (*model.Merchant, error)
}

// NewMerchantDB creates a new merchant db with given db
func NewMerchantDB(db *gorm.DB, cacher cache.Cacher) MerchantDB {
	// if cacher == nil {
	// 	return &merchantDB{db: db}
	// }
	// return NewMerchantCacheDB(cacher, &merchantDB{db: db})
	return &merchantDB{db: db}
}

type merchantDB struct {
	db *gorm.DB
}

// FindMerchantById implements MerchantDB.
func (a *merchantDB) FindMerchantById(ctx context.Context, id string) (*model.Merchant, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("article.db.FindMerchantById", "id", id)

	var ret model.Merchant
	idx, _ := strconv.Atoi(id)
	err := db.WithContext(ctx).
		First(&ret, "id = ? AND deleted_at_unix = 0", idx).Error

	if err != nil {
		logger.Errorw("failed to find article", "err", err)
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (a *merchantDB) RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx := a.db.Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "start tx")
	}

	ctx = database.WithDB(ctx, tx)
	if err := f(ctx); err != nil {
		if err1 := tx.Rollback().Error; err1 != nil {
			return errors.Wrap(err, fmt.Sprintf("rollback tx: %v", err1.Error()))
		}
		return errors.Wrap(err, "invoke function")
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}
	return nil
}

func (a *merchantDB) SaveMerchant(ctx context.Context, merchant *model.Merchant) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("merchant.db.SaveMerchant", "merchant", merchant)

	if err := db.WithContext(ctx).Create(merchant).Error; err != nil {
		logger.Errorw("merchant.db.SaveMerchant failed to save merchant", "err", err)
		if database.IsKeyConflictErr(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}
