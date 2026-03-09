package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInitialState(t *testing.T) {
	t.Parallel()

	m := NewModel("")
	assert.Equal(t, stateSearchInput, m.state)
}

func TestModelCreation(t *testing.T) {
	t.Parallel()

	testQuery := "mysql"
	m := NewModel(testQuery)
	assert.Equal(t, stateSearchInput, m.state)
	assert.Equal(t, testQuery, m.search.searchInput.Value())
}

func TestDefaultSearchState(t *testing.T) {
	t.Parallel()

	m := NewModel("")
	assert.NotNil(t, m.search.searchInput)
	assert.NotNil(t, m.search.list)
	assert.Equal(t, 0, len(m.search.list.Items()))
}

func TestModelInit(t *testing.T) {
	t.Parallel()

	// With initial search query
	m1 := NewModel("nginx")
	cmd1 := m1.Init()
	assert.NotNil(t, cmd1)

	// Without initial search query
	m2 := NewModel("")
	cmd2 := m2.Init()
	assert.NotNil(t, cmd2)
}

func TestStateManagementTypes(t *testing.T) {
	t.Parallel()

	assert.IsType(t, stateSearchInput, stateSearchResults)
	assert.IsType(t, stateSearchResults, stateTags)
	assert.IsType(t, stateTags, stateArchSelect)
	assert.IsType(t, stateArchSelect, stateDownload)
	assert.IsType(t, stateDownload, stateDone)
}

func TestModelUpdateReturnsCmd(t *testing.T) {
	t.Parallel()

	m := NewModel("")
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	_, cmd := m.Update(keyMsg)
	assert.NotNil(t, cmd)
}

func TestSearchToTagsStateReset(t *testing.T) {
	t.Parallel()

	testRepoName := "nginx"
	m := NewModel("")

	// Set to search results state with a selected repo
	m.state = stateSearchResults
	m.search.selected = testRepoName
	m.search.selectedDesc = "Test repo"

	// Verify state
	assert.Equal(t, stateSearchResults, m.state)
	assert.Equal(t, testRepoName, m.search.selected)

	// Transition to tags state
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

	teaModel, _ := m.Update(keyMsg)
	m = teaModel.(Model)

	// Should transition to tags
	assert.Equal(t, stateTags, m.state)
	assert.Empty(t, m.search.selected)
	assert.Empty(t, m.search.selectedDesc)
}

func TestTagsStateReset(t *testing.T) {
	t.Parallel()

	m := NewModel("")
	testTag := "latest"
	testRepo := "nginx"

	m.state = stateTags
	m.tags.repository = testRepo
	m.tags.selected = testTag
	m.tags.back = true

	t.Logf("Before transition: m.state=%d, tags.repo=%q, tags.selected=%q", m.state, m.tags.repository, m.tags.selected)

	// Then update the model to transition to search results state
	teaModel, _ := m.Update(nil)
	m = teaModel.(Model)

	t.Logf("After transition: m.state=%d, tags.repo=%q, tags.selected=%q, tags.back=%v", m.state, m.tags.repository, m.tags.selected, m.tags.back)

	assert.Equal(t, stateSearchResults, m.state)
	assert.Empty(t, m.tags.repository)
	assert.Empty(t, m.tags.selected)
	assert.False(t, m.tags.back)
}
