package transaction

import (
	"database/sql"
	"errors"
)

func IsNotFoundErr(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
