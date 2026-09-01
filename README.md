# Agent Harness

An agent harness is a runtime framework that connects an LLM to external tools and capabilities. It provides a layer where an agent can:

- **Receive instructions** from a user or an application.
- **Plan and orchestrate** steps to complete a task.
- **Invoke tools** (e.g. calculators, file access, shell commands) to gather data or take actions.
- **Return results** back to the caller in a structured way.

Think of it as the scaffolding that turns a language model into an autonomous agent: the model supplies the intelligence, while the harness supplies the environment, the tool interface, and the execution loop that ties it together.