package meta

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//
// Core Types and Structures
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
	// Format specifies the time formats to use for parsing time strings.
	// Configured via meta_format tag.
	// Can be predefined formats (RFC3339, DateOnly, etc.) or custom Go time layouts.
	// Multiple formats can be specified as comma-separated values.
	// Default: [RFC3339, "expression"] (supports RFC3339 format and time expressions like "now", "3_days_ago")
	// Examples: `meta_format:"RFC3339"`, `meta_format:"DateOnly"`, `meta_format:"1/2/2006"`, `meta_format:"RFC3339,2006-01-02 15:04:05"`
	Format []string
	// MinDate sets the minimum allowed date/time value.
	// Configured via meta_min tag.
	// Can be an absolute date/time string or a relative expression.
	// Supports exclusive boundaries with "!" prefix.
	// Default: nil (no minimum limit)
	// Examples: `meta_min:"2024-01-01"`, `meta_min:"3_days_ago"`, `meta_min:"!1_day_ago"`
	// NOTE: configured/default rounding is applied to the min date
	MinDate *dateLimit
	// MaxDate sets the maximum allowed date/time value.
	// Configured via meta_max tag.
	// Can be an absolute date/time string or a relative expression.
	// Supports exclusive boundaries with "!" prefix.
	// Default: nil (no maximum limit)
	// Examples: `meta_max:"2024-12-31"`, `meta_max:"now"`, `meta_max:"!1_day_from_now"`
	// NOTE: configured/default rounding is applied to the max date
	MaxDate *dateLimit
	// Round configures time rounding behavior for all time values (absolute and expressions).
	// Configured via meta_round tag.
	// Format: "unit:direction" where unit can be year, month, week, day, hour, minute, second, nanosecond,
	// or day names (sunday, monday, etc.), and direction can be up, down, or nearest.
	// Default: nil (no rounding applied)
	// Direction defaults to "down" if not specified (e.g., `meta_round:"day"` is equivalent to `meta_round:"day:down"`)
	// Examples: `meta_round:"day:down"`, `meta_round:"hour:up"`, `meta_round:"monday:nearest"`, `meta_round:"week:down"`
	Round *roundConfig
}

type roundConfig struct {
	unit      RoundingUnit
	direction RoundingDirection
}

type dateLimit struct {
	value      time.Time
	isAbsolute bool
	exclusive  bool
	raw        string
}

type timeComponent struct {
	year, month, day, hour, minute, second, nanosecond int
}

type expressionParser struct {
	*regexp.Regexp
	Parse func([]string) (time.Time, bool)
}

// unitRoundingMethods defines how to round standard units
type unitRoundingMethods struct {
	upFunc   func(timeComponent) timeComponent
	zeroFunc func(*timeComponent)
}

//
// Constants
//

const (
	// TimeComparisonTolerance allows for some tolerance in time comparisons
	// to handle time expression resolution differences
	TimeComparisonTolerance = time.Millisecond * 250 // 1/4 second
)

//
// Type Definitions for Rounding
//

// RoundingUnit represents the unit for time rounding
type RoundingUnit string

// RoundingDirection represents the direction for time rounding
type RoundingDirection string

const (
	// Rounding units
	UnitYear   RoundingUnit = "year"
	UnitMonth  RoundingUnit = "month"
	UnitWeek   RoundingUnit = "week"
	UnitDay    RoundingUnit = "day"
	UnitHour   RoundingUnit = "hour"
	UnitMinute RoundingUnit = "minute"
	UnitSecond RoundingUnit = "second"

	// Rounding directions
	DirectionUp      RoundingDirection = "up"
	DirectionDown    RoundingDirection = "down"
	DirectionNearest RoundingDirection = "nearest"
)

//
// Global Variables and Maps
//

// timeFormatMap maps predefined format names to their constants
var timeFormatMap = map[string]string{
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
	"DateTime":    time.DateTime,
	"DateOnly":    time.DateOnly,
	"TimeOnly":    time.TimeOnly,
}

// dayNameToWeekday maps day names to time.Weekday values
var dayNameToWeekday = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
	// Abbreviations
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

// dayNames provides consistent day name strings
var dayNames = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

// Valid traditional units for rounding
var validUnits = map[string]RoundingUnit{
	string(UnitYear):   UnitYear,
	string(UnitMonth):  UnitMonth,
	string(UnitWeek):   UnitWeek,
	string(UnitDay):    UnitDay,
	string(UnitHour):   UnitHour,
	string(UnitMinute): UnitMinute,
	string(UnitSecond): UnitSecond,
}

// Valid rounding directions
var validDirections = map[string]RoundingDirection{
	string(DirectionUp):      DirectionUp,
	string(DirectionDown):    DirectionDown,
	string(DirectionNearest): DirectionNearest,
}

// timeUnitDuration maps time unit strings to their corresponding durations
var timeUnitDuration = map[RoundingUnit]time.Duration{
	UnitHour:     time.Hour,
	UnitMinute:   time.Minute,
	UnitSecond:   time.Second,
	"nanosecond": time.Nanosecond,
}

// timeUnitAddDate maps time unit strings to their AddDate parameters
var timeUnitAddDate = map[RoundingUnit]func(int) (int, int, int){
	UnitYear:  func(delta int) (int, int, int) { return delta, 0, 0 },
	UnitMonth: func(delta int) (int, int, int) { return 0, delta, 0 },
	UnitWeek:  func(delta int) (int, int, int) { return 0, 0, delta * 7 },
	UnitDay:   func(delta int) (int, int, int) { return 0, 0, delta },
}

// absoluteTimeExpressions maps expression names to their time values
var absoluteTimeExpressions = map[string]func() time.Time{
	"now": func() time.Time { return time.Now() },
	"today": func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	},
	"yesterday": func() time.Time {
		now := time.Now()
		base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return base.AddDate(0, 0, -1)
	},
	"tomorrow": func() time.Time {
		now := time.Now()
		base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return base.AddDate(0, 0, 1)
	},
}

// unitRoundingMap maps standard units to their rounding configurations
var unitRoundingMap = map[RoundingUnit]unitRoundingMethods{
	UnitYear: {
		upFunc: func(comp timeComponent) timeComponent {
			comp.year++
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.month, comp.day, comp.hour, comp.minute, comp.second, comp.nanosecond = 1, 1, 0, 0, 0, 0
		},
	},
	UnitMonth: {
		upFunc: func(comp timeComponent) timeComponent {
			// First zero the components, then add 1 month
			comp.day, comp.hour, comp.minute, comp.second, comp.nanosecond = 1, 0, 0, 0, 0
			nextMonth := time.Date(comp.year, time.Month(comp.month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
			comp.year, comp.month = nextMonth.Year(), int(nextMonth.Month())
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.day, comp.hour, comp.minute, comp.second, comp.nanosecond = 1, 0, 0, 0, 0
		},
	},
	UnitDay: {
		upFunc: func(comp timeComponent) timeComponent {
			comp.day++
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.hour, comp.minute, comp.second, comp.nanosecond = 0, 0, 0, 0
		},
	},
	UnitHour: {
		upFunc: func(comp timeComponent) timeComponent {
			comp.hour++
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.minute, comp.second, comp.nanosecond = 0, 0, 0
		},
	},
	UnitMinute: {
		upFunc: func(comp timeComponent) timeComponent {
			comp.minute++
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.second, comp.nanosecond = 0, 0
		},
	},
	UnitSecond: {
		upFunc: func(comp timeComponent) timeComponent {
			comp.second++
			return comp
		},
		zeroFunc: func(comp *timeComponent) {
			comp.nanosecond = 0
		},
	},
}

// boundaryChecks maps unit names to their boundary checking functions
var boundaryChecks = map[string]func(time.Time) bool{
	"year": func(t time.Time) bool {
		return t.YearDay() == 1 && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
	},
	"month": func(t time.Time) bool {
		return t.Day() == 1 && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
	},
	"day": func(t time.Time) bool {
		return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
	},
	"hour": func(t time.Time) bool {
		return t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
	},
	"minute": func(t time.Time) bool {
		return t.Second() == 0 && t.Nanosecond() == 0
	},
	"second": func(t time.Time) bool {
		return t.Nanosecond() == 0
	},
}

// Pre-compile regex patterns for better performance
var (
	relativeTimeRegex = regexp.MustCompile(`^(\d+)_(year|month|week|day|hour|minute|second|nanosecond)s?_(ago|from_now)$`)
	absoluteTimeRegex = regexp.MustCompile(`^(now|today|yesterday|tomorrow)$`)
)

var timeExpressionParsers = []expressionParser{
	{
		Regexp: relativeTimeRegex,
		Parse:  parseRelativeTimeExpression,
	},
	{
		Regexp: absoluteTimeRegex,
		Parse:  parseAbsoluteTimeExpression,
	},
}

//
// Constructor
//

func NewTime(t time.Time) Time {
	return Time{t, Nullity{false}, Presence{true}, ""}
}

//
// Core Time Methods
//

func (t *Time) ParseOptions(tag reflect.StructTag) interface{} {
	opts := &TimeOptions{
		Required:     tag.Get("meta_required") == "true",
		DiscardBlank: tag.Get("meta_discard_blank") != "false",
		Null:         tag.Get("meta_null") == "true",
		Format:       parseFormats(tag.Get("meta_format")),
	}

	opts.MinDate = newDateLimit(tag.Get("meta_min"), opts)
	opts.MaxDate = newDateLimit(tag.Get("meta_max"), opts)

	// Parse rounding configuration
	if roundTag := tag.Get("meta_round"); roundTag != "" {
		opts.Round = parseRoundConfig(roundTag)
	}

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
			return t.handleZeroTimeValue(opts)
		}
		return t.handleNonZeroTimeValue(value, opts)
	case string:
		return t.FormValue(value, options)
	}

	return ErrTime
}

func (t *Time) FormValue(value string, options interface{}) Errorable {
	opts := options.(*TimeOptions)

	if value == "" {
		return t.handleEmptyValue(opts)
	}

	return t.parseTimeValue(value, opts)
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

//
// Value Handling Methods - Consolidated
//

func (t *Time) handleZeroTimeValue(opts *TimeOptions) Errorable {
	return t.handleEmptyValue(opts)
}

func (t *Time) handleNonZeroTimeValue(value time.Time, opts *TimeOptions) Errorable {
	t.Present = true
	t.Val = value
	t.applyRounding(opts)
	return t.assertTimeRange(opts)
}

func (t *Time) handleEmptyValue(opts *TimeOptions) Errorable {
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

//
// Time Parsing Methods
//

func (t *Time) parseTimeValue(value string, opts *TimeOptions) Errorable {
	for _, format := range opts.Format {
		switch format {
		case "expression":
			if t.parseTimeExpression(value, opts) {
				return t.assertTimeRange(opts)
			}
		default:
			if t.parseTimeFormat(value, format, opts) {
				return t.assertTimeRange(opts)
			}
		}
	}
	return ErrTime
}

func (t *Time) parseTimeExpression(value string, opts *TimeOptions) bool {
	if v := resolveTimeExpression(value); v != nil {
		t.Val = *v
		t.Present = true
		t.applyRounding(opts)
		return true
	}
	return false
}

func (t *Time) parseTimeFormat(value string, format string, opts *TimeOptions) bool {
	if v, err := time.Parse(format, value); err == nil {
		t.Val = v
		t.Present = true
		t.applyRounding(opts)
		return true
	}
	return false
}

//
// Format Parsing
//

func parseFormats(formatTag string) []string {
	if formatTag == "" {
		return []string{time.RFC3339, "expression"}
	}

	formats := strings.Split(formatTag, ",")
	result := make([]string, 0, len(formats))

	for _, format := range formats {
		format = strings.TrimSpace(format)
		if predefined, exists := timeFormatMap[format]; exists {
			result = append(result, predefined)
		} else {
			result = append(result, format)
		}
	}

	return result
}

//
// Expression Parsing
//

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

func parseRelativeTimeExpression(matches []string) (time.Time, bool) {
	delta, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, false
	}

	unit := RoundingUnit(matches[2])
	isAgo := matches[3] == "ago"

	var result time.Time

	// Handle AddDate units (year, month, week, day)
	if addDateFunc, exists := timeUnitAddDate[unit]; exists {
		years, months, days := addDateFunc(delta)
		if isAgo {
			result = time.Now().AddDate(-years, -months, -days)
		} else {
			result = time.Now().AddDate(years, months, days)
		}
	} else if duration, exists := timeUnitDuration[unit]; exists {
		// Handle Duration units (hour, minute, second, nanosecond)
		if isAgo {
			result = time.Now().Add(-time.Duration(delta) * duration)
		} else {
			result = time.Now().Add(time.Duration(delta) * duration)
		}
	} else {
		return time.Time{}, false
	}

	return result, true
}

func parseAbsoluteTimeExpression(matches []string) (time.Time, bool) {
	if fn, exists := absoluteTimeExpressions[matches[1]]; exists {
		return fn(), true
	}
	return time.Time{}, false
}

//
// Tag Parsing
//

func parseRoundConfig(roundTag string) *roundConfig {
	roundTag = strings.ToLower(roundTag)
	parts := strings.Split(roundTag, ":")
	unitStr := strings.TrimSpace(parts[0])

	// Parse direction (default to "down")
	direction := DirectionDown
	if len(parts) > 1 {
		parsedDirection := strings.TrimSpace(parts[1])
		if validDir, exists := validDirections[parsedDirection]; exists {
			direction = validDir
		}
	}

	// Check if unit is a day name
	if normalizedDay, isDay := normalizeDayName(unitStr); isDay {
		return &roundConfig{unit: RoundingUnit(normalizedDay), direction: direction}
	}

	// Validate traditional units
	if validUnit, exists := validUnits[unitStr]; exists {
		return &roundConfig{unit: validUnit, direction: direction}
	}

	return nil // ignore invalid units
}

func normalizeDayName(dayName string) (string, bool) {
	if dayName == "" {
		return "", false
	}

	dayName = strings.ToLower(strings.TrimSpace(dayName))
	if weekday, ok := dayNameToWeekday[dayName]; ok {
		// Return the full day name for consistency
		return dayNames[weekday], true
	}
	return "", false
}

//
// Rounding Methods
//

func (t *Time) applyRounding(opts *TimeOptions) {
	if opts.Round == nil {
		return
	}

	switch opts.Round.direction {
	case DirectionDown:
		t.Val = t.roundDown(opts.Round.unit)
	case DirectionUp:
		t.Val = t.roundUp(opts.Round.unit)
	case DirectionNearest:
		t.Val = t.roundNearest(opts.Round.unit)
	}
}

func (t *Time) roundDown(unit RoundingUnit) time.Time {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond:
		return t.roundStandardDown(unit)
	case UnitWeek:
		return t.roundDownToWeekday(time.Monday) // Default to Monday
	default:
		return t.roundDayNameDown(unit)
	}
}

func (t *Time) roundUp(unit RoundingUnit) time.Time {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond:
		return t.roundStandardUp(unit)
	case UnitWeek:
		down := t.roundDownToWeekday(time.Monday)
		return down.AddDate(0, 0, 7)
	default:
		return t.roundDayNameUp(unit)
	}
}

func (t *Time) roundNearest(unit RoundingUnit) time.Time {
	down := t.roundDown(unit)
	up := t.roundUp(unit)

	// If already at boundary, return current value
	if down.Equal(up) {
		return down
	}

	// Find the midpoint
	midpoint := down.Add(up.Sub(down) / 2)

	if t.Val.Before(midpoint) || t.Val.Equal(midpoint) {
		return down
	}
	return up
}

//
// Standard Rounding Methods - Simplified
//

func (t *Time) roundStandardDown(unit RoundingUnit) time.Time {
	comp := t.roundComponents(unit, DirectionDown)
	return t.createTimeFromComponents(comp)
}

func (t *Time) roundStandardUp(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if t.isAtBoundary(unitStr) {
		return t.Val
	}

	comp := t.roundComponents(unit, DirectionUp)
	return t.createTimeFromComponents(comp)
}

func (t *Time) roundComponents(unit RoundingUnit, direction RoundingDirection) timeComponent {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond:
		return t.roundStandardUnit(unit, direction)
	default:
		return t.getTimeComponents()
	}
}

func (t *Time) roundStandardUnit(unit RoundingUnit, direction RoundingDirection) timeComponent {
	comp := t.getTimeComponents()
	unitStr := string(unit)

	if config, exists := unitRoundingMap[unit]; exists {
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp = config.upFunc(comp)
		}
		config.zeroFunc(&comp)
	}

	return comp
}

//
// Day Name Rounding Methods
//

func (t *Time) roundDayNameDown(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if weekday, ok := dayNameToWeekday[unitStr]; ok {
		return t.roundDownToWeekday(weekday)
	}
	return t.Val
}

func (t *Time) roundDayNameUp(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if weekday, ok := dayNameToWeekday[unitStr]; ok {
		down := t.roundDownToWeekday(weekday)
		if t.Val.Equal(down) {
			return down
		}
		return down.AddDate(0, 0, 7)
	}
	return t.Val
}

func (t *Time) roundDownToWeekday(targetDay time.Weekday) time.Time {
	daysSinceStart := int(t.Val.Weekday() - targetDay)
	if daysSinceStart < 0 {
		daysSinceStart += 7
	}
	comp := timeComponent{
		year:  t.Val.Year(),
		month: int(t.Val.Month()),
		day:   t.Val.Day() - daysSinceStart,
		hour:  0, minute: 0, second: 0, nanosecond: 0,
	}
	return t.createTimeFromComponents(comp)
}

//
// Time Component Methods
//

func (t *Time) getTimeComponents() timeComponent {
	return timeComponent{
		year:       t.Val.Year(),
		month:      int(t.Val.Month()),
		day:        t.Val.Day(),
		hour:       t.Val.Hour(),
		minute:     t.Val.Minute(),
		second:     t.Val.Second(),
		nanosecond: t.Val.Nanosecond(),
	}
}

func (t *Time) createTimeFromComponents(comp timeComponent) time.Time {
	return time.Date(comp.year, time.Month(comp.month), comp.day,
		comp.hour, comp.minute, comp.second, comp.nanosecond, t.Val.Location())
}

//
// Boundary Checking Methods - Simplified
//

func (t *Time) isAtBoundary(unit string) bool {
	if check, exists := boundaryChecks[unit]; exists {
		return check(t.Val)
	}
	return t.isAtDayNameBoundary(unit)
}

func (t *Time) isAtDayNameBoundary(unit string) bool {
	weekday, ok := dayNameToWeekday[unit]
	if !ok {
		return false
	}
	down := t.roundDownToWeekday(weekday)
	return t.Val.Equal(down)
}

//
// Date Limit Methods
//

func newDateLimit(raw string, opts *TimeOptions) *dateLimit {
	if raw == "" {
		return nil
	}

	value, exclusive := parseExclusivePrefix(raw)

	// Always try to resolve expressions for min/max dates
	if v := resolveTimeExpression(value); v != nil {
		return &dateLimit{raw: value, exclusive: exclusive}
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
		// Apply the same rounding to expression-based boundaries
		if opts.Round != nil {
			tempTime := Time{Val: *v}
			tempTime.applyRounding(opts)
			return tempTime.Val
		}
		return *v
	}

	return time.Now()
}

func parseExclusivePrefix(raw string) (string, bool) {
	if strings.HasPrefix(raw, "!") {
		return strings.TrimPrefix(raw, "!"), true
	}
	return raw, false
}

//
// Range Validation Methods - Simplified
//

func (t *Time) assertTimeRange(opts *TimeOptions) Errorable {
	if opts.MinDate != nil {
		minDate := opts.MinDate.Value(opts)
		if !t.validateTimeBoundary(minDate, opts.MinDate.exclusive, true) {
			return ErrMin
		}
	}
	if opts.MaxDate != nil {
		maxDate := opts.MaxDate.Value(opts)
		if !t.validateTimeBoundary(maxDate, opts.MaxDate.exclusive, false) {
			return ErrMax
		}
	}
	return nil
}

func (t *Time) validateTimeBoundary(boundaryTime time.Time, exclusive bool, isMin bool) bool {
	if exclusive {
		return t.validateExclusiveBoundary(boundaryTime, isMin)
	}
	return t.validateInclusiveBoundary(boundaryTime, isMin)
}

func (t *Time) validateExclusiveBoundary(boundaryTime time.Time, isMin bool) bool {
	if isMin {
		// Exclusive min: value must be strictly after boundary
		return t.Val.After(boundaryTime.Add(TimeComparisonTolerance))
	}
	// Exclusive max: value must be strictly before boundary
	return t.Val.Before(boundaryTime.Add(-TimeComparisonTolerance))
}

func (t *Time) validateInclusiveBoundary(boundaryTime time.Time, isMin bool) bool {
	if isMin {
		// Inclusive min: value must be greater than or equal to boundary
		return !t.Val.Before(boundaryTime.Add(-TimeComparisonTolerance))
	}
	// Inclusive max: value must be less than or equal to boundary
	return !t.Val.After(boundaryTime.Add(TimeComparisonTolerance))
}
