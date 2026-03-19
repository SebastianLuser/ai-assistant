package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyMessage_FinanceKeywords(t *testing.T) {
	tests := []struct {
		name    string
		message string
		tags    []string
	}{
		{"gasto", "anota este gasto", []string{"finance", "sheets"}},
		{"pago", "hice un pago de 500", []string{"finance", "sheets"}},
		{"plata", "cuanta plata gaste", []string{"finance", "sheets"}},
		{"dolar", "compre dolares", []string{"finance", "sheets"}},
		{"usd", "pagué 20 usd", []string{"finance", "sheets"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyMessage(tt.message)

			for _, tag := range tt.tags {
				assert.Contains(t, result, tag)
			}
		})
	}
}

func TestClassifyMessage_SpotifyKeywords(t *testing.T) {
	result := ClassifyMessage("que estoy escuchando en spotify")

	assert.Contains(t, result, "spotify")
	assert.Contains(t, result, "music")
	assert.Contains(t, result, "player")
}

func TestClassifyMessage_CalendarKeywords(t *testing.T) {
	result := ClassifyMessage("agendame una reunion para mañana")

	assert.Contains(t, result, "calendar")
	assert.Contains(t, result, "google")
}

func TestClassifyMessage_GithubKeywords(t *testing.T) {
	result := ClassifyMessage("mostrame los issues del repo")

	assert.Contains(t, result, "github")
	assert.Contains(t, result, "code")
}

func TestClassifyMessage_TodoistKeywords(t *testing.T) {
	result := ClassifyMessage("creame una tarea en todoist")

	assert.Contains(t, result, "todoist")
	assert.Contains(t, result, "tasks")
}

func TestClassifyMessage_GmailKeywords(t *testing.T) {
	result := ClassifyMessage("tengo mails sin leer?")

	assert.Contains(t, result, "gmail")
	assert.Contains(t, result, "email")
}

func TestClassifyMessage_NotionKeywords(t *testing.T) {
	result := ClassifyMessage("guardalo en notion")

	assert.Contains(t, result, "notion")
	assert.Contains(t, result, "notes")
}

func TestClassifyMessage_ObsidianKeywords(t *testing.T) {
	result := ClassifyMessage("busca en el vault de obsidian")

	assert.Contains(t, result, "obsidian")
	assert.Contains(t, result, "notes")
}

func TestClassifyMessage_MemoryKeywords(t *testing.T) {
	result := ClassifyMessage("guarda esta nota")

	assert.Contains(t, result, "memory")
	assert.Contains(t, result, "notes")
}

func TestClassifyMessage_HabitsKeywords(t *testing.T) {
	result := ClassifyMessage("logea mi habito de ejercicio")

	assert.Contains(t, result, "habits")
	assert.Contains(t, result, "tracking")
}

func TestClassifyMessage_NoMatch_ReturnsNil(t *testing.T) {
	result := ClassifyMessage("hola como estas")

	assert.Nil(t, result)
}

func TestClassifyMessage_CaseInsensitive(t *testing.T) {
	result := ClassifyMessage("SPOTIFY me esta sonando")

	assert.Contains(t, result, "spotify")
}

func TestClassifyMessage_MultipleMatches(t *testing.T) {
	result := ClassifyMessage("anota el gasto y creame una tarea")

	assert.Contains(t, result, "finance")
	assert.Contains(t, result, "tasks")
}

func TestClassifyMessage_EmptyMessage(t *testing.T) {
	result := ClassifyMessage("")

	assert.Nil(t, result)
}

func TestClassifyMessage_JiraKeywords(t *testing.T) {
	result := ClassifyMessage("mostrame el ticket de jira")

	assert.Contains(t, result, "jira")
	assert.Contains(t, result, "tasks")
	assert.Contains(t, result, "project")
}

func TestClassifyMessage_LinksKeywords(t *testing.T) {
	result := ClassifyMessage("guardame este link")

	assert.Contains(t, result, "links")
	assert.Contains(t, result, "bookmarks")
}

func TestClassifyMessage_UniversityKeywords(t *testing.T) {
	result := ClassifyMessage("tengo parcial en la facultad")

	assert.Contains(t, result, "university")
	assert.Contains(t, result, "uade")
}

func TestClassifyMessage_NoDuplicateTags(t *testing.T) {
	// "gasto" and "pago" both map to finance+sheets, result should have no duplicates
	result := ClassifyMessage("anota el gasto del pago")

	tagCount := make(map[string]int)
	for _, tag := range result {
		tagCount[tag]++
	}

	for tag, count := range tagCount {
		assert.Equal(t, 1, count, "tag %s appears %d times", tag, count)
	}
}
