package eval

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

// MatchEvaluator represents a toolset for
// the running the `matches` function.
type MatchEvaluator interface {
	// ParseArg creates a `*regexp.Regexp` from the argument
	// of the `matches` function.
	ParseArg(interface{}) (*regexp.Regexp, error)

	// MatchValue checks if the value provided matches
	// the specified regular expression.
	MatchValue(*regexp.Regexp, interface{}) bool
}

// NewMatchEvaluator constructs an implementation
// of the MatchEvaluator interface
func NewMatchEvaluator() MatchEvaluator {
	return matchEvaluator{}
}

type matchEvaluator struct{}

func (m matchEvaluator) ParseArg(arg interface{}) (*regexp.Regexp, error) {
	switch arg := arg.(type) {
	case *regexp.Regexp:
		return arg, nil
	case string:
		return regexp.Compile(arg)
	default:
		return nil, errors.New("failed to parse match argument : unknown match argument type")
	}
}

func (m matchEvaluator) MatchValue(matchRegex *regexp.Regexp, value interface{}) bool {
	var strVal string
	if value, ok := value.(string); ok {
		strVal = value
	} else {
		strVal = fmt.Sprintf("%v", value)
	}

	return matchRegex.MatchString(strVal)
}
