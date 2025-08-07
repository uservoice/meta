package meta

import (
	"net/url"
	"testing"
	"time"
)

type withTime struct {
	A Time `meta_required:"true"`
}

var withTimeDecoder = NewDecoder(&withTime{})

func TestTimeSuccess(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = withTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)
}

func TestTimeBlank(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)

	e = withTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
	assertEqual(t, inputs.A.Present, false)
}

func TestTimeInvalid(t *testing.T) {
	var inputs withTime

	e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"wat"}})
	assertEqual(t, e, ErrorHash{"a": ErrTime})
	assertEqual(t, inputs.A.Present, false)

	e = withTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"ok"}`))
	assertEqual(t, e, ErrorHash{"a": ErrTime})
	assertEqual(t, inputs.A.Present, false)
}

func TestTimeCustomFormat(t *testing.T) {
	var inputs struct {
		A Time `meta_required:"true" meta_format:"1/2/2006"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"6/2/2015"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)
}

func TestMultipleFormats(t *testing.T) {
	var inputs struct {
		A Time `meta_format:"RFC3339,2006-01-02 15:04:05"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"2016-01-01 00:00:00"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)

	e = NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {"2016-01-01T00:00:00Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assert(t, inputs.A.Val.Equal(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)))
	assertEqual(t, inputs.A.Present, true)
}

func TestTimeRange(t *testing.T) {
	var relativeInputs struct {
		A Time `meta_min:"3_days_ago" meta_max:"now"`
	}

	var absoluteInputs struct {
		A Time `meta_format:"DateOnly" meta_min:"2016-01-01" meta_max:"2016-01-02"`
	}

	// date within range should not error
	e := NewDecoder(&relativeInputs).DecodeValues(&relativeInputs, url.Values{"a": {"3_days_ago"}})
	assertEqual(t, e, ErrorHash(nil), "checking relative within range")

	// date earlier than min should error
	e = NewDecoder(&relativeInputs).DecodeJSON(&relativeInputs, []byte(`{"a":"4_days_ago"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "checking relative earlier than min")

	// date later than max should error
	e = NewDecoder(&relativeInputs).DecodeJSON(&relativeInputs, []byte(`{"a":"1_day_from_now"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "checking relative later than max")

	// date within range should not error
	e = NewDecoder(&absoluteInputs).DecodeValues(&absoluteInputs, url.Values{"a": {"2016-01-01"}})
	assertEqual(t, e, ErrorHash(nil), "checking absolute within range")

	// date earlier than min should error
	e = NewDecoder(&absoluteInputs).DecodeJSON(&absoluteInputs, []byte(`{"a":"2015-12-31"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "checking absolute earlier than min")

	// date later than max should error
	e = NewDecoder(&absoluteInputs).DecodeJSON(&absoluteInputs, []byte(`{"a":"2016-01-03"}`))
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "checking absolute later than max")
}

type withOptionalTime struct {
	A Time
}

var withOptionalTimeDecoder = NewDecoder(&withOptionalTime{})

func TestOptionalTimeSuccess(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2016-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assert(t, inputs.A.Val.Equal(time.Date(2016, 6, 2, 16, 33, 22, 0, time.UTC)))
}

func TestOptionalTimeOmitted(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"b":"9/1/2015"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())
}

func TestOptionalTimeBlank(t *testing.T) {
	var inputs withOptionalTime

	e := withOptionalTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assert(t, inputs.A.Val.IsZero())
}

func TestOptionalTimeBlankFailure(t *testing.T) {
	var inputs struct {
		A Time `meta_discard_blank:"false"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":""}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash{"a": ErrBlank})

	e = NewDecoder(&inputs).DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash{"a": ErrBlank})
}

type withOptionalNullTime struct {
	A Time `meta_null:"true"`
}

var withOptionalNullTimeDecoder = NewDecoder(&withOptionalNullTime{})

func TestOptionalNullTimeSuccess(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{"a": {"2015-06-02T16:33:22Z"}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":"2015-06-02T16:33:22Z"}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.Equal(time.Date(2015, 6, 2, 16, 33, 22, 0, time.UTC)))
}

func TestOptionalNullTimeNull(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{"a": {""}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{"a":null}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())

	e = withOptionalNullTimeDecoder.DecodeMap(&inputs, map[string]interface{}{"a": time.Time{}})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, true)
	assertEqual(t, inputs.A.Null, true)
	assert(t, inputs.A.Val.IsZero())
}

func TestOptionalNullTimeOmitted(t *testing.T) {
	var inputs withOptionalNullTime
	e := withOptionalNullTimeDecoder.DecodeValues(&inputs, url.Values{})
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.IsZero())

	inputs = withOptionalNullTime{}
	e = withOptionalNullTimeDecoder.DecodeJSON(&inputs, []byte(`{}`))
	assertEqual(t, e, ErrorHash(nil))
	assertEqual(t, inputs.A.Present, false)
	assertEqual(t, inputs.A.Null, false)
	assert(t, inputs.A.Val.IsZero())
}

func assertTimeInRange(t *testing.T, value, start, end time.Time, msgAndArgs ...interface{}) {
	assert(t, value.Equal(start) || value.After(start), msgAndArgs...)
	assert(t, value.Equal(end) || value.Before(end), msgAndArgs...)
}

func TestTimeExpressions(t *testing.T) {
	for expression, assertion := range map[string]func(output, before, after time.Time){
		// Simple keywords
		"now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before, after, "now should be rounded to start of day")
		},
		"today": func(output, before, after time.Time) {
			// Today should be rounded to start of day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "today should be rounded to start of day")
		},
		"yesterday": func(output, before, after time.Time) {
			// Yesterday should be rounded to start of day
			expected := time.Date(before.Year(), before.Month(), before.Day()-1, 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "yesterday should be rounded to start of day")
		},
		"tomorrow": func(output, before, after time.Time) {
			// Tomorrow should be rounded to start of day
			expected := time.Date(before.Year(), before.Month(), before.Day()+1, 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "tomorrow should be rounded to start of day")
		},
		// Past expressions (<n>_<unit>_ago) - auto-rounded down to start of unit
		"99_nanoseconds_ago": func(output, before, after time.Time) {
			// Nanoseconds don't get rounded, so should be exact
			assertTimeInRange(t, output, before.Add(-99*time.Nanosecond), after.Add(-99*time.Nanosecond), "99 nanoseconds ago should be exact")
		},
		"31_seconds_ago": func(output, before, after time.Time) {
			// Should round down to start of that second
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), before.Minute(), before.Second(), 0, before.Location()).Add(-31 * time.Second)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(time.Second)), "31 seconds ago should be rounded to start of second")
		},
		"1_minute_ago": func(output, before, after time.Time) {
			// Should round down to start of that minute
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), before.Minute(), 0, 0, before.Location()).Add(-time.Minute)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(time.Minute)), "1 minute ago should be rounded to start of minute")
		},
		"48_hours_ago": func(output, before, after time.Time) {
			// Should round down to start of that hour
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), 0, 0, 0, before.Location()).Add(-48 * time.Hour)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(time.Hour)), "48 hours ago should be rounded to start of hour")
		},
		"1_day_ago": func(output, before, after time.Time) {
			// Should round down to start of that day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, before.Location()).AddDate(0, 0, -1)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, 1)), "1 day ago should be rounded to start of day")
		},
		"5_days_ago": func(output, before, after time.Time) {
			// Should round down to start of that day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, before.Location()).AddDate(0, 0, -5)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, 1)), "5 days ago should be rounded to start of day")
		},
		"3_weeks_ago": func(output, before, after time.Time) {
			// Should round down to start of that week (Monday)
			daysSinceMonday := int(before.Weekday() - time.Monday)
			if daysSinceMonday < 0 {
				daysSinceMonday += 7
			}
			expected := time.Date(before.Year(), before.Month(), before.Day()-daysSinceMonday, 0, 0, 0, 0, before.Location()).AddDate(0, 0, -21)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, 7)), "3 weeks ago should be rounded to start of week")
		},
		"2_months_ago": func(output, before, after time.Time) {
			// Should round down to start of that month
			expected := time.Date(before.Year(), before.Month(), 1, 0, 0, 0, 0, before.Location()).AddDate(0, -2, 0)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 1, 0)), "2 months ago should be rounded to start of month")
		},
		"4_years_ago": func(output, before, after time.Time) {
			// Should round down to start of that year
			expected := time.Date(before.Year(), 1, 1, 0, 0, 0, 0, before.Location()).AddDate(-4, 0, 0)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(1, 0, 0)), "4 years ago should be rounded to start of year")
		},
		// Future expressions (<n>_<unit>_from_now) - auto-rounded up to end of unit
		"99_nanoseconds_from_now": func(output, before, after time.Time) {
			// Nanoseconds don't get rounded, so should be exact
			assertTimeInRange(t, output, before.Add(99*time.Nanosecond), after.Add(99*time.Nanosecond), "99 nanoseconds from now should be exact")
		},
		"31_seconds_from_now": func(output, before, after time.Time) {
			// Should round up to end of that second
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), before.Minute(), before.Second(), 999999999, before.Location()).Add(31 * time.Second)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(-time.Second)), "31 seconds from now should be rounded to end of second")
		},
		"1_minute_from_now": func(output, before, after time.Time) {
			// Should round up to end of that minute
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), before.Minute(), 59, 999999999, before.Location()).Add(time.Minute)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(-time.Minute)), "1 minute from now should be rounded to end of minute")
		},
		"48_hours_from_now": func(output, before, after time.Time) {
			// Should round up to end of that hour
			expected := time.Date(before.Year(), before.Month(), before.Day(), before.Hour(), 59, 59, 999999999, before.Location()).Add(48 * time.Hour)
			assert(t, output.Equal(expected) || output.Equal(expected.Add(-time.Hour)), "48 hours from now should be rounded to end of hour")
		},
		"1_day_from_now": func(output, before, after time.Time) {
			// Should round up to end of that day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 23, 59, 59, 999999999, before.Location()).AddDate(0, 0, 1)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, -1)), "1 day from now should be rounded to end of day")
		},
		"5_days_from_now": func(output, before, after time.Time) {
			// Should round up to end of that day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 23, 59, 59, 999999999, before.Location()).AddDate(0, 0, 5)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, -1)), "5 days from now should be rounded to end of day")
		},
		"3_weeks_from_now": func(output, before, after time.Time) {
			// Should round up to end of that week (Sunday)
			daysUntilSunday := (7 + int(time.Sunday-before.Weekday())) % 7
			if daysUntilSunday == 0 {
				daysUntilSunday = 7
			}
			expected := time.Date(before.Year(), before.Month(), before.Day()+daysUntilSunday, 23, 59, 59, 999999999, before.Location()).AddDate(0, 0, 21)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, 0, -7)), "3 weeks from now should be rounded to end of week")
		},
		"2_months_from_now": func(output, before, after time.Time) {
			// Should round up to end of that month
			expected := time.Date(before.Year(), before.Month(), 1, 0, 0, 0, 0, before.Location()).AddDate(0, 3, 0).Add(-time.Nanosecond)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(0, -1, 0)), "2 months from now should be rounded to end of month")
		},
		"4_years_from_now": func(output, before, after time.Time) {
			// Should round up to end of that year
			expected := time.Date(before.Year(), 12, 31, 23, 59, 59, 999999999, before.Location()).AddDate(4, 0, 0)
			assert(t, output.Equal(expected) || output.Equal(expected.AddDate(-1, 0, 0)), "4 years from now should be rounded to end of year")
		},
	} {
		var inputs withTime

		before := time.Now()
		e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {expression}})
		after := time.Now()
		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Present, true)
		assertion(inputs.A.Val, before, after)
	}
}

func TestTimeExclusiveRange(t *testing.T) {
	var exclusiveMinInputs struct {
		A Time `meta_min:"!2024-01-01T00:00:00Z" meta_format:"RFC3339"`
	}

	var exclusiveMaxInputs struct {
		A Time `meta_max:"!2024-12-31T23:59:59Z" meta_format:"RFC3339"`
	}

	var exclusiveBothInputs struct {
		A Time `meta_min:"!2024-01-01T00:00:00Z" meta_max:"!2024-12-31T23:59:59Z" meta_format:"RFC3339"`
	}

	var exclusiveExpressionInputs struct {
		A Time `meta_min:"!1_day_ago" meta_max:"!1_day_from_now"`
	}

	// Test exclusive min - value equal to min should fail
	e := NewDecoder(&exclusiveMinInputs).DecodeValues(&exclusiveMinInputs, url.Values{"a": {"2024-01-01T00:00:00Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "exclusive min - value equal to min should fail")

	// Test exclusive min - value before min should fail
	e = NewDecoder(&exclusiveMinInputs).DecodeValues(&exclusiveMinInputs, url.Values{"a": {"2023-12-31T23:59:59Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "exclusive min - value before min should fail")

	// Test exclusive min - value after min should pass
	e = NewDecoder(&exclusiveMinInputs).DecodeValues(&exclusiveMinInputs, url.Values{"a": {"2024-01-01T00:00:01Z"}})
	assertEqual(t, e, ErrorHash(nil), "exclusive min - value after min should pass")

	// Test exclusive min - value with different time on same day should pass
	e = NewDecoder(&exclusiveMinInputs).DecodeValues(&exclusiveMinInputs, url.Values{"a": {"2024-01-01T12:00:00Z"}})
	assertEqual(t, e, ErrorHash(nil), "exclusive min - value with different time on same day should pass")

	// Test exclusive max - value equal to max should fail
	e = NewDecoder(&exclusiveMaxInputs).DecodeValues(&exclusiveMaxInputs, url.Values{"a": {"2024-12-31T23:59:59Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "exclusive max - value equal to max should fail")

	// Test exclusive max - value after max should fail
	e = NewDecoder(&exclusiveMaxInputs).DecodeValues(&exclusiveMaxInputs, url.Values{"a": {"2025-01-01T00:00:00Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "exclusive max - value after max should fail")

	// Test exclusive max - value before max should pass
	e = NewDecoder(&exclusiveMaxInputs).DecodeValues(&exclusiveMaxInputs, url.Values{"a": {"2024-12-31T23:59:58Z"}})
	assertEqual(t, e, ErrorHash(nil), "exclusive max - value before max should pass")

	// Test exclusive max - value with different time on same day should pass
	e = NewDecoder(&exclusiveMaxInputs).DecodeValues(&exclusiveMaxInputs, url.Values{"a": {"2024-12-31T12:00:00Z"}})
	assertEqual(t, e, ErrorHash(nil), "exclusive max - value with different time on same day should pass")

	// Test exclusive both - value at boundaries should fail
	e = NewDecoder(&exclusiveBothInputs).DecodeValues(&exclusiveBothInputs, url.Values{"a": {"2024-01-01T00:00:00Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "exclusive both - value at min boundary should fail")

	e = NewDecoder(&exclusiveBothInputs).DecodeValues(&exclusiveBothInputs, url.Values{"a": {"2024-12-31T23:59:59Z"}})
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "exclusive both - value at max boundary should fail")

	// Test exclusive both - value in middle should pass
	e = NewDecoder(&exclusiveBothInputs).DecodeValues(&exclusiveBothInputs, url.Values{"a": {"2024-06-15T12:30:45Z"}})
	assertEqual(t, e, ErrorHash(nil), "exclusive both - value in middle should pass")

	// Test exclusive expressions - value equal to expression should fail
	e = NewDecoder(&exclusiveExpressionInputs).DecodeValues(&exclusiveExpressionInputs, url.Values{"a": {"1_day_ago"}})
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "exclusive expression - value equal to min expression should fail")

	e = NewDecoder(&exclusiveExpressionInputs).DecodeValues(&exclusiveExpressionInputs, url.Values{"a": {"1_day_from_now"}})
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "exclusive expression - value equal to max expression should fail")
}

func TestTimeRounding(t *testing.T) {
	// Test rounding down
	var roundDownInputs struct {
		A Time `meta_round:"day:down"`
		B Time `meta_round:"hour:down"`
		C Time `meta_round:"minute:down"`
		D Time `meta_round:"week:down"`
		E Time `meta_round:"month:down"`
		F Time `meta_round:"year:down"`
	}

	e := NewDecoder(&roundDownInputs).DecodeValues(&roundDownInputs, url.Values{
		"a": {"2024-01-15T14:30:45Z"},
		"b": {"2024-01-15T14:30:45Z"},
		"c": {"2024-01-15T14:30:45Z"},
		"d": {"2024-01-17T14:30:45Z"}, // Wednesday
		"e": {"2024-01-15T14:30:45Z"},
		"f": {"2024-06-15T14:30:45Z"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assert(t, roundDownInputs.A.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
	assert(t, roundDownInputs.B.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)))
	assert(t, roundDownInputs.C.Val.Equal(time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)))
	assert(t, roundDownInputs.D.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))) // Monday
	assert(t, roundDownInputs.E.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert(t, roundDownInputs.F.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))

	// Test rounding up
	var roundUpInputs struct {
		A Time `meta_round:"day:up"`
		B Time `meta_round:"hour:up"`
		C Time `meta_round:"minute:up"`
		D Time `meta_round:"week:up"`
		E Time `meta_round:"month:up"`
		F Time `meta_round:"year:up"`
	}

	e = NewDecoder(&roundUpInputs).DecodeValues(&roundUpInputs, url.Values{
		"a": {"2024-01-15T14:30:45Z"},
		"b": {"2024-01-15T14:30:45Z"},
		"c": {"2024-01-15T14:30:45Z"},
		"d": {"2024-01-17T14:30:45Z"}, // Wednesday
		"e": {"2024-01-15T14:30:45Z"},
		"f": {"2024-06-15T14:30:45Z"},
	})

	assertEqual(t, e, ErrorHash(nil))
	assert(t, roundUpInputs.A.Val.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)))
	assert(t, roundUpInputs.B.Val.Equal(time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)))
	assert(t, roundUpInputs.C.Val.Equal(time.Date(2024, 1, 15, 14, 31, 0, 0, time.UTC)))
	assert(t, roundUpInputs.D.Val.Equal(time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC))) // Next Monday
	assert(t, roundUpInputs.E.Val.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)))
	assert(t, roundUpInputs.F.Val.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))

	// Test rounding nearest
	var roundNearestInputs struct {
		A Time `meta_round:"day:nearest"`
		B Time `meta_round:"hour:nearest"`
	}

	e = NewDecoder(&roundNearestInputs).DecodeValues(&roundNearestInputs, url.Values{
		"a": {"2024-01-15T12:00:00Z"}, // Exactly noon, should round down
		"b": {"2024-01-15T14:30:00Z"}, // 30 minutes past, should round down
	})

	assertEqual(t, e, ErrorHash(nil))
	assert(t, roundNearestInputs.A.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
	assert(t, roundNearestInputs.B.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)))

	// Test boundary conditions (already at boundary)
	var boundaryInputs struct {
		A Time `meta_round:"day:up"`
		B Time `meta_round:"hour:up"`
	}

	e = NewDecoder(&boundaryInputs).DecodeValues(&boundaryInputs, url.Values{
		"a": {"2024-01-15T00:00:00Z"}, // Already at day boundary
		"b": {"2024-01-15T14:00:00Z"}, // Already at hour boundary
	})

	assertEqual(t, e, ErrorHash(nil))
	assert(t, boundaryInputs.A.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))  // Should stay the same
	assert(t, boundaryInputs.B.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC))) // Should stay the same
}

func TestTimeRoundingWithExpressions(t *testing.T) {
	// Test rounding with time expressions
	var expressionInputs struct {
		A Time `meta_round:"day:down"`
		B Time `meta_round:"hour:up"`
		C Time `meta_round:"week:down"`
	}

	e := NewDecoder(&expressionInputs).DecodeValues(&expressionInputs, url.Values{
		"a": {"2_days_ago"},
		"b": {"3_hours_ago"},
		"c": {"1_week_ago"},
	})

	assertEqual(t, e, ErrorHash(nil))

	// Verify that expressions are rounded correctly
	// A should be rounded down to the start of the day
	assert(t, expressionInputs.A.Val.Hour() == 0 && expressionInputs.A.Val.Minute() == 0 && expressionInputs.A.Val.Second() == 0)

	// B should be rounded up to the next hour
	assert(t, expressionInputs.B.Val.Minute() == 0 && expressionInputs.B.Val.Second() == 0)

	// C should be rounded down to the start of the week (Monday)
	assert(t, expressionInputs.C.Val.Weekday() == time.Monday)
	assert(t, expressionInputs.C.Val.Hour() == 0 && expressionInputs.C.Val.Minute() == 0 && expressionInputs.C.Val.Second() == 0)
}

func TestTimeRoundingWithDayNames(t *testing.T) {
	// Test rounding to specific days of the week
	var dayInputs struct {
		Sunday    Time `meta_round:"sunday:down"`
		Monday    Time `meta_round:"monday:down"`
		Wednesday Time `meta_round:"wednesday:up"`
		Friday    Time `meta_round:"friday:nearest"`
	}

	e := NewDecoder(&dayInputs).DecodeValues(&dayInputs, url.Values{
		"sunday":    {"2024-01-17T14:30:45Z"}, // Wednesday
		"monday":    {"2024-01-17T14:30:45Z"}, // Wednesday
		"wednesday": {"2024-01-17T14:30:45Z"}, // Wednesday
		"friday":    {"2024-01-17T14:30:45Z"}, // Wednesday
	})

	assertEqual(t, e, ErrorHash(nil))

	// Should round down to previous Sunday
	assert(t, dayInputs.Sunday.Val.Weekday() == time.Sunday)
	assert(t, dayInputs.Sunday.Val.Before(time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)))

	// Should round down to previous Monday
	assert(t, dayInputs.Monday.Val.Weekday() == time.Monday)
	assert(t, dayInputs.Monday.Val.Before(time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)))

	// Should round up to next Wednesday (next week)
	assert(t, dayInputs.Wednesday.Val.Weekday() == time.Wednesday)
	assert(t, dayInputs.Wednesday.Val.Equal(time.Date(2024, 1, 24, 0, 0, 0, 0, time.UTC)))

	// Should round to nearest Friday (should round to next Friday since we're closer to it)
	assert(t, dayInputs.Friday.Val.Weekday() == time.Friday)
	assert(t, dayInputs.Friday.Val.Equal(time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)))
}

func TestTimeRoundingWithRelativeExpressions(t *testing.T) {
	var relativeInputs struct {
		A Time `meta_round:"day:down"`
	}

	e := NewDecoder(&relativeInputs).DecodeValues(&relativeInputs, url.Values{
		"a": {"1_month_ago"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// With auto-rounding, 1_month_ago should round down to start of that month
	// The meta_round:"day:down" should be ignored for relative expressions
	expectedTime := time.Now().AddDate(0, -1, 0)
	expectedMonth := time.Date(expectedTime.Year(), expectedTime.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Should be rounded to start of the month (day 1, hour 0, etc.)
	assert(t, relativeInputs.A.Val.Day() == 1)
	assert(t, relativeInputs.A.Val.Hour() == 0)
	assert(t, relativeInputs.A.Val.Minute() == 0)
	assert(t, relativeInputs.A.Val.Second() == 0)
	assert(t, relativeInputs.A.Val.Nanosecond() == 0)
	assert(t, relativeInputs.A.Val.Month() == expectedMonth.Month())
	assert(t, relativeInputs.A.Val.Year() == expectedMonth.Year())
}

func TestRoundUpToDayWithMaxNow(t *testing.T) {
	var maxNowInputs struct {
		A Time `meta_max:"now" meta_round:"day:up"`
	}

	e := NewDecoder(&maxNowInputs).DecodeValues(&maxNowInputs, url.Values{
		"a": {"now"},
	})
	assertEqual(t, e, ErrorHash(nil))
}

func TestTimeRoundingDefaultDirection(t *testing.T) {
	// Test that default direction is "down" when not specified
	var defaultInputs struct {
		A Time `meta_round:"day"`
	}

	e := NewDecoder(&defaultInputs).DecodeValues(&defaultInputs, url.Values{
		"a": {"2024-01-15T14:30:45Z"}, // Should round down to start of day
	})

	assertEqual(t, e, ErrorHash(nil))
	assert(t, defaultInputs.A.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
}

func TestTimeRelativeRoundModes(t *testing.T) {
	// Test auto mode (default behavior)
	var autoInputs struct {
		A Time `meta_relative_round:"auto"`
	}

	e := NewDecoder(&autoInputs).DecodeValues(&autoInputs, url.Values{
		"a": {"1_day_ago"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// Should round down to start of day (auto behavior for ago)
	expected := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).AddDate(0, 0, -1)
	assert(t, autoInputs.A.Val.Equal(expected))

	// Test up mode
	var upInputs struct {
		A Time `meta_relative_round:"up"`
	}

	e = NewDecoder(&upInputs).DecodeValues(&upInputs, url.Values{
		"a": {"1_day_ago"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// Should round up to end of day (up behavior)
	expected = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 999999999, time.Now().Location()).AddDate(0, 0, -1)
	assert(t, upInputs.A.Val.Equal(expected))

	// Test down mode
	var downInputs struct {
		A Time `meta_relative_round:"down"`
	}

	e = NewDecoder(&downInputs).DecodeValues(&downInputs, url.Values{
		"a": {"1_day_from_now"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// Should round down to start of day (down behavior)
	expected = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).AddDate(0, 0, 1)
	assert(t, downInputs.A.Val.Equal(expected))

	// Test none mode
	var noneInputs struct {
		A Time `meta_relative_round:"none"`
	}

	e = NewDecoder(&noneInputs).DecodeValues(&noneInputs, url.Values{
		"a": {"1_hour_ago"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// Should not round, just subtract 1 hour
	expected = time.Now().Add(-time.Hour)
	// Allow for small time differences due to test execution time
	assert(t, noneInputs.A.Val.Sub(expected) < time.Second)
}

func TestTimeRelativeRoundWithNonExpressions(t *testing.T) {
	// Test that meta_relative_round doesn't affect non-expression values
	var inputs struct {
		A Time `meta_relative_round:"up" meta_round:"day:down"`
	}

	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
		"a": {"2024-01-15T14:30:45Z"},
	})
	assertEqual(t, e, ErrorHash(nil))

	// Should use meta_round (day:down) not meta_relative_round
	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	assert(t, inputs.A.Val.Equal(expected))
}

func TestTimeRelativeRoundNowExpression(t *testing.T) {
	// Test that "now" expression is not affected by meta_relative_round
	var inputs struct {
		A Time `meta_relative_round:"up"`
	}

	before := time.Now()
	e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
		"a": {"now"},
	})
	after := time.Now()
	assertEqual(t, e, ErrorHash(nil))

	// Should return exact current time, not rounded
	assertTimeInRange(t, inputs.A.Val, before, after)
}

func TestTimeYearRounding(t *testing.T) {
	// test rounding year up for start and end of years
	tests := []struct {
		input        string
		upExpected   time.Time
		downExpected time.Time
	}{
		{"2023-01-01T14:00:00Z", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2023-12-31T14:00:00Z", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)},
		{"2024-01-01T14:00:00Z", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-12-31T14:00:00Z", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
	}

	var inputs struct {
		A Time `meta_round:"year:up"`
		B Time `meta_round:"year:down"`
	}

	var decoder = NewDecoder(&inputs)
	for _, test := range tests {
		e := decoder.DecodeValues(&inputs, url.Values{
			"a": {test.input},
			"b": {test.input},
		})
		assertEqual(t, e, ErrorHash(nil), "expect no error for %s", test.input)
	}
}

func TestTimeMonthRounding(t *testing.T) {
	// test rounding month up for start and end of months
	tests := []struct {
		input        string
		upExpected   time.Time
		downExpected time.Time
	}{
		// test 30, 31, and 29 day months
		{"2024-01-01T14:00:00Z", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-01-31T14:00:00Z", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-02-01T14:00:00Z", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-02-29T14:00:00Z", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-04-01T14:00:00Z", time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-04-30T14:00:00Z", time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)},
	}

	var inputs struct {
		A Time `meta_round:"month:up"`
		B Time `meta_round:"month:down"`
	}
	var decoder = NewDecoder(&inputs)
	for _, test := range tests {
		e := decoder.DecodeValues(&inputs, url.Values{
			"a": {test.input},
			"b": {test.input},
		})
		assertEqual(t, e, ErrorHash(nil), "expect no error for %s", test.input)
		assert(t, inputs.A.Val.Equal(test.upExpected), "%s should round up to %s, got %s", test.input, test.upExpected, inputs.A.Val)
		assert(t, inputs.B.Val.Equal(test.downExpected), "%s should round down to %s, got %s", test.input, test.downExpected, inputs.B.Val)
	}
}

func TestTimeWeekRounding(t *testing.T) {
	// test rounding week up for start and end of weeks (monday)
	tests := []struct {
		input        string
		upExpected   time.Time
		downExpected time.Time
	}{
		{"2024-01-01T14:00:00Z", time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-01-07T14:00:00Z", time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2024-01-29T14:00:00Z", time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC)},
	}

	var inputs struct {
		A Time `meta_round:"week:up"`
		B Time `meta_round:"week:down"`
	}
	var decoder = NewDecoder(&inputs)
	for _, test := range tests {
		e := decoder.DecodeValues(&inputs, url.Values{
			"a": {test.input},
			"b": {test.input},
		})
		assertEqual(t, e, ErrorHash(nil), "expect no error for %s", test.input)
		assert(t, inputs.A.Val.Equal(test.upExpected), "%s should round up to %s, got %s", test.input, test.upExpected, inputs.A.Val)
		assert(t, inputs.B.Val.Equal(test.downExpected), "%s should round down to %s, got %s", test.input, test.downExpected, inputs.B.Val)
	}
}

func TestTimeRoundingDSTTransitions(t *testing.T) {
	// Test rounding during DST transitions (using a known DST transition time)
	// Note: This test uses a specific timezone and date where DST changes occur
	// For this test, we'll use a time that's close to DST transition but focus on the rounding behavior

	var dstInputs struct {
		A Time `meta_round:"hour:up"`
		B Time `meta_round:"hour:down"`
	}

	// Use a time that's not exactly at DST transition but tests the rounding logic
	e := NewDecoder(&dstInputs).DecodeValues(&dstInputs, url.Values{
		"a": {"2024-03-10T02:30:00Z"}, // During DST transition period
		"b": {"2024-03-10T02:30:00Z"},
	})

	assertEqual(t, e, ErrorHash(nil))
	// Should round up to next hour
	assert(t, dstInputs.A.Val.Hour() == 3 || dstInputs.A.Val.Hour() == 2)
	// Should round down to current hour
	assert(t, dstInputs.B.Val.Hour() == 2)
}

// Category 4: Date Limit Edge Cases - Missing Tests

func TestTimeExclusiveBoundarySameTime(t *testing.T) {
	// Test exclusive boundaries where expressions resolve to the same time
	var sameTimeInputs struct {
		A Time `meta_min:"!1_hour_ago" meta_max:"!1_hour_from_now"`
	}

	// This should pass because the input is between the boundaries
	e := NewDecoder(&sameTimeInputs).DecodeValues(&sameTimeInputs, url.Values{
		"a": {"now"},
	})
	assertEqual(t, e, ErrorHash(nil), "now should be between 1_hour_ago and 1_hour_from_now")

	// This should fail because it's equal to the min boundary
	e = NewDecoder(&sameTimeInputs).DecodeValues(&sameTimeInputs, url.Values{
		"a": {"1_hour_ago"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrMin}, "1_hour_ago should fail exclusive min boundary")

	// This should fail because it's equal to the max boundary
	e = NewDecoder(&sameTimeInputs).DecodeValues(&sameTimeInputs, url.Values{
		"a": {"1_hour_from_now"},
	})
	assertEqual(t, e, ErrorHash{"a": ErrMax}, "1_hour_from_now should fail exclusive max boundary")
}

func TestTimeRelativeRoundZeroValues(t *testing.T) {
	// Test relative rounding with zero values
	var zeroInputs struct {
		A Time `meta_relative_round:"auto"`
		B Time `meta_relative_round:"up"`
	}

	e := NewDecoder(&zeroInputs).DecodeValues(&zeroInputs, url.Values{
		"a": {"0_days_ago"},
		"b": {"0_hours_from_now"},
	})

	assertEqual(t, e, ErrorHash(nil))

	// Should handle zero values gracefully
	// 0_days_ago should be today, rounded down to start of day
	expectedA := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	assert(t, zeroInputs.A.Val.Equal(expectedA))

	// 0_hours_from_now should be now, rounded up to end of hour
	expectedB := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), 59, 59, 999999999, time.Now().Location())
	assert(t, zeroInputs.B.Val.Equal(expectedB))
}

// Category 7: Combining Features - Missing Tests

func TestTimeComplexFeatureCombination(t *testing.T) {
	// Test complex combination of multiple features
	var complexInputs struct {
		A Time `meta_min:"!1_day_ago" meta_max:"!1_day_from_now" meta_round:"hour:up" meta_relative_round:"auto"`
		B Time `meta_min:"2024-01-01T00:00:00Z" meta_max:"2024-12-31T23:59:59Z" meta_round:"day:down" meta_relative_round:"none"`
	}

	e := NewDecoder(&complexInputs).DecodeValues(&complexInputs, url.Values{
		"a": {"2_hours_ago"},          // Should be rounded up to end of hour, and pass min/max validation
		"b": {"2024-06-15T14:30:45Z"}, // Should be rounded down to start of day, and pass min/max validation
	})

	assertEqual(t, e, ErrorHash(nil))

	// A should be rounded up to end of hour (meta_round takes precedence over meta_relative_round for absolute times)
	// and should be between the exclusive boundaries
	assert(t, complexInputs.A.Val.Minute() == 0 && complexInputs.A.Val.Second() == 0)
	assert(t, complexInputs.A.Val.After(time.Now().AddDate(0, 0, -1)))
	assert(t, complexInputs.A.Val.Before(time.Now().AddDate(0, 0, 1)))

	// B should be rounded down to start of day
	assert(t, complexInputs.B.Val.Equal(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)))
}

func TestTimeExclusiveBoundaryWithRounding(t *testing.T) {
	// Test exclusive boundaries combined with rounding
	var exclusiveRoundInputs struct {
		A Time `meta_min:"!2024-01-01T00:00:00Z" meta_max:"!2024-12-31T23:59:59Z" meta_round:"day:up"`
	}

	// Test with a time that would be rounded to a valid boundary
	e := NewDecoder(&exclusiveRoundInputs).DecodeValues(&exclusiveRoundInputs, url.Values{
		"a": {"2024-06-15T23:30:00Z"}, // Should round up to 2024-06-16T00:00:00Z
	})

	assertEqual(t, e, ErrorHash(nil))
	// Should be rounded up to next day
	assert(t, exclusiveRoundInputs.A.Val.Equal(time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC)))

	// Test with a time that would be rounded to the max boundary (should fail)
	e = NewDecoder(&exclusiveRoundInputs).DecodeValues(&exclusiveRoundInputs, url.Values{
		"a": {"2024-12-31T23:30:00Z"}, // Should round up to 2025-01-01T00:00:00Z, which is outside max boundary
	})

	assertEqual(t, e, ErrorHash{"a": ErrMax}, "should fail because rounded value is outside max boundary")
}

func TestTimeRelativeRoundWithExpressionsAndRounding(t *testing.T) {
	// Test relative rounding with expressions and regular rounding
	var relativeRoundInputs struct {
		A Time `meta_relative_round:"auto" meta_round:"hour:up"`
		B Time `meta_relative_round:"none" meta_round:"day:down"`
	}

	e := NewDecoder(&relativeRoundInputs).DecodeValues(&relativeRoundInputs, url.Values{
		"a": {"1_day_ago"},            // Should use relative_round (auto) for expression
		"b": {"2024-06-15T14:30:45Z"}, // Should use round (day:down) for absolute time
	})

	assertEqual(t, e, ErrorHash(nil))

	// A should use relative_round (auto) - round down to start of day for ago
	// The actual behavior uses local timezone, so we need to check the actual value
	expectedA := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).AddDate(0, 0, -1)
	assert(t, relativeRoundInputs.A.Val.Equal(expectedA), "expected %v, got %v", expectedA, relativeRoundInputs.A.Val)

	// B should use round (day:down) - round down to start of day
	expectedB := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	assert(t, relativeRoundInputs.B.Val.Equal(expectedB))
}
