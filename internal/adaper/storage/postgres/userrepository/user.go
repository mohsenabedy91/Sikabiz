package userrepository

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
	"github.com/mohsenabedy91/Sikabiz/pkg/metrics"
	"github.com/mohsenabedy91/Sikabiz/pkg/serviceerror"
)

type UserRepository struct {
	log logger.Logger
	tx  *sql.Tx
}

func NewUserRepository(log logger.Logger, tx *sql.Tx) *UserRepository {
	return &UserRepository{
		log: log,
		tx:  tx,
	}
}

func (r *UserRepository) GetByID(uuid uuid.UUID) (*domain.User, error) {
	rows, err := r.tx.Query(
		`SELECT u.uuid, u.first_name, u.last_name, u.email, u.phone_number, a.street, a.city, a.state, a.zip_code, a.country FROM users AS u 
				INNER JOIN addresses as a on u.id = a.user_id AND a.deleted_at IS NULL
               	WHERE u.deleted_at IS NULL AND u.uuid = $1 FOR UPDATE SKIP LOCKED`,
		uuid,
	)
	if err != nil {
		metrics.DbCall.WithLabelValues("users", "GetByID", "Failed").Inc()

		r.log.Error(logger.Database, logger.DatabaseSelect, err.Error(), nil)
		return nil, serviceerror.NewServerError()
	}

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			r.log.Error(logger.Database, logger.DatabaseSelect, err.Error(), nil)
		}
	}(rows)

	var user domain.User
	var addresses []*domain.Address

	for rows.Next() {
		var address domain.Address
		if err = rows.Scan(
			&user.Base.UUID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.PhoneNumber,
			&address.Street,
			&address.City,
			&address.State,
			&address.ZipCode,
			&address.Country,
		); err != nil {
			metrics.DbCall.WithLabelValues("users", "GetByID", "Failed").Inc()

			r.log.Error(logger.Database, logger.DatabaseSelect, err.Error(), nil)
			return nil, serviceerror.NewServerError()
		}

		addresses = append(addresses, &address)
	}

	if err = rows.Err(); err != nil {
		metrics.DbCall.WithLabelValues("users", "GetByID", "Failed").Inc()

		r.log.Error(logger.Database, logger.DatabaseSelect, err.Error(), nil)
		return nil, serviceerror.NewServerError()
	}

	metrics.DbCall.WithLabelValues("users", "GetByUUID", "Success").Inc()

	user.Addresses = addresses
	return &user, nil
}

func (r *UserRepository) Save(user *domain.User) (uint64, error) {
	var userID uint64
	err := r.tx.QueryRow(
		`INSERT INTO users (first_name, last_name, email, phone_number) 
				VALUES ($1, $2, $3, $4) 
				RETURNING id`,
		user.FirstName,
		user.LastName,
		user.Email,
		user.PhoneNumber,
	).Scan(&userID)
	if err != nil {
		metrics.DbCall.WithLabelValues("users", "Save", "Failed").Inc()

		r.log.Error(logger.Database, logger.DatabaseInsert, err.Error(), map[logger.ExtraKey]interface{}{
			logger.InsertDBArg: user,
		})
		return userID, serviceerror.NewServerError()
	}

	metrics.DbCall.WithLabelValues("users", "Save", "Success").Inc()
	return userID, nil
}
