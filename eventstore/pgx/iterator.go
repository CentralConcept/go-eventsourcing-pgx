package pgx

import (
	"errors"
	"github.com/hallgren/eventsourcing/core"
	"github.com/jackc/pgx/v5"
	"time"
)

type iterator struct {
	rows pgx.Rows
}

// Next return true if there are more data
func (i *iterator) Next() bool {
	return i.rows.Next()
}

func (i *iterator) Value() (core.Event, error) {
	if i.rows == nil {
		return core.Event{}, errors.New("no rows")
	}

	r := i.rows
	var (
		globalVersion, version               core.Version
		id, reason, aggregateType, timestamp string
		data, metadata                       []byte
	)

	if err := r.Scan(&globalVersion, &id, &version, &reason, &aggregateType, &timestamp, &data, &metadata); err != nil {
		return core.Event{}, err
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return core.Event{}, err
	}

	return core.Event{
		AggregateID:   id,
		Version:       version,
		GlobalVersion: globalVersion,
		AggregateType: aggregateType,
		Timestamp:     t,
		Data:          data,
		Metadata:      metadata,
		Reason:        reason,
	}, nil
}

func (i *iterator) Close() {
	i.rows.Close()
}
