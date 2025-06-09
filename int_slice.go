package meta

import (
	"reflect"
	"strings"
)

//
// Int64Slice, TODO: Uint64Slice
//

type Int64Slice struct {
	Presence
	Nullity
	Val  []int64
	Path string
}

type IntSliceOptions struct {
	*IntOptions
	*SliceOptions
}

func (i *Int64Slice) ParseOptions(tag reflect.StructTag) interface{} {
	var tempI Int64
	opts := tempI.ParseOptions(tag)

	return &IntSliceOptions{
		IntOptions:   opts.(*IntOptions),
		SliceOptions: ParseSliceOptions(tag),
	}
}

func (n *Int64Slice) JSONValue(path string, i interface{}, options interface{}) Errorable {
	opts := options.(*IntSliceOptions)
	intOpts := opts.IntOptions
	sliceOpts := opts.SliceOptions

	n.Path = path
	n.Val = nil
	n.Present = true
	n.Null = false

	if i == nil {
		n.Null = true
		if sliceOpts.DiscardBlank {
			n.Present = false
			return nil
		} else if sliceOpts.Null {
			return nil
		}
		return ErrBlank
	}

	var errorsInSlice ErrorSlice
	switch value := i.(type) {
	case string:
		return n.FormValue(value, options)
	case []interface{}:
		n.Val = []int64{}

		if sliceOpts.MinLengthPresent && len(value) < sliceOpts.MinLength {
			return ErrMinLength
		}

		if sliceOpts.MaxLengthPresent && len(value) > sliceOpts.MaxLength {
			return ErrMaxLength
		}

		if len(value) == 0 {
			if sliceOpts.DiscardBlank {
				n.Present = false
				return nil
			} else if sliceOpts.Blank {
				return nil
			}
			return ErrBlank
		}

		for _, v := range value {
			var num Int64
			if err := num.JSONValue("", v, intOpts); err != nil {
				errorsInSlice = append(errorsInSlice, err)
			} else {
				errorsInSlice = append(errorsInSlice, nil)
				n.Val = append(n.Val, num.Val)
			}
		}
		if errorsInSlice.Len() > 0 {
			return errorsInSlice
		}

		if sliceOpts.MinLengthPresent && len(n.Val) < sliceOpts.MinLength {
			return ErrMinLength
		}

		if sliceOpts.MaxLengthPresent && len(n.Val) > sliceOpts.MaxLength {
			return ErrMaxLength
		}

		if len(n.Val) == 0 {
			if sliceOpts.DiscardBlank {
				n.Present = false
				return nil
			} else if sliceOpts.Blank {
				return nil
			}
			return ErrBlank
		}
	}
	return nil
}

func (i *Int64Slice) FormValue(value string, options interface{}) Errorable {
	var tempI Int64

	opts := options.(*IntSliceOptions)
	intOpts := opts.IntOptions
	sliceOpts := opts.SliceOptions

	i.Val = []int64{}
	i.Present = true
	i.Null = false

	value = strings.TrimSpace(value)

	if value == "" {
		if sliceOpts.DiscardBlank {
			i.Present = false
			return nil
		} else if sliceOpts.Blank {
			return nil
		}
		return ErrBlank
	}

	strs := strings.Split(value, ",")

	if sliceOpts.MinLengthPresent && len(strs) < sliceOpts.MinLength {
		return ErrMinLength
	}

	if sliceOpts.MaxLengthPresent && len(strs) > sliceOpts.MaxLength {
		return ErrMaxLength
	}

	var errorsInSlice ErrorSlice
	for _, s := range strs {
		tempI.Val = 0
		if err := tempI.FormValue(s, intOpts); err != nil {
			errorsInSlice = append(errorsInSlice, err)
		} else {
			errorsInSlice = append(errorsInSlice, nil)
			i.Val = append(i.Val, tempI.Val)
		}
	}

	if errorsInSlice.Len() > 0 {
		return errorsInSlice
	}

	if sliceOpts.MinLengthPresent && len(i.Val) < sliceOpts.MinLength {
		return ErrMinLength
	}

	if sliceOpts.MaxLengthPresent && len(i.Val) > sliceOpts.MaxLength {
		return ErrMaxLength
	}

	if len(i.Val) == 0 {
		if sliceOpts.DiscardBlank {
			i.Present = false
			return nil
		} else if sliceOpts.Blank {
			return nil
		}
		return ErrBlank
	}

	return nil
}

func (s Int64Slice) MarshalJSON() ([]byte, error) {
	if len(s.Val) > 0 {
		return MetaJson.Marshal(s.Val)
	}
	return nullString, nil
}

func (s *Int64Slice) UnmarshalJSON(data []byte) error {
	var value []int64
	err := MetaJson.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	*s = Int64Slice{Val: value}
	return nil
}
