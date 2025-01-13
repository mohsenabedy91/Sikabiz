package userrepository

import (
	"database/sql"
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"github.com/mohsenabedy91/Sikabiz/pkg/metrics"
	"github.com/mohsenabedy91/Sikabiz/pkg/serviceerror"
)

type AddressRepository struct {
	log logger.Logger
	tx  *sql.Tx
}

func NewAddressRepository(log logger.Logger, tx *sql.Tx) *AddressRepository {
	return &AddressRepository{
		log: log,
		tx:  tx,
	}
}

func (r *AddressRepository) Save(userID uint64, addresses []*domain.Address) error {
	stmt, err := r.tx.Prepare(`INSERT INTO addresses (street, city, state, zip_code, country, user_id) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		metrics.DbCall.WithLabelValues("addresses", "Save", "Failed").Inc()

		r.log.Error(logger.Database, logger.DatabasePrepare, err.Error(), nil)
		return serviceerror.NewServerError()
	}
	defer func(stmt *sql.Stmt) {
		if err = stmt.Close(); err != nil {
			r.log.Error(logger.Database, logger.DatabaseInsert, err.Error(), nil)
		}
	}(stmt)

	for _, address := range addresses {
		if _, err = stmt.Exec(address.Street, address.City, address.State, address.ZipCode, address.Country, userID); err != nil {
			metrics.DbCall.WithLabelValues("addresses", "Save", "Failed").Inc()

			r.log.Error(logger.Database, logger.DatabaseInsert, err.Error(), map[logger.ExtraKey]interface{}{
				"userID":   userID,
				"street":   address.Street,
				"city":     address.City,
				"state":    address.State,
				"zip_code": address.ZipCode,
				"country":  address.Country,
			})
			return serviceerror.NewServerError()
		}
	}

	metrics.DbCall.WithLabelValues("addresses", "Save", "Success").Inc()

	return nil
}
