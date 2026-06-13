package helper

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func RpcError(code connect.Code, message string) error {
	return connect.NewError(code, errors.New(message))
}

func ParsedStringToUUID(s string) (pgtype.UUID, error) {
	parsedString, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, RpcError(connect.CodeInvalidArgument, "Cannot convert to uuid")
	}
	return pgtype.UUID{
		Bytes: parsedString,
		Valid: true,
	}, nil
}
