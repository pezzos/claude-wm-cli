# /2-Test-design
Design comprehensive testing strategy with MCP-powered UI testing capabilities.

## Pre-Design Intelligence (MANDATORY)
1. **Load Test Patterns**: Use `mcp__mem0__search_coding_preferences` to find similar testing approaches
2. **Analyze UI Components**: Use `mcp__consult7__consultation` to identify UI elements requiring testing
3. **Get Testing Documentation**: Use `mcp__context7__` for current testing framework best practices

## Test Design Steps
1. **Enhance TEST.md**: Generate comprehensive test scenarios (template pre-populated by preprocessing)
2. **Unit & Integration Tests**: Define traditional testing approaches
3. **UI Automation Tests**: Design MCP-powered browser testing when UI components present
4. **Test Data & Validation**: Plan comprehensive data requirements
5. **Failure Scenarios**: Design error handling and edge case tests

## MCP UI Testing Integration (When Applicable)
- **Playwright Tests**: For React/Vue/Angular applications use `mcp__playwright__browser_*` tools
- **Puppeteer Tests**: For Node.js applications use `mcp__puppeteer__puppeteer_*` tools  
- **Visual Regression**: Design screenshot-based tests for UI consistency
- **Cross-Browser Testing**: Plan automated testing across different browsers
- **Performance Testing**: Include Core Web Vitals and loading time validation
- **Accessibility Testing**: Automated a11y validation in test suite

## Test Categories to Include
- **Manual Tests**: Critical user journeys requiring human validation
- **Automated Unit Tests**: Function-level testing
- **Automated Integration Tests**: Component interaction testing  
- **Automated UI Tests**: Browser-based interaction testing (MCP-powered)
- **Performance Tests**: Load and stress testing scenarios
- **Security Tests**: Input validation and vulnerability testing

## Important
Cover happy path, edge cases, and error conditions. Design tests before implementing. Include MCP UI automation for web interfaces to enable continuous regression testing.

# Exit codes:
- 0: Success
- 1: Needs iteration
- 2: Blocked
- 3: User input needed