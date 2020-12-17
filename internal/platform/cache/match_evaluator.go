package cache

import (
	"context"
	"github.com/b2wdigital/restQL-golang/v4/internal/eval"
	"github.com/b2wdigital/restQL-golang/v4/internal/platform/conf"
	"github.com/b2wdigital/restQL-golang/v4/pkg/restql"
	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"regexp"
)

type MatchEvaluatorCacheOptions func(*MatchEvaluatorCache)

func WithParserArgumentCache(c *Cache) func(*MatchEvaluatorCache) {
	return func(mec *MatchEvaluatorCache) {
		mec.parseArgCache = c
	}
}

func WithMatchValueCache(c *Cache) func(*MatchEvaluatorCache) {
	return func(mec *MatchEvaluatorCache) {
		mec.matchValueCache = c
	}
}

// MatchEvaluatorCache is a caching wrapper that implements the eval.MatchEvaluator interface.
type MatchEvaluatorCache struct {
	log             restql.Logger
	matchEvaluator  eval.MatchEvaluator
	parseArgCache   *Cache
	matchValueCache *Cache
	rawCache        gcache.Cache
}

func NewMatchEvaluatorCache(log restql.Logger, me eval.MatchEvaluator, cfg *conf.Config, options ...MatchEvaluatorCacheOptions) *MatchEvaluatorCache {
	mec := &MatchEvaluatorCache{log: log, matchEvaluator: me}
	for _, option := range options {
		option(mec)
	}

	mec.rawCache = gcache.New(cfg.Cache.Matches.ResultMaxSize).LRU().Build()

	return mec
}

func (mc *MatchEvaluatorCache) ParseArg(arg interface{}) (*regexp.Regexp, error) {
	if mc.parseArgCache == nil {
		return mc.matchEvaluator.ParseArg(arg)
	}

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
	regex string
	value interface{}
}

func (mc *MatchEvaluatorCache) MatchValue(matchRegex *regexp.Regexp, value interface{}) bool {
	key := matchValueCacheKey{regex: matchRegex.String(), value: value}
	result, err := mc.rawCache.Get(key)
	if err != nil {
		match := mc.MatchValue(matchRegex, value)
		err := mc.rawCache.Set(key, match)
		if err != nil {
			mc.log.Error("failed to set match value on cache", err)
		}

		return match
	}

	match, ok := result.(bool)
	if !ok {
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
//func MatchValueCacheLoader(me eval.MatchEvaluator) Loader {
//	return func(ctx context.Context, key interface{}) (interface{}, error) {
//		cacheKey, ok := key.(matchValueCacheKey)
//		if !ok {
//			return nil, errors.Errorf("invalid key type : got %T", key)
//		}
//
//		match := me.MatchValue(cacheKey.regex, cacheKey.value)
//		return match, nil
//	}
//}
