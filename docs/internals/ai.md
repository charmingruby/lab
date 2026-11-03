# AI capabilities

Conventions for developer-defined AI capabilities inside a domain, following the [domain spine](./architecture.md#domain-structure). Examples are illustrative, not existing APIs; implement only the components the task requires.

## Choosing the capability

- Keep AI behavior self-contained in the module that owns the task. Multiple agents in one module do not justify moving code to `internal/shared/`.
- Agents are defined in code. Dynamic configuration or discovery needs a separate product requirement.
- An agent belongs to the module that owns its task, including coordination tasks subject to [cross-module boundaries](#cross-module-boundaries).

| Requirement | Starting point |
| --- | --- |
| Generate or classify with one model response | Usecase + prompt + inference port |
| Let the model request tools over several rounds | Usecase + configured agent runtime + tools |
| Expose application capabilities to an MCP client | MCP delivery handlers + usecases; no model required |
| Consume tools from an external MCP server | Outbound port + MCP adapter |

## Vocabulary

| Concept | Responsibility |
| --- | --- |
| Agent definition | Instructions, available tools, model selection, and execution limits for a task |
| Runtime | Executes inference and the tool-calling loop; holds no business rules |
| Prompt | Instructions/template and its rendering with explicit input |
| Tool | A capability with a name, input contract, and execution function |
| Tool call | One model-requested invocation of a tool; not a new tool definition |
| Inference provider | External service translating messages into model output |
| MCP server | Inbound protocol exposing tools, resources, or prompts |
| MCP client | Outbound integration consuming capabilities from another server |

## Module layout

Example with HTTP and MCP entry points, two agents, and two inference providers:

```text
internal/<domain>/
├── <domain>.go                   # composition root
├── delivery/
│   ├── http/endpoint/
│   └── mcp/                     # inbound handlers and protocol DTOs
├── usecase/
├── model/
├── repository/
├── public/
├── ai/
│   ├── runtime/
│   │   ├── runner.go            # inference/tool loop
│   │   ├── agent.go             # agent configuration types
│   │   └── tool.go              # tool execution contract
│   ├── agent/
│   │   ├── triage.go            # task-specific definitions
│   │   └── resolution.go
│   ├── tool/
│   │   ├── get_record.go        # adapters to application capabilities
│   │   └── update_record.go
│   └── prompt/
│       ├── prompt.go            # rendering and embedded templates
│       ├── triage.md
│       └── resolution.md
└── client/
    ├── inference.go             # domain-owned inference port and DTOs
    ├── inference/
    │   ├── openai.go
    │   ├── anthropic.go
    │   └── memory.go
    ├── external_tools.go        # only if external tools are consumed
    └── external_tools/
        └── mcp.go
```

`ai/` groups separate packages, not a mandatory facade. Each agent definition selects its prompts, tools, model, and limits; agents may reuse prompts and tools without sharing execution state.

Providers and outbound MCP follow [external-integrations.md](./external-integrations.md), including raw SDK clients and test fakes. SDK and MCP DTOs stay at adapter/delivery boundaries; define only the inference features the application needs.

## Dependencies and composition

| Package | Allowed collaborators for this capability |
| --- | --- |
| `usecase` | Repositories, model, client ports, `ai/prompt`, and a configured concrete `ai/runtime` runner |
| `ai/runtime` | Domain inference port and tool contracts; never usecases, repositories, agent definitions, or concrete tool adapters |
| `ai/agent` | Runtime configuration, prompts, and supplied tools; no database or SDK access |
| `ai/tool` | Tool contracts and the action usecases implementing each business capability |
| `ai/prompt` | Templates and explicit rendering input; no usecases, repositories, or provider calls |
| `client` (port package) | Owns provider-neutral boundary DTOs; does not import `ai/*` or SDK packages |

The runtime's tool contract is an execution callback, not an interface around a golden-source repository or usecase. Keep the runner concrete; do not introduce service interfaces just to mock it.

Business operations, local or external, go through action usecases. Tool adapters translate arguments/results; outbound adapters translate protocols. Neither decides business authorization or invariants.

Wire in `<domain>.go`. Keep the initiating operation separate from the action instance used by tools when they would otherwise require each other. For example, retain the existing `usecase.Usecase` for ordinary actions and add a concrete `usecase.Triage` for agent execution in the same package:

```go
// Illustrative composition; repository/provider initialization omitted.
actions := usecase.New(recordRepo, txManager, notifier)
getRecord := tool.NewGetRecord(actions)
updateRecord := tool.NewUpdateRecord(actions)
definition := agent.NewTriage(getRecord, updateRecord)
runner := runtime.New(provider, definition)
triage := usecase.NewTriage(recordRepo, runner)
// Inject triage and actions into their respective delivery handlers.
```

`agent.NewTriage` assembles the prompt, allowed tools, model, and limits. `usecase.Triage` owns request validation, authorized context loading, invocation, and outcome mapping. It imports `ai/runtime`, not `ai/agent` or `ai/tool`; those packages are assembled in the composition root. `client` DTO ownership prevents a second cycle between the inference port and runtime types.

This keeps both imports and construction acyclic without changing non-AI usecases. Do not solve cycles with global registries or mutable late wiring. Do not expose the initiating operation as its own tool; agent-to-agent delegation needs an explicit bounded design.

## Execution flows

### Single inference

The entry point calls a usecase. The usecase validates access, loads context, renders a prompt with `ai/prompt`, calls the inference port, validates the required result shape, and maps the outcome. No runner or tools are necessary.

### Agent with tools

1. The entry point delegates to the initiating usecase.
2. The usecase validates the request, loads authorized context, and starts the configured runner with execution-specific input and constraints.
3. The runner calls the inference provider. The response is either a final answer or tool calls.
4. The runner dispatches tool calls through adapters to the responsible action usecases, applying the [execution guarantees](#execution-guarantees).
5. The runner returns tool results to the model and repeats within its limits.
6. The initiating usecase handles the final outcome, persists required state, and maps errors for delivery.

### MCP server

An external MCP client calls an inbound handler, which delegates to a usecase and translates the result into the protocol response. This includes application tools, prompts, and resources. HTTP, MCP, and local agent tools may call the same action usecase directly; they do not call each other's handlers or use loopback MCP requests.

Use `mcp/` when it is the only inbound mechanism, or `delivery/mcp/` beside `delivery/http/` when both exist. Follow the [delivery mechanism layout](./architecture.md#delivery-mechanism-layout). Exposing MCP does not imply running an agent; an MCP action may optionally start one through its usecase.

### MCP client

When an agent needs an external MCP tool, its action usecase authorizes the operation and invokes a consumer-owned outbound port implemented by the MCP adapter. The runtime receives the result through the local tool contract. This is outbound traffic and does not trigger `delivery/`.

## Prompts and models

- Start with templates versioned in the module and embedded in the binary. Rendering takes explicit data; it does not load its own context from persistence.
- Usecases and agent definitions may render prompts independently; prompts do not need an agent.
- Editable/published prompts, if required later, use domain models, usecases, and the concrete repository. Do not add a persistence port for the application's own database.
- Choose fixed models/providers in composition. If selection varies by tenant or task, the usecase owns that policy and passes the allowed selection into execution.
- Keep provider capabilities explicit. Reject unsupported requirements instead of silently dropping tools, output constraints, or other requested behavior.

## Execution guarantees

- Treat model output, retrieved content, and external tool results as untrusted input. Prompt instructions are not an authorization boundary.
- The application chooses the tools available for each run. Propagate trusted actor/tenant context to action usecases; never accept model-provided identity as authority.
- The runner rejects unknown or unavailable tools; tool adapters decode and validate arguments before invoking action usecases. Require confirmation when the action's policy calls for it.
- Bound model iterations, total tool calls, and elapsed time. Propagate `context.Context` cancellation through providers and tools. Define the outcome when a limit is reached or a provider/tool fails; never silently treat an incomplete run as success.
- Keep conversation history, tool results, counters, and other mutable state isolated per execution. Do not store run state on reusable agent definitions.
- Do not hold a database transaction open across inference or external tool calls. Each action owns its transaction; a sequence of successful tool calls is not one atomic operation.
- Do not retry mutating tools automatically without an explicit idempotency policy. Provider retries must be bounded and must not replay completed actions accidentally.
- Execute tool calls sequentially unless ordering, concurrency, and side effects are explicitly defined otherwise.
- Record agent identity, execution outcome, model usage where available, and tool outcomes using existing observability conventions. Do not log full prompts, secrets, or sensitive tool arguments/results by default.

## Cross-module boundaries

Reads follow [cross-module-reads.md](./cross-module-reads.md). Commands between modules have no approved boundary yet; request an architectural decision before implementing them. AI does not grant access to another module's internals.

## Testing and verification

Conventions, ordinary usecase/delivery coverage, and verification commands belong to [testing.md](./testing.md). Additional AI-specific cases:

- **Prompts:** deterministic rendering, required inputs, and relevant template variants; no provider calls.
- **Runtime:** final answers, multiple tool rounds, unknown/unavailable tools, invalid arguments, limits, cancellation, provider/tool failures, and state isolation. Supply inference through a real or stateful in-memory adapter following [external-integrations.md](./external-integrations.md#testing-an-integration).
- **Tools:** verify argument/result translation and rejection of unauthorized actions through real action usecases.
- **Agent definitions:** prove relevant prompt/tool/limit configuration through focused behavior tests; one agent must not accidentally acquire another's tools.
- **Provider/MCP adapters:** verify message, tool-call, and error translation, including rejection of unsupported provider capabilities. Live paid-provider calls are not a prerequisite for deterministic tests.

## Implementation checklist

1. Choose the owning module and capability using the opening table.
2. Define usecase outcomes, required prompts/tools, and provider contracts.
3. Wire the components following the acyclic composition example.
4. Apply the execution guarantees and test the affected behavior.
