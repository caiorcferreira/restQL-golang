package cache

import (
	"context"
	"github.com/b2wdigital/restQL-golang/v4/internal/eval"
	"github.com/b2wdigital/restQL-golang/v4/pkg/restql"
	"github.com/pkg/errors"
	"regexp"
)

// MatchEvaluatorCache is a caching wrapper that implements the eval.MatchEvaluator interface.
type MatchEvaluatorCache struct {
	log             restql.Logger
	parseArgCache   *Cache
	matchValueCache *Cache
}

func NewMatchEvaluatorCache(log restql.Logger, parseArgCache *Cache, matchValueCache *Cache) *MatchEvaluatorCache {
	return &MatchEvaluatorCache{
		log:             log,
		parseArgCache:   parseArgCache,
		matchValueCache: matchValueCache,
	}
}

func (mc *MatchEvaluatorCache) ParseArg(arg interface{}) (*regexp.Regexp, error) {
	result, err := mc.parseArgCache.Get(context.Background(), arg)
	if err != nil {
		return nil, err
	}

	parsedArg, ok := result.(*regexp.Regexp)
	if !ok {
		mc.log.Info("failed to convert cache content", "content", result)
		return nil, errors.New("failed to convert cache content")
	}

	return parsedArg, nil
}

type matchValueCacheKey struct {
	regex *regexp.Regexp
	value interface{}
}

func (mc *MatchEvaluatorCache) MatchValue(matchRegex *regexp.Regexp, value interface{}) bool {
	key := matchValueCacheKey{regex: matchRegex, value: value}
	result, err := mc.matchValueCache.Get(context.Background(), key)
	if err != nil {
		return false
	}

	match, ok := result.(bool)
	if !ok {
		mc.log.Info("failed to convert cache content", "content", result)
		return false
	}

	return match
}

// ParseArgCacheLoader is the strategy to parse
// `matches` argument for the cached MatchEvaluator.
func ParseArgCacheLoader(me eval.MatchEvaluator) Loader {
	return func(ctx context.Context, key interface{}) (interface{}, error) {
		parsedArg, err := me.ParseArg(key)
		if err != nil {
			return nil, err
		}

		return parsedArg, nil
	}
}

// MatchValueCacheLoader is the strategy to check the target value
// against the `matches` argument for the cached MatchEvaluator.
func MatchValueCacheLoader(me eval.MatchEvaluator) Loader {
	return func(ctx context.Context, key interface{}) (interface{}, error) {
		cacheKey, ok := key.(matchValueCacheKey)
		if !ok {
			return nil, errors.Errorf("invalid key type : got %T", key)
		}

		match := me.MatchValue(cacheKey.regex, cacheKey.value)
		return match, nil
	}
}
