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
	Round        *roundConfig
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

func NewTime(t time.Time) Time {
	return Time{t, Nullity{false}, Presence{true}, ""}
}

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

// Day name mappings
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

// RoundingUnit represents the unit for time rounding
type RoundingUnit string

// RoundingDirection represents the direction for time rounding
type RoundingDirection string

const (
	// Rounding units
	UnitYear       RoundingUnit = "year"
	UnitMonth      RoundingUnit = "month"
	UnitWeek       RoundingUnit = "week"
	UnitDay        RoundingUnit = "day"
	UnitHour       RoundingUnit = "hour"
	UnitMinute     RoundingUnit = "minute"
	UnitSecond     RoundingUnit = "second"
	UnitNanosecond RoundingUnit = "nanosecond"

	// Rounding directions
	DirectionUp      RoundingDirection = "up"
	DirectionDown    RoundingDirection = "down"
	DirectionNearest RoundingDirection = "nearest"
)

// Valid traditional units for rounding
var validUnits = map[string]RoundingUnit{
	string(UnitYear):       UnitYear,
	string(UnitMonth):      UnitMonth,
	string(UnitWeek):       UnitWeek,
	string(UnitDay):        UnitDay,
	string(UnitHour):       UnitHour,
	string(UnitMinute):     UnitMinute,
	string(UnitSecond):     UnitSecond,
	string(UnitNanosecond): UnitNanosecond,
}

// Valid rounding directions
var validDirections = map[string]RoundingDirection{
	string(DirectionUp):      DirectionUp,
	string(DirectionDown):    DirectionDown,
	string(DirectionNearest): DirectionNearest,
}

// DayName represents a day of the week
type DayName string

const (
	DaySunday    DayName = "sunday"
	DayMonday    DayName = "monday"
	DayTuesday   DayName = "tuesday"
	DayWednesday DayName = "wednesday"
	DayThursday  DayName = "thursday"
	DayFriday    DayName = "friday"
	DaySaturday  DayName = "saturday"
)

// dayNames provides consistent day name strings
var dayNames = []DayName{DaySunday, DayMonday, DayTuesday, DayWednesday, DayThursday, DayFriday, DaySaturday}

// parseFormats extracts and validates time format strings from the tag
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

// Helper function to normalize day name
func normalizeDayName(dayName string) (DayName, bool) {
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

// TimeComponent represents the components of a time that can be rounded
type timeComponent struct {
	year, month, day, hour, minute, second, nanosecond int
}

// getTimeComponents extracts the current time components
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

// createTimeFromComponents creates a time.Time from components
func (t *Time) createTimeFromComponents(comp timeComponent) time.Time {
	return time.Date(comp.year, time.Month(comp.month), comp.day,
		comp.hour, comp.minute, comp.second, comp.nanosecond, t.Val.Location())
}

// zeroComponentsBelow sets all time components below the specified unit to zero
func (comp *timeComponent) zeroComponentsBelow(unit RoundingUnit) {
	switch unit {
	case UnitYear:
		comp.month, comp.day, comp.hour, comp.minute, comp.second, comp.nanosecond = 1, 1, 0, 0, 0, 0
	case UnitMonth:
		comp.day, comp.hour, comp.minute, comp.second, comp.nanosecond = 1, 0, 0, 0, 0
	case UnitDay:
		comp.hour, comp.minute, comp.second, comp.nanosecond = 0, 0, 0, 0
	case UnitHour:
		comp.minute, comp.second, comp.nanosecond = 0, 0, 0
	case UnitMinute:
		comp.second, comp.nanosecond = 0, 0
	case UnitSecond:
		comp.nanosecond = 0
	}
}

// roundStandardUnit handles rounding for standard time units
func (t *Time) roundStandardUnit(unit RoundingUnit, direction RoundingDirection) timeComponent {
	comp := t.getTimeComponents()
	unitStr := string(unit)

	switch unit {
	case UnitYear:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp.year++
		}
		comp.zeroComponentsBelow(unit)

	case UnitMonth:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			nextMonth := t.Val.AddDate(0, 1, 0)
			comp.year, comp.month = nextMonth.Year(), int(nextMonth.Month())
		}
		comp.zeroComponentsBelow(unit)

	case UnitDay:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp.day++
		}
		comp.zeroComponentsBelow(unit)

	case UnitHour:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp.hour++
		}
		comp.zeroComponentsBelow(unit)

	case UnitMinute:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp.minute++
		}
		comp.zeroComponentsBelow(unit)

	case UnitSecond:
		if direction == DirectionUp && !t.isAtBoundary(unitStr) {
			comp.second++
		}
		comp.zeroComponentsBelow(unit)

	case UnitNanosecond:
		// nanosecond is the smallest unit, no rounding needed
		return comp
	}

	return comp
}

// roundComponents calculates the rounded components for a given unit and direction
func (t *Time) roundComponents(unit RoundingUnit, direction RoundingDirection) timeComponent {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond, UnitNanosecond:
		return t.roundStandardUnit(unit, direction)
	case UnitWeek:
		// Handle week rounding separately as it's more complex
		return t.getTimeComponents()
	default:
		return t.getTimeComponents()
	}
}

// Helper function to check if time components are zero from a given level down
func (t *Time) isZeroFrom(level string) bool {
	switch level {
	case "hour":
		return t.Val.Hour() == 0 && t.Val.Minute() == 0 && t.Val.Second() == 0 && t.Val.Nanosecond() == 0
	case "minute":
		return t.Val.Minute() == 0 && t.Val.Second() == 0 && t.Val.Nanosecond() == 0
	case "second":
		return t.Val.Second() == 0 && t.Val.Nanosecond() == 0
	case "nanosecond":
		return t.Val.Nanosecond() == 0
	default:
		return false
	}
}

// isAtStandardBoundary checks if time is at a boundary for standard units
func (t *Time) isAtStandardBoundary(unit string) bool {
	switch unit {
	case "year":
		return t.Val.YearDay() == 1 && t.isZeroFrom("hour")
	case "month":
		return t.Val.Day() == 1 && t.isZeroFrom("hour")
	case "day":
		return t.isZeroFrom("hour")
	case "hour":
		return t.isZeroFrom("minute")
	case "minute":
		return t.isZeroFrom("second")
	case "second":
		return t.isZeroFrom("nanosecond")
	default:
		return false
	}
}

// isAtDayNameBoundary checks if time is at a boundary for day names
func (t *Time) isAtDayNameBoundary(unit string) bool {
	weekday, ok := dayNameToWeekday[unit]
	if !ok {
		return false
	}
	down := t.roundDownToWeekday(weekday)
	return t.Val.Equal(down)
}

// Helper function to check if time is already at a boundary
func (t *Time) isAtBoundary(unit string) bool {
	if t.isAtStandardBoundary(unit) {
		return true
	}
	return t.isAtDayNameBoundary(unit)
}

// Helper function to round down to a specific weekday
func (t *Time) roundDownToWeekday(targetDay time.Weekday) time.Time {
	daysSinceStart := int(t.Val.Weekday() - targetDay)
	if daysSinceStart < 0 {
		daysSinceStart += DaysInWeek
	}
	comp := timeComponent{
		year:  t.Val.Year(),
		month: int(t.Val.Month()),
		day:   t.Val.Day() - daysSinceStart,
		hour:  0, minute: 0, second: 0, nanosecond: 0,
	}
	return t.createTimeFromComponents(comp)
}

func (t *Time) roundDown(unit RoundingUnit) time.Time {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond, UnitNanosecond:
		return t.roundStandardDown(unit)
	case UnitWeek:
		return t.roundDownToWeekday(time.Monday) // Default to Monday
	default:
		return t.roundDayNameDown(unit)
	}
}

// roundStandardDown handles rounding down for standard units
func (t *Time) roundStandardDown(unit RoundingUnit) time.Time {
	comp := t.roundComponents(unit, DirectionDown)
	if unit == UnitNanosecond {
		return t.Val.Truncate(time.Nanosecond)
	}
	return t.createTimeFromComponents(comp)
}

// roundDayNameDown handles rounding down for day names
func (t *Time) roundDayNameDown(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if weekday, ok := dayNameToWeekday[unitStr]; ok {
		return t.roundDownToWeekday(weekday)
	}
	return t.Val
}

func (t *Time) roundUp(unit RoundingUnit) time.Time {
	switch unit {
	case UnitYear, UnitMonth, UnitDay, UnitHour, UnitMinute, UnitSecond, UnitNanosecond:
		return t.roundStandardUp(unit)
	case UnitWeek:
		down := t.roundDownToWeekday(time.Monday)
		return down.AddDate(0, 0, DaysInWeek)
	default:
		return t.roundDayNameUp(unit)
	}
}

// roundStandardUp handles rounding up for standard units
func (t *Time) roundStandardUp(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if t.isAtBoundary(unitStr) {
		return t.Val
	}

	comp := t.roundComponents(unit, DirectionUp)
	if unit == UnitNanosecond {
		return t.Val // nanosecond is the smallest unit
	}
	return t.createTimeFromComponents(comp)
}

// roundDayNameUp handles rounding up for day names
func (t *Time) roundDayNameUp(unit RoundingUnit) time.Time {
	unitStr := string(unit)
	if weekday, ok := dayNameToWeekday[unitStr]; ok {
		down := t.roundDownToWeekday(weekday)
		if t.Val.Equal(down) {
			return down
		}
		return down.AddDate(0, 0, DaysInWeek)
	}
	return t.Val
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

// handleZeroTimeValue processes zero time values according to options
func (t *Time) handleZeroTimeValue(opts *TimeOptions) Errorable {
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

// handleNonZeroTimeValue processes non-zero time values
func (t *Time) handleNonZeroTimeValue(value time.Time, opts *TimeOptions) Errorable {
	t.Present = true
	t.Val = value
	t.applyRounding(opts)
	return t.assertTimeRange(opts)
}

// parseExclusivePrefix extracts the exclusive prefix and returns the cleaned value
func parseExclusivePrefix(raw string) (string, bool) {
	if strings.HasPrefix(raw, "!") {
		return strings.TrimPrefix(raw, "!"), true
	}
	return raw, false
}

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
		return *v
	}

	return time.Now()
}

const (
	// DaysInWeek represents the number of days in a week
	DaysInWeek = 7

	// TimeComparisonTolerance allows for some tolerance in time comparisons
	// to handle time expression resolution differences
	TimeComparisonTolerance = time.Millisecond * 250 // 1/4 second

	// HoursInDay represents the number of hours in a day
	HoursInDay = 24
)

// validateExclusiveBoundary handles exclusive boundary validation
func (t *Time) validateExclusiveBoundary(boundaryTime time.Time, isMin bool) bool {
	if isMin {
		// Exclusive min: value must be strictly after boundary
		return t.Val.After(boundaryTime.Add(TimeComparisonTolerance))
	}
	// Exclusive max: value must be strictly before boundary
	return t.Val.Before(boundaryTime.Add(-TimeComparisonTolerance))
}

// validateInclusiveBoundary handles inclusive boundary validation
func (t *Time) validateInclusiveBoundary(boundaryTime time.Time, isMin bool) bool {
	if isMin {
		// Inclusive min: value must be greater than or equal to boundary
		return !t.Val.Before(boundaryTime.Add(-TimeComparisonTolerance))
	}
	// Inclusive max: value must be less than or equal to boundary
	return !t.Val.After(boundaryTime.Add(TimeComparisonTolerance))
}

// validateTimeBoundary checks if a time value is within the specified boundary
func (t *Time) validateTimeBoundary(boundaryTime time.Time, exclusive bool, isMin bool) bool {
	if exclusive {
		return t.validateExclusiveBoundary(boundaryTime, isMin)
	}
	return t.validateInclusiveBoundary(boundaryTime, isMin)
}

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

// TimeUnit represents a time unit for expressions
type TimeUnit string

const (
	ExprUnitYear       TimeUnit = "year"
	ExprUnitMonth      TimeUnit = "month"
	ExprUnitWeek       TimeUnit = "week"
	ExprUnitDay        TimeUnit = "day"
	ExprUnitHour       TimeUnit = "hour"
	ExprUnitMinute     TimeUnit = "minute"
	ExprUnitSecond     TimeUnit = "second"
	ExprUnitNanosecond TimeUnit = "nanosecond"
)

// timeUnitDuration maps time unit strings to their corresponding durations
var timeUnitDuration = map[TimeUnit]time.Duration{
	ExprUnitHour:       time.Hour,
	ExprUnitMinute:     time.Minute,
	ExprUnitSecond:     time.Second,
	ExprUnitNanosecond: time.Nanosecond,
}

// timeUnitAddDate maps time unit strings to their AddDate parameters
var timeUnitAddDate = map[TimeUnit]func(int) (int, int, int){
	ExprUnitYear:  func(delta int) (int, int, int) { return delta, 0, 0 },
	ExprUnitMonth: func(delta int) (int, int, int) { return 0, delta, 0 },
	ExprUnitWeek:  func(delta int) (int, int, int) { return 0, 0, delta * DaysInWeek },
	ExprUnitDay:   func(delta int) (int, int, int) { return 0, 0, delta },
}

// parseRelativeTimeExpression handles expressions like "5_days_ago" or "2_hours_from_now"
func parseRelativeTimeExpression(matches []string) (time.Time, bool) {
	delta, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, false
	}
	if matches[3] == "ago" {
		delta = -delta
	}

	unit := TimeUnit(matches[2])

	// Handle AddDate units (year, month, week, day)
	if addDateFunc, exists := timeUnitAddDate[unit]; exists {
		years, months, days := addDateFunc(delta)
		return time.Now().AddDate(years, months, days), true
	}

	// Handle Duration units (hour, minute, second, nanosecond)
	if duration, exists := timeUnitDuration[unit]; exists {
		return time.Now().Add(time.Duration(delta) * duration), true
	}

	return time.Time{}, false
}

// parseAbsoluteTimeExpression handles expressions like "now", "today", "yesterday", "tomorrow"
func parseAbsoluteTimeExpression(matches []string) (time.Time, bool) {
	base := time.Now().Truncate(HoursInDay * time.Hour)
	switch matches[1] {
	case "now":
		return time.Now(), true
	case "today":
		return base, true
	case "yesterday":
		return base.AddDate(0, 0, -1), true
	case "tomorrow":
		return base.AddDate(0, 0, 1), true
	}
	return time.Time{}, false
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

// handleEmptyValue processes empty values according to the options
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

// parseTimeExpression handles expression-based time parsing
func (t *Time) parseTimeExpression(value string, opts *TimeOptions) bool {
	if v := resolveTimeExpression(value); v != nil {
		t.Val = *v
		t.Present = true
		t.applyRounding(opts)
		return true
	}
	return false
}

// parseTimeFormat handles format-based time parsing
func (t *Time) parseTimeFormat(value string, format string, opts *TimeOptions) bool {
	if v, err := time.Parse(format, value); err == nil {
		t.Val = v
		t.Present = true
		t.applyRounding(opts)
		return true
	}
	return false
}

// parseTimeValue attempts to parse a time value using the given formats
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
