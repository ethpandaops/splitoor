# Splitoor

Splitoor is a comprehensive Ethereum monitoring and management tool that monitors 0xSplits contracts, validators, and Safe transactions while providing CLI utilities for managing 0xSplits v1 contracts.

## Project Structure
Claude MUST read the `.cursor/rules/project_architecture.mdc` file before making any structural changes to the project.

## Code Standards  
Claude MUST read the `.cursor/rules/code_standards.mdc` file before writing any code in this project.

## Development Workflow
Claude MUST read the `.cursor/rules/development_workflow.mdc` file before making changes to build, test, or deployment configurations.

## Component Documentation
Individual components have their own CLAUDE.md files with component-specific rules. Always check for and read component-level documentation when working on specific parts of the codebase.

## Quick Start for Development

### Building the Project
```bash
go build -o splitoor .
```

### Running Tests
```bash
go test ./...
```

### Running the Monitor
```bash
./splitoor monitor --config example_config.yaml
```

### Key Commands to Remember
- **Linting**: Run `golangci-lint run` before committing
- **Testing**: Run `go test -race ./...` to catch race conditions
- **Coverage**: Run `go test -coverprofile=coverage.out ./...` to check test coverage

## Important Considerations

### When Working with Ethereum Nodes
- Always handle connection failures gracefully
- Use the connection pool for load balancing
- Include appropriate timeouts for RPC calls

### When Adding New Features
- Expose relevant Prometheus metrics
- Add appropriate logging at different levels
- Write comprehensive tests including edge cases
- Update configuration structures if needed

### When Modifying Monitoring Logic
- Consider the impact on notification frequency
- Test with different event scenarios
- Ensure metrics are updated correctly
- Document any new event types

## Testing Guidelines
- Mock external dependencies (Ethereum nodes, APIs)
- Use table-driven tests for multiple scenarios
- Test error conditions thoroughly
- Verify metrics are recorded correctly