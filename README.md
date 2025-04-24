# GoragAgent - AI-Powered Question Answering System

## Overview
GoragAgent is an intelligent question-answering system that leverages Language Models (LLMs) to provide accurate, context-aware responses to user queries about tax information, tourist attractions, and travel costs for different locations.

## Prerequisites
- Go 1.21 or higher (only for building from source)
- OpenAI API key (for AI-enhanced responses)
- Git (only for building from source)

## Quick Start

### Option 1: Using Pre-built Binary
If you downloaded the pre-built binary:

1. Set up your OpenAI API key:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

2. Run the binary:
```bash
./bin/goragagent query
```

Note: Without the API key, the system will run in basic mode without AI-enhanced responses.

### Option 2: Building from Source

1. Clone the repository:
```bash
git clone https://github.com/glebsonmac/goragagent
cd goragagent
```

2. Install dependencies:
```bash
go mod download
go mod tidy
```

3. Set up your OpenAI API key:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

4. Build the project:
```bash
go build -o bin/goragagent
```

## Running the Application

### Interactive Mode
Start an interactive session where you can ask questions about locations:
```bash
./bin/goragagent query
```

Example queries:
```bash
> Tell me about California
> What's the tax rate in Travis County?
> What are the tourist attractions in New York?
> How much does it cost to visit Miami?
```

### Using Data Files
You can specify custom data files:
```bash
./bin/goragagent query --data path/to/your/data.csv
```

Default data files are located in the `data/` directory:
- `data/state_taxes.csv`: Tax information
- `data/tourist_info.csv`: Tourist attractions and best times to visit
- `data/travel_costs.csv`: Travel costs and expenses

## Running Tests

To run all tests (unit, integration, and security tests):
```bash
go test ./... -v
```

Note: Integration tests require an OpenAI API key to be set in the environment.

## Data File Format
The application accepts CSV files with the following format:

```csv
location,tax_rate,source
Travis County,1.9%,tax_policies_2023.pdf
Williamson County,2.1%,tax_records_2023.csv
```

## Project Structure
```
goragagent/
├── bin/           # Pre-built binary
├── cmd/           # Command implementations
├── data/          # CSV data files
├── tests/         # Test suites
│   ├── unit/     # Unit tests
│   ├── integration/  # Integration tests
│   └── security/    # Security tests
└── main.go       # Application entry point
```

## Error Handling
- The application validates all input files for security
- Invalid file paths or formats will result in clear error messages
- Rate limits and API errors are handled gracefully

## Future Improvements

### Data Validation and Error Handling
- Improve CSV format validation to provide more specific error messages
- Enhance error messages for invalid data formats to be more user-friendly
- Standardize error message format across all validation checks

### Search and Response Improvements
- Enhance the search functionality for empty queries and non-matching results
- Improve the "No Match" response to provide more helpful suggestions
- Add fuzzy matching for location names to handle typos and variations
- Implement better handling of empty queries with context-aware suggestions

### Test Coverage
The following test improvements are planned:
- Refine data format validation tests to match exact error messages
- Enhance search functionality tests for edge cases
- Add more comprehensive tests for special characters and international locations
- Improve test coverage for the query response system

### Additional Features
- Multi-source dataset integration
- Conversational memory for follow-up questions
- Enhanced vector similarity search
- Web interface using Go's standard http package or frameworks like Gin
- Caching mechanism for faster responses
- More sophisticated prompt engineering
- Additional data sources and formats support
- gRPC API support

## Contributing
Contributions are welcome! Please feel free to submit issues and pull requests.

## License
MIT License - see LICENSE file for details

