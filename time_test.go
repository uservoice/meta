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
			assertTimeInRange(t, output, before, after, "now should be exact current time")
		},
		"today": func(output, before, after time.Time) {
			// Today should be start of day
			expected := time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "today should be start of day")
		},
		"yesterday": func(output, before, after time.Time) {
			// Yesterday should be start of previous day
			expected := time.Date(before.Year(), before.Month(), before.Day()-1, 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "yesterday should be start of previous day")
		},
		"tomorrow": func(output, before, after time.Time) {
			// Tomorrow should be start of next day
			expected := time.Date(before.Year(), before.Month(), before.Day()+1, 0, 0, 0, 0, before.Location())
			assert(t, output.Equal(expected), "tomorrow should be start of next day")
		},
		// Relative expressions - no rounding by default
		"99_nanoseconds_ago": func(output, before, after time.Time) {
			// Should be exact time
			assertTimeInRange(t, output, before.Add(-99*time.Nanosecond), after.Add(-99*time.Nanosecond), "99 nanoseconds ago should be exact")
		},
		"31_seconds_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(-31*time.Second), after.Add(-31*time.Second), "31 seconds ago should be exact")
		},
		"1_minute_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(-time.Minute), after.Add(-time.Minute), "1 minute ago should be exact")
		},
		"48_hours_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(-48*time.Hour), after.Add(-48*time.Hour), "48 hours ago should be exact")
		},
		"1_day_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, -1), after.AddDate(0, 0, -1), "1 day ago should be exact")
		},
		"5_days_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, -5), after.AddDate(0, 0, -5), "5 days ago should be exact")
		},
		"3_weeks_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, -21), after.AddDate(0, 0, -21), "3 weeks ago should be exact")
		},
		"2_months_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, -2, 0), after.AddDate(0, -2, 0), "2 months ago should be exact")
		},
		"4_years_ago": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(-4, 0, 0), after.AddDate(-4, 0, 0), "4 years ago should be exact")
		},
		// Future expressions - no rounding by default
		"99_nanoseconds_from_now": func(output, before, after time.Time) {
			// Should be exact time
			assertTimeInRange(t, output, before.Add(99*time.Nanosecond), after.Add(99*time.Nanosecond), "99 nanoseconds from now should be exact")
		},
		"31_seconds_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(31*time.Second), after.Add(31*time.Second), "31 seconds from now should be exact")
		},
		"1_minute_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(time.Minute), after.Add(time.Minute), "1 minute from now should be exact")
		},
		"48_hours_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.Add(48*time.Hour), after.Add(48*time.Hour), "48 hours from now should be exact")
		},
		"1_day_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, 1), after.AddDate(0, 0, 1), "1 day from now should be exact")
		},
		"5_days_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, 5), after.AddDate(0, 0, 5), "5 days from now should be exact")
		},
		"3_weeks_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 0, 21), after.AddDate(0, 0, 21), "3 weeks from now should be exact")
		},
		"2_months_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(0, 2, 0), after.AddDate(0, 2, 0), "2 months from now should be exact")
		},
		"4_years_from_now": func(output, before, after time.Time) {
			// Should be exact time, no rounding
			assertTimeInRange(t, output, before.AddDate(4, 0, 0), after.AddDate(4, 0, 0), "4 years from now should be exact")
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
	// Test comprehensive rounding scenarios covering all time units and directions

	// Test basic rounding scenarios with fixed struct tags
	t.Run("basic rounding", func(t *testing.T) {
		var inputs struct {
			DayDown     Time `meta_round:"day:down"`
			HourDown    Time `meta_round:"hour:down"`
			MinuteDown  Time `meta_round:"minute:down"`
			WeekDown    Time `meta_round:"week:down"`
			MonthDown   Time `meta_round:"month:down"`
			YearDown    Time `meta_round:"year:down"`
			DayUp       Time `meta_round:"day:up"`
			HourUp      Time `meta_round:"hour:up"`
			MinuteUp    Time `meta_round:"minute:up"`
			WeekUp      Time `meta_round:"week:up"`
			MonthUp     Time `meta_round:"month:up"`
			YearUp      Time `meta_round:"year:up"`
			DayNearest  Time `meta_round:"day:nearest"`
			HourNearest Time `meta_round:"hour:nearest"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"day_down":     {"2024-01-15T14:30:45Z"},
			"hour_down":    {"2024-01-15T14:30:45Z"},
			"minute_down":  {"2024-01-15T14:30:45Z"},
			"week_down":    {"2024-01-17T14:30:45Z"}, // Wednesday
			"month_down":   {"2024-01-15T14:30:45Z"},
			"year_down":    {"2024-06-15T14:30:45Z"},
			"day_up":       {"2024-01-15T14:30:45Z"},
			"hour_up":      {"2024-01-15T14:30:45Z"},
			"minute_up":    {"2024-01-15T14:30:45Z"},
			"week_up":      {"2024-01-17T14:30:45Z"}, // Wednesday
			"month_up":     {"2024-01-15T14:30:45Z"},
			"year_up":      {"2024-06-15T14:30:45Z"},
			"day_nearest":  {"2024-01-15T12:00:00Z"}, // Exactly noon
			"hour_nearest": {"2024-01-15T14:30:00Z"}, // 30 minutes past
		})

		assertEqual(t, e, ErrorHash(nil))

		// Test round down results
		assert(t, inputs.DayDown.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.HourDown.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)))
		assert(t, inputs.MinuteDown.Val.Equal(time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)))
		assert(t, inputs.WeekDown.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))) // Monday
		assert(t, inputs.MonthDown.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.YearDown.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))

		// Test round up results
		assert(t, inputs.DayUp.Val.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.HourUp.Val.Equal(time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)))
		assert(t, inputs.MinuteUp.Val.Equal(time.Date(2024, 1, 15, 14, 31, 0, 0, time.UTC)))
		assert(t, inputs.WeekUp.Val.Equal(time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC))) // Next Monday
		assert(t, inputs.MonthUp.Val.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.YearUp.Val.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))

		// Test round nearest results
		assert(t, inputs.DayNearest.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.HourNearest.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)))
	})

	// Test boundary conditions
	t.Run("boundary conditions", func(t *testing.T) {
		var inputs struct {
			DayBoundary  Time `meta_round:"day:up"`
			HourBoundary Time `meta_round:"hour:up"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"day_boundary":  {"2024-01-15T00:00:00Z"}, // Already at day boundary
			"hour_boundary": {"2024-01-15T14:00:00Z"}, // Already at hour boundary
		})

		assertEqual(t, e, ErrorHash(nil))
		assert(t, inputs.DayBoundary.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))   // Should stay the same
		assert(t, inputs.HourBoundary.Val.Equal(time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC))) // Should stay the same
	})

	// Test day name rounding
	t.Run("day name rounding", func(t *testing.T) {
		var inputs struct {
			Sunday    Time `meta_round:"sunday:down"`
			Monday    Time `meta_round:"monday:down"`
			Wednesday Time `meta_round:"wednesday:up"`
			Friday    Time `meta_round:"friday:nearest"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"sunday":    {"2024-01-17T14:30:45Z"}, // Wednesday
			"monday":    {"2024-01-17T14:30:45Z"}, // Wednesday
			"wednesday": {"2024-01-17T14:30:45Z"}, // Wednesday
			"friday":    {"2024-01-17T14:30:45Z"}, // Wednesday
		})

		assertEqual(t, e, ErrorHash(nil))

		// Should round down to previous Sunday
		assert(t, inputs.Sunday.Val.Weekday() == time.Sunday)
		assert(t, inputs.Sunday.Val.Before(time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)))

		// Should round down to previous Monday
		assert(t, inputs.Monday.Val.Weekday() == time.Monday)
		assert(t, inputs.Monday.Val.Before(time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)))

		// Should round up to next Wednesday (next week)
		assert(t, inputs.Wednesday.Val.Weekday() == time.Wednesday)
		assert(t, inputs.Wednesday.Val.Equal(time.Date(2024, 1, 24, 0, 0, 0, 0, time.UTC)))

		// Should round to nearest Friday (should round to next Friday since we're closer to it)
		assert(t, inputs.Friday.Val.Weekday() == time.Friday)
		assert(t, inputs.Friday.Val.Equal(time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)))
	})

	// Test month edge cases (31, 30, 29 day months)
	t.Run("month edge cases", func(t *testing.T) {
		var inputs struct {
			Month31Start Time `meta_round:"month:up"`
			Month31End   Time `meta_round:"month:up"`
			Month30Start Time `meta_round:"month:up"`
			Month30End   Time `meta_round:"month:up"`
			LeapFebStart Time `meta_round:"month:up"`
			LeapFebEnd   Time `meta_round:"month:up"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"month31_start":  {"2024-01-01T14:00:00Z"},
			"month31_end":    {"2024-01-31T14:00:00Z"},
			"month30_start":  {"2024-04-01T14:00:00Z"},
			"month30_end":    {"2024-04-30T14:00:00Z"},
			"leap_feb_start": {"2024-02-01T14:00:00Z"},
			"leap_feb_end":   {"2024-02-29T14:00:00Z"},
		})

		assertEqual(t, e, ErrorHash(nil))

		assert(t, inputs.Month31Start.Val.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)), "month 31 start %v, %v", inputs.Month31Start.Val, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.Month31End.Val.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)), "month 31 end %v, %v", inputs.Month31End.Val, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.Month30Start.Val.Equal(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)), "month 30 start %v, %v", inputs.Month30Start.Val, time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.Month30End.Val.Equal(time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)), "month 30 end %v, %v", inputs.Month30End.Val, time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.LeapFebStart.Val.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)), "leap feb start %v, %v", inputs.LeapFebStart.Val, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.LeapFebEnd.Val.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)), "leap feb end %v, %v", inputs.LeapFebEnd.Val, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
	})

	// Test year edge cases
	t.Run("year edge cases", func(t *testing.T) {
		var inputs struct {
			YearStart     Time `meta_round:"year:up"`
			YearEnd       Time `meta_round:"year:up"`
			LeapYearStart Time `meta_round:"year:down"`
			LeapYearEnd   Time `meta_round:"year:down"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"year_start":      {"2024-01-01T14:00:00Z"},
			"year_end":        {"2024-12-31T14:00:00Z"},
			"leap_year_start": {"2024-01-01T14:00:00Z"},
			"leap_year_end":   {"2024-12-31T14:00:00Z"},
		})

		assertEqual(t, e, ErrorHash(nil))

		assert(t, inputs.YearStart.Val.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), "year start %v, %v", inputs.YearStart.Val, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.YearEnd.Val.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)), "year end %v, %v", inputs.YearEnd.Val, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.LeapYearStart.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), "leap year start %v, %v", inputs.LeapYearStart.Val, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		assert(t, inputs.LeapYearEnd.Val.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), "leap year end %v, %v", inputs.LeapYearEnd.Val, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	})

	// Test week edge cases
	t.Run("week edge cases", func(t *testing.T) {
		var inputs struct {
			WeekStart  Time `meta_round:"week:up"`
			WeekEnd    Time `meta_round:"week:up"`
			WeekMiddle Time `meta_round:"week:up"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"week_start":  {"2024-01-01T14:00:00Z"},
			"week_end":    {"2024-01-07T14:00:00Z"},
			"week_middle": {"2024-01-29T14:00:00Z"},
		})

		assertEqual(t, e, ErrorHash(nil))

		assert(t, inputs.WeekStart.Val.Equal(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.WeekEnd.Val.Equal(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)))
		assert(t, inputs.WeekMiddle.Val.Equal(time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)))
	})

	// Test default direction (should be down)
	t.Run("default direction", func(t *testing.T) {
		var inputs struct {
			DefaultDirection Time `meta_round:"day"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"default_direction": {"2024-01-15T14:30:45Z"}, // Should round down to start of day
		})

		assertEqual(t, e, ErrorHash(nil))
		assert(t, inputs.DefaultDirection.Val.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
	})

	// Test DST transitions
	t.Run("dst transitions", func(t *testing.T) {
		var inputs struct {
			DstUp   Time `meta_round:"hour:up"`
			DstDown Time `meta_round:"hour:down"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"dst_up":   {"2024-03-10T02:30:00Z"}, // During DST transition period
			"dst_down": {"2024-03-10T02:30:00Z"},
		})

		assertEqual(t, e, ErrorHash(nil), "dst transitions")
		// Should round up to next hour
		assert(t, inputs.DstUp.Val.Hour() == 3 || inputs.DstUp.Val.Hour() == 2, "dst up %v", inputs.DstUp.Val)
		// Should round down to current hour
		assert(t, inputs.DstDown.Val.Hour() == 2, "dst down %v", inputs.DstDown.Val)
	})

	// Test expression rounding
	t.Run("expression rounding", func(t *testing.T) {
		var inputs struct {
			ExpressionDay   Time `meta_round:"day:down"`
			ExpressionHour  Time `meta_round:"hour:up"`
			ExpressionWeek  Time `meta_round:"week:down"`
			ExpressionMonth Time `meta_round:"day:down"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"expression_day":   {"2_days_ago"},
			"expression_hour":  {"3_hours_ago"},
			"expression_week":  {"1_week_ago"},
			"expression_month": {"1_month_ago"},
		})

		assertEqual(t, e, ErrorHash(nil))

		// Verify that expressions are rounded correctly
		// Day should be rounded down to the start of the day
		assert(t, inputs.ExpressionDay.Val.Hour() == 0 && inputs.ExpressionDay.Val.Minute() == 0 && inputs.ExpressionDay.Val.Second() == 0)

		// Hour should be rounded up to the next hour
		assert(t, inputs.ExpressionHour.Val.Minute() == 0 && inputs.ExpressionHour.Val.Second() == 0)

		// Week should be rounded down to the start of the week (Monday)
		assert(t, inputs.ExpressionWeek.Val.Weekday() == time.Monday)
		assert(t, inputs.ExpressionWeek.Val.Hour() == 0 && inputs.ExpressionWeek.Val.Minute() == 0 && inputs.ExpressionWeek.Val.Second() == 0)

		// Month should be rounded down to start of day
		assert(t, inputs.ExpressionMonth.Val.Hour() == 0)
		assert(t, inputs.ExpressionMonth.Val.Minute() == 0)
		assert(t, inputs.ExpressionMonth.Val.Second() == 0)
		assert(t, inputs.ExpressionMonth.Val.Nanosecond() == 0)
	})

	// Test non-expression rounding
	t.Run("non-expression rounding", func(t *testing.T) {
		var inputs struct {
			NonExpression Time `meta_round:"day:down"`
		}

		e := NewDecoder(&inputs).DecodeValues(&inputs, url.Values{
			"non_expression": {"2024-01-15T14:30:45Z"},
		})

		assertEqual(t, e, ErrorHash(nil))

		// Should use meta_round (day:down)
		expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		assert(t, inputs.NonExpression.Val.Equal(expected))
	})
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

func TestTimeZeroValues(t *testing.T) {
	// Test zero values with expressions
	var zeroInputs struct {
		A Time
		B Time
	}

	e := NewDecoder(&zeroInputs).DecodeValues(&zeroInputs, url.Values{
		"a": {"0_days_ago"},
		"b": {"0_hours_from_now"},
	})

	assertEqual(t, e, ErrorHash(nil))

	// Should handle zero values gracefully - no rounding by default
	// 0_days_ago should be exact time (no rounding)
	expectedA := time.Now()
	assert(t, zeroInputs.A.Val.Sub(expectedA) < time.Second)

	// 0_hours_from_now should be now (no rounding)
	expectedB := time.Now()
	assert(t, zeroInputs.B.Val.Sub(expectedB) < time.Second)
}

// Category 7: Combining Features - Missing Tests

func TestTimeComplexFeatureCombination(t *testing.T) {
	// Test complex combination of multiple features
	var complexInputs struct {
		A Time `meta_min:"!1_day_ago" meta_max:"!1_day_from_now" meta_round:"hour:up"`
		B Time `meta_min:"2024-01-01T00:00:00Z" meta_max:"2024-12-31T23:59:59Z" meta_round:"day:down"`
	}

	e := NewDecoder(&complexInputs).DecodeValues(&complexInputs, url.Values{
		"a": {"2_hours_ago"},          // Should be rounded up to end of hour, and pass min/max validation
		"b": {"2024-06-15T14:30:45Z"}, // Should be rounded down to start of day, and pass min/max validation
	})

	assertEqual(t, e, ErrorHash(nil))

	// A should be rounded up to end of hour and should be between the exclusive boundaries
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
