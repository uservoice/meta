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

func assertTimeInRange(t *testing.T, value, start, end time.Time) {
	assert(t, value.Equal(start) || value.After(start))
	assert(t, value.Equal(end) || value.Before(end))
}

func TestTimeExpressions(t *testing.T) {
	for expression, assertion := range map[string]func(output, before, after time.Time){
		// Simple keywords
		"now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before, after)
		},
		"today": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day(), 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day(), 0, 0, 0, 0, time.UTC),
			)
		},
		"yesterday": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day()-1, 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day()-1, 0, 0, 0, 0, time.UTC),
			)
		},
		"tomorrow": func(output, before, after time.Time) {
			assertTimeInRange(t, output,
				time.Date(before.Year(), before.Month(), before.Day()+1, 0, 0, 0, 0, time.UTC),
				time.Date(after.Year(), after.Month(), after.Day()+1, 0, 0, 0, 0, time.UTC),
			)
		},
		// Past expressions (<n>_<unit>_ago)
		"99_nanoseconds_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-99*time.Nanosecond), after.Add(-99*time.Nanosecond))
		},
		"31_seconds_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-31*time.Second), after.Add(-31*time.Second))
		},
		"1_minute_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-time.Minute), after.Add(-time.Minute))
		},
		"48_hours_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(-48*time.Hour), after.Add(-48*time.Hour))
		},
		"1_day_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -1), after.AddDate(0, 0, -1))
		},
		"5_days_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -5), after.AddDate(0, 0, -5))
		},
		"3_weeks_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, -21), after.AddDate(0, 0, -21))
		},
		"2_months_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, -2, 0), after.AddDate(0, -2, 0))
		},
		"4_years_ago": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(-4, 0, 0), after.AddDate(-4, 0, 0))
		},
		// Future expressions (<n>_<unit>_from_now)
		"99_nanoseconds_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(99*time.Nanosecond), after.Add(99*time.Nanosecond))
		},
		"31_seconds_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(31*time.Second), after.Add(31*time.Second))
		},
		"1_minute_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(time.Minute), after.Add(time.Minute))
		},
		"48_hours_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.Add(48*time.Hour), after.Add(48*time.Hour))
		},
		"1_day_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 1), after.AddDate(0, 0, 1))
		},
		"5_days_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 5), after.AddDate(0, 0, 5))
		},
		"3_weeks_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 0, 21), after.AddDate(0, 0, 21))
		},
		"2_months_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(0, 2, 0), after.AddDate(0, 2, 0))
		},
		"4_years_from_now": func(output, before, after time.Time) {
			assertTimeInRange(t, output, before.AddDate(4, 0, 0), after.AddDate(4, 0, 0))
		},
	} {
		var inputs withTime

		before := time.Now()
		e := withTimeDecoder.DecodeValues(&inputs, url.Values{"a": {expression}})
		after := time.Now()
		assertEqual(t, e, ErrorHash(nil))
		assertEqual(t, inputs.A.Present, true)
		assertion(inputs.A.Val.UTC(), before.UTC(), after.UTC())
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
