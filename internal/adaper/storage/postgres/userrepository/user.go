package userrepository

import (
	"database/sql"
	"errors"
	"github.com/mohsenabedy91/Sikabiz/internal/adaper/storage/postgres"
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

func (r *UserRepository) GetByID(id uint64) (*domain.User, error) {
	row := r.tx.QueryRow(
		"SELECT id, name, email, phone_number FROM users WHERE deleted_at IS NULL AND id = $1",
		id,
	)
	user, err := scanUser(row)
	if err != nil {
		metrics.DbCall.WithLabelValues("users", "GetByUUID", "Failed").Inc()

		r.log.Error(logger.Database, logger.DatabaseSelect, err.Error(), nil)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, serviceerror.New(serviceerror.RecordNotFound)
		}
		return nil, serviceerror.NewServerError()
	}

	metrics.DbCall.WithLabelValues("users", "GetByUUID", "Success").Inc()

	return &user, nil
}

func scanUser(scanner postgres.Scanner) (domain.User, error) {
	var user domain.User
	var name sql.NullString

	if err := scanner.Scan(
		&user.Base.ID,
		&name,
		&user.Email,
		&user.PhoneNumber,
	); err != nil {
		return domain.User{}, err
	}

	return user, nil
}
