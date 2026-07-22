---
name: investigate-codebase
description: Explore and explain an unfamiliar repository or subsystem without changing files. Use before implementation, debugging, planning, architecture review, or code review when the agent must trace entry points, call sites, data flow, tests, conventions, and ownership boundaries. Produce evidence-backed findings and explicit assumptions.
---

# Investigate Codebase

Understand the relevant system before proposing or making changes. Remain read-only unless the user explicitly changes the task.

## Set the question

Translate the request into a concrete investigation question. Define:

- the behaviour or subsystem in scope;
- the boundaries and non-goals;
- the evidence needed to answer;
- any terms that need repository-specific meaning.

## Read instructions and map the repository

1. Read applicable `AGENTS.md`, `CLAUDE.md`, README, architecture notes, and package manifests.
2. Identify languages, build systems, package managers, services, and test frameworks.
3. Find likely entry points and ownership boundaries.
4. Use file names and directory structure only as leads; confirm conclusions from code.

## Trace behaviour

Follow the relevant path end to end:

- external entry point;
- validation and authorisation;
- orchestration or business logic;
- persistence, queues, caches, and external services;
- result mapping and error handling;
- logging, metrics, and retries;
- tests that exercise the path.

Search for interfaces, implementations, call sites, configuration, feature flags, migrations, and generated code. Inspect both happy and failure paths.

## Gather evidence

Prefer:

- definitions and call sites;
- tests and fixtures;
- configuration actually loaded at runtime;
- versioned schemas and migrations;
- build or dependency metadata;
- git history only when it answers why something exists.

Do not infer current behaviour from comments alone.

## Identify constraints

Record:

- contracts and invariants;
- compatibility requirements;
- transaction and concurrency boundaries;
- security and privacy boundaries;
- performance-sensitive paths;
- project conventions;
- areas where evidence is incomplete or contradictory.

## Output

**Summary**  
Answer the investigation question directly.

**Execution path**  
Describe the relevant flow in order, naming files and symbols.

**Key constraints**  
List the contracts and boundaries a change must preserve.

**Tests and validation**  
State what currently covers the behaviour and where gaps exist.

**Unknowns and assumptions**  
Separate unknown facts from working assumptions.

**Recommended next step**  
Give one proportionate next action. Do not propose a rewrite unless the evidence requires one.
