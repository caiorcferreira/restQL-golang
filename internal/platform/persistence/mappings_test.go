package persistence

import (
	"context"
	"fmt"
	"github.com/b2wdigital/restQL-golang/v6/pkg/restql"
	"io/ioutil"
	"testing"

	"github.com/b2wdigital/restQL-golang/v6/internal/platform/logger"
	"github.com/b2wdigital/restQL-golang/v6/test"
)

const mytenant = "mytenant"

func TestMappingsReader_Env(t *testing.T) {
	envSource := stubEnvSource{
		getAll: map[string]string{
			"RESTQL_MAPPING_mytenant_HERO":     "http://hero.api/",
			"RESTQL_MAPPING_mytenant_SIDEKICK": "http://sidekick.api/",
			"RESTQL_MAPPING_":                  "http://failed.api/",
			"RESTQL_MAPPING_mytenant_VILLAIN":  "http://villain.api/",
			"TEST_VAR":                         "foo",
		},
	}
	db := stubDatabase{}

	reader := NewMappingReader(noOpLogger, envSource, map[string]map[string]string{}, db)

	heroMapping, err := restql.NewMapping("hero", "http://hero.api/")
	test.VerifyError(t, err)

	sidekickMapping, err := restql.NewMapping("sidekick", "http://sidekick.api/")
	test.VerifyError(t, err)

	villainMapping, err := restql.NewMapping("villain", "http://villain.api/")
	test.VerifyError(t, err)

	expected := map[string]restql.Mapping{
		"hero":     heroMapping,
		"sidekick": sidekickMapping,
		"villain":  villainMapping,
	}

	mappings, err := reader.FromTenant(context.Background(), mytenant)

	test.VerifyError(t, err)
	test.Equal(t, mappings, expected)
}

func TestMappingsReader_Local(t *testing.T) {
	envSource := stubEnvSource{getAll: map[string]string{}}
	local := map[string]map[string]string{
		mytenant: {
			"villain": "http://villain.api/",
		},
	}
	db := stubDatabase{}

	reader := NewMappingReader(noOpLogger, envSource, local, db)

	villainMapping, err := restql.NewMapping("villain", "http://villain.api/")
	test.VerifyError(t, err)

	expected := map[string]restql.Mapping{
		"villain": villainMapping,
	}

	mappings, err := reader.FromTenant(context.Background(), mytenant)

	test.VerifyError(t, err)
	test.Equal(t, mappings, expected)
}

func TestMappingsReader_Database(t *testing.T) {
	envSource := stubEnvSource{getAll: map[string]string{}}
	local := map[string]map[string]string{}

	heroMapping, err := restql.NewMapping("hero", "http://hero.api/")
	test.VerifyError(t, err)

	sidekickMapping, err := restql.NewMapping("sidekick", "http://sidekick.api/")
	test.VerifyError(t, err)

	db := stubDatabase{findMappingsForTenant: []restql.Mapping{heroMapping, sidekickMapping}}

	reader := NewMappingReader(noOpLogger, envSource, local, db)

	expected := map[string]restql.Mapping{
		"hero":     heroMapping,
		"sidekick": sidekickMapping,
	}

	mappings, err := reader.FromTenant(context.Background(), mytenant)

	test.VerifyError(t, err)
	test.Equal(t, mappings, expected)
}

func TestMappingsReader_ShouldOverwriteMappings(t *testing.T) {
	heroMapping, err := restql.NewMapping("hero", "https://hero.com/api/")
	test.VerifyError(t, err)

	sidekickMapping, err := restql.NewMapping("sidekick", "https://sidekick.com/api")
	test.VerifyError(t, err)

	villainMapping, err := restql.NewMapping("villain", "http://villain.api/")
	test.VerifyError(t, err)

	local := map[string]map[string]string{
		mytenant: {
			"hero":     "http://hero.api/",
			"sidekick": "http://sidekick.api/",
			"villain":  "http://villain.api/",
		},
	}
	db := stubDatabase{
		findMappingsForTenant: []restql.Mapping{sidekickMapping},
	}
	envSource := stubEnvSource{
		getAll: map[string]string{
			fmt.Sprintf("RESTQL_MAPPING_%s_HERO", mytenant): "https://hero.com/api/",
		},
	}

	reader := NewMappingReader(noOpLogger, envSource, local, db)

	expected := map[string]restql.Mapping{
		"hero":     heroMapping,
		"sidekick": sidekickMapping,
		"villain":  villainMapping,
	}

	mappings, err := reader.FromTenant(context.Background(), mytenant)

	test.VerifyError(t, err)
	test.Equal(t, mappings, expected)
}

var noOpLogger = logger.New(ioutil.Discard, logger.LogOptions{})

type stubDatabase struct {
	findMappingsForTenant []restql.Mapping
	findQuery             restql.SavedQueryRevision
}

func (s stubDatabase) FindQueriesForNamespace(ctx context.Context, namespace string, archived bool) ([]restql.SavedQuery, error) {
	panic("implement me")
}

func (s stubDatabase) FindQueryWithAllRevisions(ctx context.Context, namespace string, queryName string, archived bool) (restql.SavedQuery, error) {
	panic("implement me")
}

func (s stubDatabase) UpdateQueryArchiving(ctx context.Context, namespace string, queryName string, archived bool) error {
	panic("implement me")
}

func (s stubDatabase) UpdateRevisionArchiving(ctx context.Context, namespace string, name string, revision int, archived bool) error {
	panic("implement me")
}

func (s stubDatabase) Name() string {
	panic("implement me")
}

func (s stubDatabase) FindAllNamespaces(ctx context.Context) ([]string, error) {
	panic("implement me")
}

func (s stubDatabase) CreateQueryRevision(ctx context.Context, namespace string, queryName string, content string) error {
	panic("implement me")
}

func (s stubDatabase) FindAllTenants(ctx context.Context) ([]string, error) {
	panic("implement me")
}

func (s stubDatabase) CreateMapping(ctx context.Context, tenantID string, mappingsName string, url string) error {
	panic("implement me")
}

func (s stubDatabase) SetMapping(ctx context.Context, tenantID string, mappingsName string, url string) error {
	panic("implement me")
}

func (s stubDatabase) FindMappingsForTenant(ctx context.Context, tenantID string) ([]restql.Mapping, error) {
	return s.findMappingsForTenant, nil
}

func (s stubDatabase) FindQuery(ctx context.Context, namespace string, name string, revision int) (restql.SavedQueryRevision, error) {
	return s.findQuery, nil
}

type stubEnvSource struct {
	getAll map[string]string
}

func (s stubEnvSource) GetString(key string) string {
	return ""
}

func (s stubEnvSource) GetAll() map[string]string {
	return s.getAll
}
