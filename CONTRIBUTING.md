# Contributing to KubeArchInspect Docker Extension

Thank you for your interest in contributing! This document provides guidelines and steps for contributing.

## Development Setup

### Prerequisites

- Docker Desktop 4.8.0+
- Go 1.21+
- Node.js 18+
- A Kubernetes cluster (kind, minikube, or Docker Desktop's built-in Kubernetes)

### Setting Up the Development Environment

1. **Clone the repository**
   ```bash
   git clone https://github.com/ajeetraina/kubearchinspect-docker-extension.git
   cd kubearchinspect-docker-extension
   ```

2. **Install backend dependencies**
   ```bash
   cd backend
   go mod download
   ```

3. **Install frontend dependencies**
   ```bash
   cd ui
   npm install
   ```

4. **Build and install the extension**
   ```bash
   docker build --tag=kubearchinspect:dev .
   docker extension install kubearchinspect:dev
   ```

5. **Enable hot-reload for UI development**
   ```bash
   cd ui
   npm run dev
   # In another terminal:
   docker extension dev ui-source kubearchinspect:dev http://localhost:5173
   ```

## Code Style

### Go Backend

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing

### TypeScript Frontend

- Use TypeScript strict mode
- Follow React functional component patterns
- Use Material-UI components consistently

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test locally with Docker Desktop
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to your fork (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Reporting Issues

When reporting issues, please include:

- Docker Desktop version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Any error messages

## Feature Requests

Feature requests are welcome! Please describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Questions?

Feel free to open an issue for any questions about contributing.
