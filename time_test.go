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

func TestTimeRoundingInvalidUnit(t *testing.T) {
	// Test that invalid rounding units are ignored
	var invalidInputs struct {
		A Time `meta_round:"invalid:down"`
	}

	e := NewDecoder(&invalidInputs).DecodeValues(&invalidInputs, url.Values{
		"a": {"2024-01-15T14:30:45Z"},
	})

	assertEqual(t, e, ErrorHash(nil))
	// Should not apply any rounding, so the time should be exactly as parsed
	assert(t, invalidInputs.A.Val.Equal(time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)))
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
