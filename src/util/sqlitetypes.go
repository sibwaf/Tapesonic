package util

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type TimestampWrapper struct {
	value time.Time
}

func NewTimestampWrapperOrNull(value *time.Time) *TimestampWrapper {
	if value == nil || value.IsZero() {
		return nil
	}

	return &TimestampWrapper{value: *value}
}

func NewTimestampWrapper(value time.Time) TimestampWrapper {
	return TimestampWrapper{value: value}
}

func (wrapper TimestampWrapper) Unwrap() time.Time {
	return wrapper.value
}

func (wrapper *TimestampWrapper) UnwrapNullable() *time.Time {
	if wrapper == nil {
		return nil
	}

	return &wrapper.value
}

func (result *TimestampWrapper) Scan(value any) error {
	timestamp, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to read timestamp, not an int64: %v", value)
	}

	*result = TimestampWrapper{
		value: time.UnixMilli(timestamp),
	}

	return nil
}

func (wrapper TimestampWrapper) Value() (driver.Value, error) {
	if wrapper.value.IsZero() {
		return nil, nil
	}

	return wrapper.value.UnixMilli(), nil
}

func (TimestampWrapper) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "BIGINT"
}
