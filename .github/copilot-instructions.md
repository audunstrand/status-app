# Development Workflow

## Test-Driven Development (TDD)
1. **Write tests first**: Update or create tests for the new functionality before writing code
2. **Run tests**: Verify tests fail appropriately
3. **Implement code**: Write minimal code to make tests pass
4. **Run tests again**: Verify tests now pass
5. **Commit**: If tests pass, commit the changes

## Code Quality
- **Make the change easy, then make the easy change**: Do necessary analysis and tidying before adding functionality
- **Run tests after every meaningful change**: Don't accumulate untested code

## Deployment
- **Verify deployments**: After pushing to git, check GitHub Actions status to confirm deployment succeeded
