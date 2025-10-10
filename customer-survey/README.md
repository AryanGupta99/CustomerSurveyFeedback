# Customer Survey Application

## Overview
The Customer Survey Application is a utility designed to gather feedback from users through a simple survey form. The application prompts users with a popup asking if they would like to participate in a survey that takes approximately 10 seconds to complete. The survey consists of three questions rated on a scale of 1 to 10.

## Project Structure
```
customer-survey
├── cmd
│   └── survey
│       └── main.go          # Entry point of the application
├── internal
│   ├── ui
│   │   ├── popup.go         # Logic for displaying the popup
│   │   └── form.go          # Structure and behavior of the survey form
│   └── survey
│       ├── handler.go       # HTTP handler functions for form submission
│       └── questions.go     # Definitions of survey questions
├── pkg
│   └── model
│       └── response.go      # Data model for survey responses
├── configs
│   └── config.yaml          # Configuration settings for the application
├── go.mod                   # Go module file
├── Makefile                 # Build instructions and commands
└── README.md                # Documentation for the project
```

## Setup Instructions
1. Clone the repository to your local machine.
2. Navigate to the project directory.
3. Run `go mod tidy` to install the necessary dependencies.
4. Configure the `config.yaml` file with your API keys and endpoint URLs for Zoho Forms.
5. Build the application using the Makefile or run `go run cmd/survey/main.go` to start the application.

## Usage
Upon launching the application, users will see a popup asking if they would like to fill out the survey. If they agree, they will be presented with a form containing three questions to rate on a scale of 1 to 10. Once completed, the responses will be processed and sent to the designated endpoint.

## Contributing
Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.