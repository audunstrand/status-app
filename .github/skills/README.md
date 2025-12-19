# GitHub Copilot Agent Skills

This directory contains agent skills adapted from [claude-skillz](https://github.com/NTCoding/claude-skillz) 
for use with GitHub Copilot.

## Available Skills

- **tdd-process**: Strict test-driven development with red-green-refactor cycles
- **writing-tests**: Best practices for writing effective unit tests
- **task-workflow**: Persistent task tracking and session management
- **switch-persona**: Switch between different AI personas mid-conversation
- **implementation-analysis**: Analyze code flow before implementing to prevent guessing
- **design-analysis**: Identify refactoring opportunities across design dimensions
- **software-design-principles**: Object-oriented design principles and best practices
- **critical-peer-personality**: Professional skeptical communication style that coaches
- **independent-research**: Research and explore topics independently before asking
- **concise-output**: Keep responses concise and focused
- **observability-debugging**: Debug-first approach with logging and observability
- **data-visualization**: Create effective data visualizations and charts
- **confidence-honesty**: Honest communication about confidence levels
- **questions-are-not-instructions**: Don't treat questions as instructions to execute
- **create-tasks**: Break down features into actionable task lists

## How to Use

These skills are automatically loaded by GitHub Copilot when:

1. **In VS Code Insiders**: Enable `chat.useAgentSkills` in settings
2. **With Copilot CLI**: Skills load automatically when invoking agents
3. **With Copilot Coding Agent**: Skills are used when working on repository issues

## Skill Structure

Each skill is in its own directory with a `SKILL.md` file:

```
.github/skills/
├── tdd-process/
│   └── SKILL.md
├── software-design-principles/
│   └── SKILL.md
└── ...
```

## Credits

Original skills by [NTCoding](https://github.com/NTCoding/claude-skillz).
Adapted for GitHub Copilot Agent Skills format.

## Learn More

- [GitHub Copilot Agent Skills Documentation](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)
- [Original claude-skillz Repository](https://github.com/NTCoding/claude-skillz)
