package eval

import (
	"context"
	"github.com/b2wdigital/restQL-golang/internal/domain"
	"github.com/b2wdigital/restQL-golang/internal/parser"
	"github.com/b2wdigital/restQL-golang/internal/platform/plugins"
	"github.com/b2wdigital/restQL-golang/internal/runner"
	"github.com/pkg/errors"
)

var (
	ErrInvalidRevision  = errors.New("revision must be greater than 0")
	ErrInvalidQueryId   = errors.New("query id must be not empty")
	ErrInvalidNamespace = errors.New("namespace must be not empty")
)

type Evaluator struct {
	log            domain.Logger
	parser         parser.Parser
	mappingsReader MappingsReader
	queryReader    QueryReader
	runner         runner.Runner
	pluginsManager plugins.Manager
}

func NewEvaluator(log domain.Logger, mr MappingsReader, qr QueryReader, r runner.Runner, p parser.Parser, pm plugins.Manager) Evaluator {
	return Evaluator{
		log:            log,
		mappingsReader: mr,
		queryReader:    qr,
		runner:         r,
		parser:         p,
		pluginsManager: pm,
	}
}

func (e Evaluator) SavedQuery(ctx context.Context, queryOpts domain.QueryOptions, queryInput domain.QueryInput) (domain.Resources, error) {
	err := validateQueryOptions(queryOpts)
	if err != nil {
		return nil, err
	}

	queryTxt, err := e.queryReader.Get(ctx, queryOpts.Namespace, queryOpts.Id, queryOpts.Revision)
	if err != nil {
		return nil, err
	}

	query, err := e.parser.Parse(queryTxt)
	if err != nil {
		e.log.Debug("failed to parse query", "error", err)
		return nil, ParserError{errors.Wrap(err, "invalid query syntax")}
	}

	mappings, err := e.mappingsReader.FromTenant(ctx, queryOpts.Tenant)
	if err != nil {
		e.log.Error("failed to fetch mappings", err)
		return nil, err
	}

	queryCtx := domain.QueryContext{
		Mappings: mappings,
		Options:  queryOpts,
		Input:    queryInput,
	}

	e.pluginsManager.RunBeforeQuery(queryTxt, queryCtx)

	resources, err := e.runner.ExecuteQuery(ctx, query, queryCtx)
	switch {
	case err == runner.ErrQueryTimedOut:
		return nil, TimeoutError{Err: err}
	case err != nil:
		return nil, err
	}

	resources, err = ApplyFilters(query, resources)
	if err != nil {
		e.log.Debug("failed to apply filters", "error", err)
		return nil, err
	}

	resources = ApplyAggregators(query, resources)

	e.pluginsManager.RunAfterQuery(queryTxt, resources)

	return resources, nil
}

func validateQueryOptions(queryOpts domain.QueryOptions) error {
	if queryOpts.Revision <= 0 {
		return ValidationError{ErrInvalidRevision}
	}

	if queryOpts.Id == "" {
		return ValidationError{ErrInvalidQueryId}
	}

	if queryOpts.Namespace == "" {
		return ValidationError{ErrInvalidNamespace}
	}

	return nil
}
