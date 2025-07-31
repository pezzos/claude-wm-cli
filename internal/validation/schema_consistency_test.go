package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSchemaConsistency vérifie que les structures Go sont cohérentes avec les schémas JSON
func TestSchemaConsistency(t *testing.T) {
	tests := []struct {
		name          string
		jsonFile      string
		schemaFile    string
		goStructCheck func(t *testing.T)
	}{
		{
			name:          "stories.json structure consistency",
			jsonFile:      "stories.json",
			schemaFile:    "stories.schema.json",
			goStructCheck: verifyStoriesStructure,
		},
		{
			name:          "epics.json structure consistency", 
			jsonFile:      "epics.json",
			schemaFile:    "epics.schema.json",
			goStructCheck: verifyEpicsStructure,
		},
		{
			name:          "current-story.json structure consistency",
			jsonFile:      "current-story.json", 
			schemaFile:    "current-story.schema.json",
			goStructCheck: verifyCurrentStoryStructure,
		},
		{
			name:          "current-epic.json structure consistency",
			jsonFile:      "current-epic.json",
			schemaFile:    "current-epic.schema.json", 
			goStructCheck: verifyCurrentEpicStructure,
		},
		{
			name:          "current-task.json structure consistency",
			jsonFile:      "current-task.json",
			schemaFile:    "current-task.schema.json",
			goStructCheck: verifyCurrentTaskStructure,
		},
		{
			name:          "iterations.json structure consistency",
			jsonFile:      "iterations.json",
			schemaFile:    "iterations.schema.json",
			goStructCheck: verifyIterationsStructure,
		},
		{
			name:          "metrics.json structure consistency",
			jsonFile:      "metrics.json",
			schemaFile:    "metrics.schema.json",
			goStructCheck: verifyMetricsStructure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Vérifier que le schéma existe
			schemaPath := filepath.Join("../../internal/config/system/commands/templates/schemas", tt.schemaFile)
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				t.Errorf("Schema file not found: %s", schemaPath)
				return
			}

			// Charger et valider le schéma
			schemaData, err := os.ReadFile(schemaPath)
			if err != nil {
				t.Errorf("Failed to read schema file: %v", err)
				return
			}

			var schema map[string]interface{}
			if err := json.Unmarshal(schemaData, &schema); err != nil {
				t.Errorf("Invalid JSON schema: %v", err)
				return
			}

			// Vérifier les propriétés requises du schéma
			if _, ok := schema["properties"]; !ok {
				t.Errorf("Schema missing 'properties' field")
				return
			}

			// Exécuter la vérification spécifique de la structure Go
			if tt.goStructCheck != nil {
				tt.goStructCheck(t)
			}
		})
	}
}

// verifyStoriesStructure vérifie que les structures Go pour stories.json sont cohérentes
func verifyStoriesStructure(t *testing.T) {
	// Vérifier que cmd/story.go utilise la bonne structure
	storyGoPath := "../../cmd/story.go"
	content, err := os.ReadFile(storyGoPath)
	if err != nil {
		t.Errorf("Failed to read story.go: %v", err)
		return
	}

	contentStr := string(content)
	
	// Vérifier que Stories est défini comme map[string]struct et non []struct
	if strings.Contains(contentStr, "Stories []struct") {
		t.Errorf("cmd/story.go: Stories should be map[string]struct, not []struct")
	}
	
	if !strings.Contains(contentStr, "Stories map[string]struct") {
		t.Errorf("cmd/story.go: Stories should be defined as map[string]struct")
	}

	// Vérifier que cmd/ticket.go utilise la bonne structure
	ticketGoPath := "../../cmd/ticket.go"
	ticketContent, err := os.ReadFile(ticketGoPath)
	if err != nil {
		t.Errorf("Failed to read ticket.go: %v", err)
		return
	}

	ticketContentStr := string(ticketContent)
	
	// Vérifier que Stories est défini comme map[string]struct et non []struct
	if strings.Contains(ticketContentStr, "Stories []struct") {
		t.Errorf("cmd/ticket.go: Stories should be map[string]struct, not []struct")
	}
	
	if !strings.Contains(ticketContentStr, "Stories map[string]struct") {
		t.Errorf("cmd/ticket.go: Stories should be defined as map[string]struct")
	}
}

// verifyEpicsStructure vérifie que les structures Go pour epics.json sont cohérentes
func verifyEpicsStructure(t *testing.T) {
	epicGoPath := "../../cmd/epic.go"
	content, err := os.ReadFile(epicGoPath)
	if err != nil {
		t.Errorf("Failed to read epic.go: %v", err)
		return
	}

	contentStr := string(content)
	
	// Vérifier que le fichier contient les structures appropriées pour epics.json
	if !strings.Contains(contentStr, "epics.json") {
		t.Errorf("cmd/epic.go should reference epics.json")
	}
}

// verifyCurrentStoryStructure vérifie que les structures Go pour current-story.json sont cohérentes
func verifyCurrentStoryStructure(t *testing.T) {
	// Cette structure est utilisée dans plusieurs fichiers, on vérifie les principaux
	files := []string{
		"../../cmd/ticket.go",
		"../../internal/navigation/context.go",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "current-story.json") {
			// Le fichier utilise current-story.json, vérifier la structure
			if !strings.Contains(contentStr, `"story"`) {
				t.Errorf("%s: current-story.json structure should have 'story' field", file)
			}
		}
	}
}

// verifyCurrentEpicStructure vérifie que les structures Go pour current-epic.json sont cohérentes
func verifyCurrentEpicStructure(t *testing.T) {
	// Cette structure est utilisée dans internal/navigation/context.go principalement
	contextGoPath := "../../internal/navigation/context.go"
	content, err := os.ReadFile(contextGoPath)
	if err != nil {
		t.Errorf("Failed to read context.go: %v", err)
		return
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "current-epic.json") {
		// Le fichier utilise current-epic.json, vérifier la structure
		if !strings.Contains(contentStr, `"epic"`) {
			t.Errorf("internal/navigation/context.go: current-epic.json structure should have 'epic' field")
		}
	}
}

// verifyCurrentTaskStructure vérifie que les structures Go pour current-task.json sont cohérentes
func verifyCurrentTaskStructure(t *testing.T) {
	// Cette structure est utilisée dans internal/navigation/context.go et workflow/analyzer.go
	files := []string{
		"../../internal/navigation/context.go",
		"../../internal/workflow/analyzer.go",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("Failed to read %s: %v", file, err)
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "current-task.json") {
			// Le fichier utilise current-task.json, vérifier les champs requis
			requiredFields := []string{"id", "title", "description", "type", "priority", "status"}
			for _, field := range requiredFields {
				if !strings.Contains(contentStr, fmt.Sprintf(`"%s"`, field)) {
					t.Errorf("%s: current-task.json structure should have '%s' field", file, field)
				}
			}
		}
	}
}

// verifyIterationsStructure vérifie que les structures Go pour iterations.json sont cohérentes
func verifyIterationsStructure(t *testing.T) {
	// La structure principale est définie dans internal/preprocessing/task_preprocessing.go
	preprocessingGoPath := "../../internal/preprocessing/task_preprocessing.go"
	content, err := os.ReadFile(preprocessingGoPath)
	if err != nil {
		t.Errorf("Failed to read task_preprocessing.go: %v", err)
		return
	}

	contentStr := string(content)
	
	// Vérifier que IterationsData a les champs requis
	if !strings.Contains(contentStr, "type IterationsData struct") {
		t.Errorf("internal/preprocessing/task_preprocessing.go: IterationsData struct not found")
		return
	}

	// Vérifier les champs requis avec leurs tags JSON
	requiredFields := []string{
		`TaskContext.*json:"task_context"`,
		`Iterations.*json:"iterations"`,
		`FinalOutcome.*json:"final_outcome"`,
		`Recommendations.*json:"recommendations"`,
	}
	
	for _, fieldPattern := range requiredFields {
		matched, err := filepath.Match("*"+fieldPattern+"*", contentStr)
		if err != nil || !matched {
			// Utiliser une regex plus simple
			if !strings.Contains(contentStr, strings.Split(fieldPattern, `.*json:"`)[1][:len(strings.Split(fieldPattern, `.*json:"`)[1])-1]) {
				t.Errorf("internal/preprocessing/task_preprocessing.go: missing field with json tag %s", fieldPattern)
			}
		}
	}

	// Vérifier que cmd/interactive.go utilise bien cette structure
	interactiveGoPath := "../../cmd/interactive.go"
	interactiveContent, err := os.ReadFile(interactiveGoPath)
	if err != nil {
		t.Errorf("Failed to read interactive.go: %v", err)
		return
	}

	interactiveContentStr := string(interactiveContent)
	if strings.Contains(interactiveContentStr, "iterations.json") {
		// Vérifier qu'il utilise preprocessing.IterationsData
		if !strings.Contains(interactiveContentStr, "preprocessing.IterationsData") {
			t.Errorf("cmd/interactive.go: should use preprocessing.IterationsData for iterations.json")
		}
	}
}

// verifyMetricsStructure vérifie que les structures Go pour metrics.json sont cohérentes
func verifyMetricsStructure(t *testing.T) {
	// Pour l'instant, metrics.json n'est pas directement parsé par des fichiers Go
	// mais on peut vérifier que le schéma est valide
	schemaPath := "../../internal/config/system/commands/templates/schemas/metrics.schema.json"
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Errorf("Failed to read metrics.schema.json: %v", err)
		return
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Errorf("Invalid metrics.schema.json: %v", err)
	}
}