package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

// Record represents information from any of our data sources
type Record struct {
	Location string
	DataType string
	Values   map[string]string
	Source   string
}

// Interaction stores a user interaction
type Interaction struct {
	Location  string
	Question  string
	Timestamp time.Time
}

var (
	dataFiles = map[string]string{
		"tax":     "data/state_taxes.csv",
		"tourist": "data/tourist_info.csv",
		"cost":    "data/travel_costs.csv",
	}
	lastQuery    string
	lastLocation string
	interactions []Interaction // Store all interactions
	memorySize   = 5           // Number of interactions to remember
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Start an interactive query session",
	Long: `Start an interactive session where you can ask questions about locations.
Example: "Tell me about California"`,
	Run: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
}

// LoadData reads and parses CSV files into Records
func LoadData(filePath string, dataType string) ([]Record, error) {
	if err := validateFilePath(filePath); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or missing data rows")
	}

	var result []Record
	headers := records[0]

	for i := 1; i < len(records); i++ {
		record := records[i]
		if err := validateRecord(record); err != nil {
			return nil, fmt.Errorf("invalid data at row %d: %v", i+1, err)
		}

		values := make(map[string]string)
		for j, header := range headers {
			if header != "location" && header != "source" && j < len(record) {
				values[header] = record[j]
			}
		}

		result = append(result, Record{
			Location: record[0],
			DataType: dataType,
			Values:   values,
			Source:   record[len(record)-1],
		})
	}

	return result, nil
}

// validateFilePath performs security checks on the file path
func validateFilePath(path string) error {
	// Check if path is absolute
	if filepath.IsAbs(path) {
		return fmt.Errorf("must use relative path to data directory")
	}

	// Check file extension
	if ext := filepath.Ext(path); ext != ".csv" {
		return fmt.Errorf("invalid file extension: must be .csv")
	}

	// Check for path traversal attempts
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: directory traversal not allowed")
	}

	// Check for command injection attempts
	if strings.Contains(path, "$") || strings.Contains(path, "`") {
		return fmt.Errorf("invalid character in field")
	}

	// Check for hidden files
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return fmt.Errorf("invalid filename: hidden files not allowed")
	}

	// Check for URLs
	if strings.Contains(path, "://") {
		return fmt.Errorf("invalid path: URLs not allowed")
	}

	return nil
}

// validateRecord performs security checks on CSV record content
func validateRecord(record []string) error {
	for _, field := range record {
		// Check for null bytes
		if strings.Contains(field, "\x00") {
			return fmt.Errorf("invalid character in field")
		}

		// Check for command injection attempts
		if strings.Contains(field, "$") || strings.Contains(field, "`") {
			return fmt.Errorf("invalid character in field")
		}

		// Check for path traversal in fields
		if strings.Contains(field, "../") || strings.Contains(field, "..\\") {
			return fmt.Errorf("invalid character in path")
		}

		// Check field length (limit to 1KB)
		if len(field) > 1024 {
			return fmt.Errorf("field too long: maximum length is 1KB")
		}
	}

	return nil
}

// FindRelevantInfo searches for information based on the query
func FindRelevantInfo(query string, records []Record) (string, string) {
	query = strings.ToLower(strings.TrimSpace(query))
	var mainResponse []string
	var followUp string
	seenTypes := make(map[string]bool)

	// Check for memory-related queries
	if strings.Contains(query, "previous") || strings.Contains(query, "history") ||
		strings.Contains(query, "locations") || strings.Contains(query, "remember") ||
		strings.Contains(query, "which") && strings.Contains(query, "ask") {
		return getMemoryInfo(), ""
	}

	// Handle follow-up questions about costs, attractions, etc.
	if strings.Contains(query, "cost") || strings.Contains(query, "price") ||
		strings.Contains(query, "attractions") || strings.Contains(query, "visit") ||
		strings.Contains(query, "tax") {
		if lastLocation != "" {
			query = lastLocation + " " + query
		}
	}

	// Handle follow-up questions
	if strings.HasPrefix(query, "what about") || strings.HasPrefix(query, "how about") {
		query = strings.TrimPrefix(query, "what about")
		query = strings.TrimPrefix(query, "how about")
		query = strings.TrimSpace(query)
	}

	// If query is very short and we have a last location, assume it's about the last location
	if len(strings.Fields(query)) <= 2 && lastLocation != "" &&
		!strings.Contains(strings.ToLower(query), strings.ToLower(lastLocation)) {
		query = lastLocation + " " + query
	}

	// Split query into words for better matching
	queryWords := strings.Fields(query)
	foundLocation := ""

	// First pass: find the location
	for _, record := range records {
		location := strings.ToLower(record.Location)
		for _, word := range queryWords {
			if strings.Contains(location, word) {
				foundLocation = record.Location
				break
			}
		}
		if foundLocation != "" {
			break
		}
	}

	// Second pass: gather all information for the found location
	if foundLocation != "" {
		for _, record := range records {
			if record.Location == foundLocation && !seenTypes[record.DataType] {
				seenTypes[record.DataType] = true
				info := formatRecordInfo(record)
				mainResponse = append(mainResponse, info)
			}
		}

		// If it's a new location
		if foundLocation != lastLocation {
			lastLocation = foundLocation
			followUp = fmt.Sprintf("\nWould you like to know more about %s? You can ask about:\n"+
				"- Tourist attractions and best time to visit\n"+
				"- Average daily costs and expenses\n"+
				"- Tax rates and financial information", foundLocation)
		}

		// Add to interactions history
		addInteraction(foundLocation, query)
	}

	if len(mainResponse) == 0 {
		if lastLocation != "" {
			return fmt.Sprintf("I assume you're asking about %s, but I don't have that specific information. Try asking about:\n"+
				"- Tourist attractions and best time to visit\n"+
				"- Average daily costs and expenses\n"+
				"- Tax rates and financial information", lastLocation), ""
		}
		return fmt.Sprintf("No information found. Available locations: %s",
			getAvailableLocations(records)), ""
	}

	return strings.Join(mainResponse, "\n"), followUp
}

// addInteraction adds a new interaction to the memory
func addInteraction(location, question string) {
	interaction := Interaction{
		Location:  location,
		Question:  question,
		Timestamp: time.Now(),
	}

	// Add to the beginning of the slice
	interactions = append([]Interaction{interaction}, interactions...)

	// Keep only the last memorySize interactions
	if len(interactions) > memorySize {
		interactions = interactions[:memorySize]
	}
}

// getMemoryInfo returns a formatted string of recent interactions
func getMemoryInfo() string {
	if len(interactions) == 0 {
		return "You haven't asked about any locations yet."
	}

	var result strings.Builder
	result.WriteString("Recent locations you've asked about:\n")

	// Create a map to track unique locations
	uniqueLocations := make(map[string]bool)

	for _, interaction := range interactions {
		uniqueLocations[interaction.Location] = true
	}

	// List unique locations
	result.WriteString("\nUnique locations discussed:\n")
	for location := range uniqueLocations {
		result.WriteString(fmt.Sprintf("- %s\n", location))
	}

	// Show last interaction details
	result.WriteString(fmt.Sprintf("\nMost recent query was about: %s\n", interactions[0].Location))

	return result.String()
}

// formatRecordInfo formats the record information based on its type
func formatRecordInfo(record Record) string {
	switch record.DataType {
	case "tax":
		return fmt.Sprintf("According to %s, the tax rate in %s is %s",
			record.Source, record.Location, record.Values["tax_rate"])
	case "tourist":
		return fmt.Sprintf("Tourist Information for %s (Source: %s):\n"+
			"  - Main Attractions: %s\n"+
			"  - Best Time to Visit: %s",
			record.Location, record.Source,
			record.Values["attractions"],
			record.Values["best_time"])
	case "cost":
		return fmt.Sprintf("Travel Costs for %s (Source: %s):\n"+
			"  - Average Daily Cost: $%s\n"+
			"  - Hotel: $%s per night\n"+
			"  - Food: $%s per day",
			record.Location, record.Source,
			record.Values["daily_cost"],
			record.Values["hotel_avg"],
			record.Values["food_avg"])
	default:
		return fmt.Sprintf("Information about %s: %v (Source: %s)",
			record.Location, record.Values, record.Source)
	}
}

func getAvailableLocations(records []Record) string {
	locations := make(map[string]bool)
	for _, record := range records {
		locations[record.Location] = true
	}

	var uniqueLocations []string
	for location := range locations {
		uniqueLocations = append(uniqueLocations, location)
	}
	return strings.Join(uniqueLocations, ", ")
}

// GenerateAnswer uses the OpenAI API to generate an answer
func GenerateAnswer(client *openai.Client, mainInfo, followUp, question string) (string, error) {
	if client == nil {
		if followUp != "" {
			return mainInfo + followUp, nil
		}
		return mainInfo, nil
	}

	var prompt string
	if lastQuery != "" {
		prompt = fmt.Sprintf("Previous question: %s\nCurrent question: %s\nInformation: %s",
			lastQuery, question, mainInfo)
	} else {
		prompt = fmt.Sprintf("Question: %s\nInformation: %s", question, mainInfo)
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		if followUp != "" {
			return mainInfo + followUp, nil
		}
		return mainInfo, nil
	}

	answer := resp.Choices[0].Message.Content
	if followUp != "" {
		answer += followUp
	}
	return answer, nil
}

func runQuery(cmd *cobra.Command, args []string) {
	// Load records from all data files
	var allRecords []Record
	for dataType, file := range dataFiles {
		records, err := LoadData(file, dataType)
		if err != nil {
			fmt.Printf("Warning: Error loading %s: %v\n", file, err)
			continue
		}
		allRecords = append(allRecords, records...)
	}

	if len(allRecords) == 0 {
		fmt.Println("Error: No records loaded")
		return
	}

	// Initialize OpenAI client if API key is available
	var client *openai.Client
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		client = openai.NewClient(apiKey)
	} else {
		fmt.Println("\nNote: OPENAI_API_KEY not set. Running in basic mode without AI-enhanced responses.")
	}

	fmt.Println("\nWelcome to the Travel Information System!")
	fmt.Println("Ask questions about any location (or type 'exit' to quit)")
	fmt.Println("Example: 'Tell me about California'")
	fmt.Println("You can also ask follow-up questions like 'What about New York?'")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		question := scanner.Text()
		if strings.ToLower(question) == "exit" {
			break
		}

		if question == "" {
			continue
		}

		// Find relevant information
		mainInfo, followUp := FindRelevantInfo(question, allRecords)

		// Generate answer
		answer, err := GenerateAnswer(client, mainInfo, followUp, question)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		fmt.Printf("\n%s\n", answer)

		// Store the current question for context
		lastQuery = question
	}
}
