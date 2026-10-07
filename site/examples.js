window.MERMAID_ASCII_EXAMPLES = [
  {
    id: "node-shapes",
    name: "Node Shapes (extended mode)",
    mermaid: "graph LR\nStart([Start]) --> Check{Ready?} --> Run[[Run]] --> DB[(DB)] --> Finish(Done)",
  },
  {
    id: "colored-flow",
    name: "Colored Flow (extended mode)",
    mermaid: "graph LR\nclassDef box color:#3366cc,stroke:#00aa66\nA[Start] -->|two words| B[Finish]\nclass A,B box\nlinkStyle 0 stroke:#cc3344,color:#9966cc",
  },
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
