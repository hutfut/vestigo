package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/shopspring/decimal"
)

// Decimal is a custom GraphQL scalar backed by shopspring/decimal.
type Decimal decimal.Decimal

func (d Decimal) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(decimal.Decimal(d).String()))
}

func (d *Decimal) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case string:
		val, err := decimal.NewFromString(v)
		if err != nil {
			return fmt.Errorf("invalid decimal: %w", err)
		}
		*d = Decimal(val)
		return nil
	case float64:
		*d = Decimal(decimal.NewFromFloat(v))
		return nil
	case int64:
		*d = Decimal(decimal.NewFromInt(v))
		return nil
	case json.Number:
		val, err := decimal.NewFromString(v.String())
		if err != nil {
			return fmt.Errorf("invalid decimal: %w", err)
		}
		*d = Decimal(val)
		return nil
	default:
		return fmt.Errorf("invalid decimal type: %T", v)
	}
}

// Date is a custom GraphQL scalar for date-only values (YYYY-MM-DD).
type Date time.Time

func (d Date) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(time.Time(d).Format("2006-01-02")))
}

func (d *Date) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("date must be a string")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	*d = Date(t)
	return nil
}

// DateTime is a custom GraphQL scalar for timestamps (RFC3339).
type DateTime time.Time

func (d DateTime) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(time.Time(d).Format(time.RFC3339)))
}

func (d *DateTime) UnmarshalGQL(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("datetime must be a string")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("invalid datetime format: %w", err)
	}
	*d = DateTime(t)
	return nil
}

// Ensure our scalars implement the right gqlgen interfaces.
var (
	_ graphql.Marshaler   = Decimal{}
	_ graphql.Unmarshaler = (*Decimal)(nil)
	_ graphql.Marshaler   = Date{}
	_ graphql.Unmarshaler = (*Date)(nil)
	_ graphql.Marshaler   = DateTime{}
	_ graphql.Unmarshaler = (*DateTime)(nil)
)
