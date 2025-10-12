window.MERMAID_ASCII_EXAMPLES = [
  {
    id: "simple-flow",
    name: "Simple Flow",
    mermaid: "graph LR\nA[Start] --> B[Do thing]\nB --> C{Decision}\nC -->|Yes| D[Celebrate]\nC -->|No| E[Retry]",
  },
  {
    id: "process",
    name: "Process",
    mermaid:
      "graph TD\nstart([Start]) --> qa[Quality Check]\nqa -->|Pass| ship[Ship]\nqa -->|Fail| fix[Fix]\nfix --> qa",
  },
  {
    id: "deployment",
    name: "Deployment Pipeline",
    mermaid:
      "graph LR\nDev((Dev)) --> CI[CI]\nCI --> Build[Build]\nBuild --> Deploy{Deploy?}\nDeploy -->|Yes| Prod((Prod))\nDeploy -->|No| Dev",
  },
  {
    id: "sequence",
    name: "Sequence",
    mermaid:
      "graph TD\nUser((User)) -->|Request| API[API]\nAPI --> DB[(DB)]\nDB --> API\nAPI -->|Response| User",
  },
];
