### RESTful API Development with Go and Redis

The primary goal of this project is to facilitate an in-depth understanding of basic Back-End operation principles and the RESTful API architecture. Utilizing the Go programming language, an HTTP server was constructed, and an extensive introduction to the API development process through RESTful API architecture was achieved. Redis was chosen as the database for its swift data processing and flexible data management capabilities, enhancing the overall performance and efficiency of the project.

The project essentially encompasses a RESTful API that allows users to register, log in, retrieve and update their information, and submit match scores. These scores are then processed and integrated into a leaderboard, all of which are stored in Redis for efficient data management.

### Getting Started

This section outlines the prerequisites and installation steps to set up and run the project on your local machine.

#### Prerequisites

List of tools and technologies required and their installation guide:

```
- Brew (for macOS users)
- Go (programming language)
- Redis (database)
- Postman (for sending GET and POST requests)
```

#### Installation

1. Clone the project repository:

```
git clone https://github.com/Dzdrgl/redis-Api.git
cd redis-Api
```

2. Install dependencies (Go modules and other packages):

Ensure you have Brew installed, then use it to install Go and Redis quickly:

```
brew install go
brew install redis
```

Then,

```
go mod download
```

3. Start the application:

```
go run .
```

### Usage

Use Postman to send requests for testing the functionalities of the API.

### Development

#### Future Enhancements

Features and improvements to be added to the project in the future:

- Implementing JWT or OAuth for enhanced authentication.
- Database optimization and comprehensive database integration.
- Customized error handling for an improved user experience.
