package configuration_test

import (
	"testing"

	"github.com/korchasa/speelka-agent-go/internal/configuration"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestValidatePromptTemplate(t *testing.T) {
	// Create a logger for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(logger)

	t.Run("valid template with all placeholders", func(t *testing.T) {
		// Test a template with both required placeholders
		template := `This is a template with {{query}} and {{tools}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with additional placeholders", func(t *testing.T) {
		// Test a template with required and additional placeholders
		template := `Template with {{query}}, {{tools}}, and {{extra}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("valid template with whitespace in placeholders", func(t *testing.T) {
		// Test a template with whitespace in the placeholder syntax
		template := `Template with {{ query }} and {{ tools }} placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.NoError(t, err)
	})

	t.Run("invalid template missing query placeholder", func(t *testing.T) {
		// Test a template missing the query placeholder
		template := `Template with only {{tools}} placeholder`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): query")
	})

	t.Run("invalid template missing tools placeholder", func(t *testing.T) {
		// Test a template missing the tools placeholder
		template := `Template with only {{query}} placeholder`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): tools")
	})

	t.Run("invalid empty template", func(t *testing.T) {
		// Test an empty template
		template := ``
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("invalid template with no placeholders", func(t *testing.T) {
		// Test a template with no placeholders
		template := `This is a template without any placeholders`
		err := cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
	})

	t.Run("different argument name", func(t *testing.T) {
		// Test with a different argument name than "query"
		template := `Template with {{input}} and {{tools}} placeholders`
		err := cm.TestValidatePromptTemplate(template, "input")
		assert.NoError(t, err)

		// Should fail when looking for a different argument name
		err = cm.TestValidatePromptTemplate(template, "query")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required placeholder(s): query")
	})
}

func TestExtractPlaceholders(t *testing.T) {
	// Create a logger for testing
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	// Create a configuration manager
	cm := configuration.NewConfigurationManager(logger)

	t.Run("extract multiple placeholders", func(t *testing.T) {
		template := `This is a {{test}} template with {{multiple}} placeholders including {{tools}}`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"test", "multiple", "tools"}, placeholders)
	})

	t.Run("extract placeholders with whitespace", func(t *testing.T) {
		template := `This has {{ spaced }} placeholders and {{unspaced}} ones`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"spaced", "unspaced"}, placeholders)
	})

	t.Run("handle no placeholders", func(t *testing.T) {
		template := `This template has no placeholders`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})

	t.Run("handle empty template", func(t *testing.T) {
		template := ``
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.Empty(t, placeholders)
	})

	t.Run("handle complex nested content", func(t *testing.T) {
		template := `Complex template with {{placeholder}} and code snippets like if (x == y) { return true; }`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"placeholder"}, placeholders)
	})

	t.Run("extract placeholder with numbers and underscores", func(t *testing.T) {
		template := `Template with {{place_holder_123}} containing numbers and underscores`
		placeholders, err := cm.TestExtractPlaceholders(template)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"place_holder_123"}, placeholders)
	})
}
