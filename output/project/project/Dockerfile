{
  "files": {
    "Dockerfile": "FROM golang:1.21-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o main .\nEXPOSE 8080\nCMD [\"./main\"]",
    "docker-compose.yml": "version: '3.8'\nservices:\n  app:\n    build: .\n    ports:\n      - \"8080:8080\""
  }
}