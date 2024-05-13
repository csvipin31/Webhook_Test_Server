## 📖 Webhook Test Server

## 🛠️ Overview

The Webhook Test Server is designed to handle and simulate webhook events for various services using Go and DynamoDB. This project includes functionality for receiving webhook payloads, validating them, and storing the information in DynamoDB, making it ideal for testing and developing applications that rely on webhook data.

## Features

- `Webhook Event Handling`: Processes and responds to incoming webhook events.
- `Data Validation and Storage`: Validates incoming data and stores it in DynamoDB.
- `Health Checks`: Provides endpoints for readiness and liveliness checks of the service.

## ▶️ Getting Started

### Prerequisites

- Go (version 1.14 or later recommended)
- AWS account with DynamoDB
- Configure AWS credentials (AWS CLI or environment variables)

1. Clone the repository:

    ```bash
    git clone repo_link.git
    ```

2. Navigate to the project directory:

    ```bash
    cd root_dir
    ```

### Environment Setup

Create a `.env` file in the project root directory and provide the necessary environment variables:

        DYNAMODB_REGION=
        DYNAMODB_ORDER_TABLE_NAME=
        SERVER_PORT=8080
        DYNAMODB_ENDPOINT=

### Running the Server

To start the server, run:

``` 
go run .
```

### Running Tests

To execute tests, use:
```
go test ./...
```
