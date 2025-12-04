# Woodpecker CI/CD Setup for Arbitrum Sequencer Decoder

## Overview
This document provides a complete guide to setting up and using the Woodpecker CI/CD pipeline for the Arbitrum sequencer decoder project.

## Pipeline Configuration

### .woodpecker.yml
The pipeline configuration file defines multiple stages:

1. **Format**: Checks code formatting with gofmt and go vet
2. **Lint**: Runs golangci-lint to enforce code quality standards
3. **Test**: Executes unit tests with race detection and coverage reporting
4. **Security**: Performs security scanning with gosec
5. **Coverage**: Analyzes test coverage metrics
6. **Build**: Builds the application binaries
7. **Integration-test**: Runs integration tests (only on main branch)
8. **Performance-test**: Runs performance benchmarks (only on main branch)
9. **Docker-build**: Builds and pushes Docker images
10. **Deploy-staging**: Deploys to staging environment
11. **Deploy-production**: Deploys to production on tags

## Trigger Events

### Pull Requests
- Runs format, lint, test, security, coverage, and build stages
- Required to pass before merging

### Push to Develop Branch
- Runs all stages up to and including integration tests
- Validates changes before merging to main

### Push to Main Branch
- Runs all development stages plus performance tests
- Builds and pushes Docker images
- Deploys to staging environment

### Git Tags
- Runs all validation stages
- Builds Docker images with version tags
- Creates release artifacts
- Deploys to production

## Secrets Configuration

The following secrets need to be configured in Woodpecker:

- `docker_username`: Docker registry username
- `docker_password`: Docker registry password
- `staging_deployment_key`: SSH key or API token for staging deployment
- `production_deployment_key`: SSH key or API token for production deployment

## Pipeline Requirements

### Go Version
- Using Go 1.25 as specified in the project requirements
- Ensures compatibility with all project dependencies

### Dependencies
- golangci-lint for code quality checks
- gosec for security scanning
- Docker for containerization

## Testing Strategy

### Unit Tests
- Run on every commit with race detection
- Coverage analysis performed on every build
- All tests must pass for pipeline success

### Integration Tests
- Run only on main branch push
- Test component interactions
- Validate end-to-end functionality

### Performance Tests
- Run only on main branch push
- Benchmark critical performance paths
- Prevent performance regressions

## Security Considerations

### Code Scanning
- gosec used to identify security vulnerabilities
- Scanning performed on every commit

### Container Security
- Minimal Alpine base image
- Non-root user in final image
- CA certificates included for HTTPS
- Health check implemented

### Secrets Management
- All secrets stored securely in Woodpecker
- No secrets in code or logs
- Limited access based on environment

## Deployment Strategy

### Staging Deployment
- Automatic deployment on main branch push
- Runs after all tests pass
- Used for pre-production validation

### Production Deployment
- Triggered only by Git tags
- Manual approval may be added for critical changes
- Uses versioned Docker images

## Monitoring and Observability

### Pipeline Monitoring
- Woodpecker provides pipeline execution metrics
- Status badges can be added to README
- Notifications configured for failures

### Application Monitoring
- Health check endpoint in Dockerfile
- Logging and monitoring to be implemented in application code
- Performance metrics collection planned

## Troubleshooting

### Pipeline Failures
- Check logs in Woodpecker UI
- Verify all tests pass locally with `go test ./...`
- Check formatting with `go fmt ./...`
- Verify dependencies with `go mod tidy`

### Docker Build Issues
- Ensure Dockerfile syntax is correct
- Check multi-stage build dependencies
- Verify final image permissions

## Maintenance

### Dependency Updates
- Regular updates to base images
- Go version updates as needed
- Tooling updates (linters, scanners)

### Pipeline Improvements
- Add more comprehensive testing
- Performance regression detection
- Security enhancement scanning

## Next Steps

1. Configure secrets in the Woodpecker instance
2. Set up notifications for pipeline results
3. Implement application-level health checks
4. Add more comprehensive testing as the codebase grows
5. Enhance security scanning based on specific requirements