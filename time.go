package meta

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

//
// Time
//

type Time struct {
	Val time.Time
	Nullity
	Presence
	Path string
}

type TimeOptions struct {
	Required     bool
	DiscardBlank bool
	Null         bool
	Format       []string
	MinDate      *dateLimit
	MaxDate      *dateLimit
}

type dateLimit struct {
	value      time.Time
	isAbsolute bool
	exclusive  bool
	raw        string
}

func NewTime(t time.Time) Time {
	return Time{t, Nullity{false}, Presence{true}, ""}
}

func (t *Time) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &TimeOptions{
		Required:     tag.Get("meta_required") == "true",
		DiscardBlank: tag.Get("meta_discard_blank") != "false",
		Null:         tag.Get("meta_null") == "true",
		Format:       []string{time.RFC3339, "expression"},
	}

	if tag.Get("meta_format") != "" {
		formats := strings.Split(tag.Get("meta_format"), ",")
		opts.Format = []string{}
		for _, format := range formats {
			switch strings.TrimSpace(format) {
			case "ANSIC":
				opts.Format = append(opts.Format, time.ANSIC)
			case "UnixDate":
				opts.Format = append(opts.Format, time.UnixDate)
			case "RubyDate":
				opts.Format = append(opts.Format, time.RubyDate)
			case "RFC822":
				opts.Format = append(opts.Format, time.RFC822)
			case "RFC822Z":
				opts.Format = append(opts.Format, time.RFC822Z)
			case "RFC850":
				opts.Format = append(opts.Format, time.RFC850)
			case "RFC1123":
				opts.Format = append(opts.Format, time.RFC1123)
			case "RFC1123Z":
				opts.Format = append(opts.Format, time.RFC1123Z)
			case "RFC3339":
				opts.Format = append(opts.Format, time.RFC3339)
			case "RFC3339Nano":
				opts.Format = append(opts.Format, time.RFC3339Nano)
			case "Kitchen":
				opts.Format = append(opts.Format, time.Kitchen)
			case "Stamp":
				opts.Format = append(opts.Format, time.Stamp)
			case "StampMilli":
				opts.Format = append(opts.Format, time.StampMilli)
			case "StampMicro":
				opts.Format = append(opts.Format, time.StampMicro)
			case "StampNano":
				opts.Format = append(opts.Format, time.StampNano)
			case "DateTime":
				opts.Format = append(opts.Format, time.DateTime)
			case "DateOnly":
				opts.Format = append(opts.Format, time.DateOnly)
			case "TimeOnly":
				opts.Format = append(opts.Format, time.TimeOnly)
			default:
				opts.Format = append(opts.Format, strings.TrimSpace(format))
			}
		}
	}

	opts.MinDate = newDateLimit(tag.Get("meta_min"), opts)
	opts.MaxDate = newDateLimit(tag.Get("meta_max"), opts)

	return opts
}

func (t *Time) JSONValue(path string, i interface{}, options interface{}) Errorable {
	t.Path = path
	if i == nil {
		return t.FormValue("", options)
	}

	switch value := i.(type) {
	case time.Time:
		opts := options.(*TimeOptions)
		if value.IsZero() {
			if opts.Null {
				t.Present = true
				t.Null = true
				return nil
			}
			if opts.Required {
				return ErrBlank
			}
			if !opts.DiscardBlank {
				t.Present = true
				return ErrBlank
			}
			return nil
		}
		t.Present = true
		t.Val = value
		return t.assertTimeRange(opts)
	case string:
		return t.FormValue(value, options)
	}

	return ErrTime
}

func newDateLimit(raw string, opts *TimeOptions) *dateLimit {
	if raw == "" {
		return nil
	}

	// Check for exclusive prefix
	exclusive := false
	value := raw
	if strings.HasPrefix(raw, "!") {
		exclusive = true
		value = strings.TrimPrefix(raw, "!")
	}

	if slices.Contains(opts.Format, "expression") {
		if v := resolveTimeExpression(value); v != nil {
			return &dateLimit{raw: value, exclusive: exclusive}
		}
	}

	t := Time{}
	if err := t.FormValue(value, opts); err != nil {
		return nil
	}

	return &dateLimit{value: t.Val, isAbsolute: true, exclusive: exclusive, raw: value}
}

func (d *dateLimit) Value(opts *TimeOptions) time.Time {
	if d.isAbsolute {
		return d.value
	} else if v := resolveTimeExpression(d.raw); v != nil {
		return *v
	}

	return time.Now()
}

// Allow for some tolerance in time comparisons to handle time expression resolution differences
const TOLERANCE = time.Millisecond * 250 // 1/4 second

func (t *Time) assertTimeRange(opts *TimeOptions) Errorable {
	if opts.MinDate != nil {
		minDate := opts.MinDate.Value(opts)
		if opts.MinDate.exclusive {
			// Exclusive min: value must be strictly after min date (within tolerance)
			if !t.Val.After(minDate.Add(TOLERANCE)) {
				return ErrMin
			}
		} else {
			// Inclusive min: value must be greater than or equal to min date (within tolerance)
			if t.Val.Before(minDate.Add(-TOLERANCE)) {
				return ErrMin
			}
		}
	}
	if opts.MaxDate != nil {
		maxDate := opts.MaxDate.Value(opts)
		if opts.MaxDate.exclusive {
			// Exclusive max: value must be strictly before max date (within tolerance)
			if !t.Val.Before(maxDate.Add(-TOLERANCE)) {
				return ErrMax
			}
		} else {
			// Inclusive max: value must be less than or equal to max date (within tolerance)
			if t.Val.After(maxDate.Add(TOLERANCE)) {
				return ErrMax
			}
		}
	}
	return nil
}

func resolveTimeExpression(value string) *time.Time {
	for _, parser := range timeExpressionParsers {
		submatches := parser.Regexp.FindStringSubmatch(value)
		if len(submatches) == 0 {
			continue
		}
		if v, ok := parser.Parse(submatches); ok {
			return &v
		}
	}
	return nil
}

type expressionParser struct {
	*regexp.Regexp
	Parse func([]string) (time.Time, bool)
}

var timeExpressionParsers = []expressionParser{
	{
		Regexp: regexp.MustCompile(`^(\d+)_(year|month|week|day|hour|minute|second|nanosecond)s?_(ago|from_now)$`),
		Parse: func(matches []string) (time.Time, bool) {
			delta, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, false
			}
			if matches[3] == "ago" {
				delta = -delta
			}
			switch matches[2] {
			case "year":
				return time.Now().AddDate(delta, 0, 0), true
			case "month":
				return time.Now().AddDate(0, delta, 0), true
			case "week":
				return time.Now().AddDate(0, 0, delta*7), true
			case "day":
				return time.Now().AddDate(0, 0, delta), true
			case "hour":
				return time.Now().Add(time.Duration(delta) * time.Hour), true
			case "minute":
				return time.Now().Add(time.Duration(delta) * time.Minute), true
			case "second":
				return time.Now().Add(time.Duration(delta) * time.Second), true
			case "nanosecond":
				return time.Now().Add(time.Duration(delta) * time.Nanosecond), true
			}

			return time.Time{}, false
		},
	},
	{
		Regexp: regexp.MustCompile(`^now$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now(), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^today$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24 * time.Hour), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^yesterday$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24*time.Hour).AddDate(0, 0, -1), true
		},
	},
	{
		Regexp: regexp.MustCompile(`^tomorrow$`),
		Parse: func(matches []string) (time.Time, bool) {
			return time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1), true
		},
	},
}

func (t *Time) FormValue(value string, options interface{}) Errorable {
	opts := options.(*TimeOptions)

	if value == "" {
		if opts.Null {
			t.Present = true
			t.Null = true
			return nil
		}
		if opts.Required {
			return ErrBlank
		}
		if !opts.DiscardBlank {
			t.Present = true
			return ErrBlank
		}
		return nil
	}

	for _, format := range opts.Format {
		switch format {
		case "expression":
			if v := resolveTimeExpression(value); v != nil {
				t.Val = *v
				t.Present = true
				return t.assertTimeRange(opts)
			}
		default:
			if v, err := time.Parse(format, value); err == nil {
				t.Val = v
				t.Present = true
				return t.assertTimeRange(opts)
			}
		}
	}

	return ErrTime
}

func (t Time) Value() (driver.Value, error) {
	if t.Present && !t.Null {
		return t.Val, nil
	}
	return nil, nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.Present && !t.Null {
		return MetaJson.Marshal(t.Val)
	}
	return nullString, nil
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if bytes.Equal(nullString, b) {
		t.Nullity = Nullity{true}
		return nil
	}
	err := MetaJson.Unmarshal(b, &t.Val)
	if err != nil {
		return err
	}
	t.Presence = Presence{true}
	t.Nullity = Nullity{false}
	return nil
}
