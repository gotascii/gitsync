---
name: project-management
description: >
  Load project management system for discovery, planning, and execution.
  Use when user mentions "project" or "projects" in any context:
  searching, pending, backlog, creating, updating, executing, categorizing, archiving.
---

# Project Management System

This skill handles ALL interactions with project files including discovery, planning, execution, and retrospectives.

## Project System Rules

Mandatory rules for project planning, documentation and execution in `docs/projects/`.

### Overview

```text
Repo → Project → Issue(s) → Steps + Tests
```

**Components:**

- **Project**: Single scope with target architecture (file: `docs/projects/name.md`)
- **Issue**: Work item with deliverable (header: `### [ ] I1: Description`)
- **Step**: Implementation action (checkbox under Steps section)
- **Test**: Verification activity (checkbox under Tests section)

**Lifecycle**: Active in `docs/projects/`, completed moved to `docs/projects/completed/[N]-name.md`

### Planning

Activities performed once when creating a project plan.

#### File Naming and Location

**Active projects:**

- Location: `docs/projects/`
- Format: `name.md` (use kebab-case)
- Example: `docs/projects/backup-automation.md`

#### Template Structure

- Conform to the Project Plan Template below
- **Overview**: Brief description of project goal
- **Relevant System Architecture Before Project Started**: Immutable baseline - what exists now
- **Target Architecture**: Immutable baseline - what should exist after
- **Issues**: Work items with Steps and Tests

#### Issue Design

**Steps vs Tests:**

- Steps = implementation actions
- Tests = verification activities

**Examples:**

*Config-only issue:*

```markdown
#### Steps
- [ ] Create config.yaml

#### Tests
- [ ] File exists with valid YAML
```

*Deployable issue:*

```markdown
#### Steps
- [ ] Create config.yaml:
  ```yaml
  service:
    name: foo
    port: 8080
  ```

- [ ] Deploy: helm install foo

#### Tests

- [ ] Helm release deployed
- [ ] Pods healthy
- [ ] Service accessible

```

#### Diagrams

Use mermaid tool for architecture diagrams when available.

### Execution

Activities performed iteratively while working through a project plan.

#### Core Loop

For each unchecked item in a project plan:

1. **Read** next unchecked Step/Test
2. **Execute** the action
3. **Update** mark checkbox `[x]` immediately
4. **Repeat** until Issue complete

**Example:**

```text
Read: "- [ ] Deploy app"
Execute: helm install ...
Update: "- [x] Deploy app"
```

#### Constraints

- **NO TodoWrite tool** - use project plan checkboxes only
- **NO batch updates** - mark after EACH completion
- **Sequential completion** - Steps → Tests → Issue header

#### Marking Items Complete

**Steps and Tests:**
Update checkboxes immediately after completing each action.

```markdown
- [ ] Deploy service
```

becomes

```markdown
- [x] Deploy service
```

#### Marking Issues Complete

Issue header checkbox `[x]` ONLY when:

- ALL Steps are checked `[x]`
- AND ALL Tests are checked `[x]`

If any Step or Test is incomplete, the issue remains `[ ]`.

**Example:**

1. Complete all Steps checkboxes
2. Complete all Tests checkboxes
3. Edit the issue heading line: change `### [ ] I1:` to `### [x] I1:`
4. **DO NOT add a new heading line**

**CRITICAL**: Update the EXISTING heading checkbox. DO NOT add a duplicate heading.

**WRONG:**

```markdown
### [ ] I1: Deploy service

#### Steps
- [x] Create manifest
- [x] Apply manifest

#### Tests
- [x] Pod running

### [x] I1: Deploy service    <-- DUPLICATE HEADING - WRONG!
```

**CORRECT:**

```markdown
### [x] I1: Deploy service    <-- UPDATE EXISTING HEADING

#### Steps
- [x] Create manifest
- [x] Apply manifest

#### Tests
- [x] Pod running
```

### Project Completion

When all issues in a project are complete:

1. Move file from `docs/projects/name.md` to `docs/projects/completed/[N]-name.md`
2. `[N]` is the next sequential number in the completed directory
3. Update any references to the project file path in documentation

## Project Plan Template

```markdown
# [Project Name]

## Overview

Thorough description of project scope, goals, and business value.

## Relevant System Architecture Before Project Started

High-level architectural overview of **relevant** existing infrastructure at project start. Only include components, services, and architectural elements that are directly related to this project's scope. Focus on what exists that this project will build upon, integrate with, or replace. Omit unrelated infrastructure.

## Target Architecture

​```text
Mermaid diagram of desired end state
​```

## Issues

### [x] I1: Full description of the issue or goal

Detailed description of the deliverable or desired end state. Can include sub-sections to capture relevant implementation details if necessary.

Example:

**Update Service Configuration:**

- An over-arching aspect of the new service config
- Another high-level aspect of the new service config

#### Details About the System Prior to the Start of this Project that are Relevant to this Issue

Detailed description of pre-existing implementation as it exists at project start, specific to this issue. This can include code patterns, file locations, configurations, and other specifics that are going to be impacted or changed by the work involved with this specific issue. Can be broken into sections if considerable pre-existing implementation exists. The goal here is that we can look back when this project is complete and see a summary of what things looked like before we made all the changes associated with this issue. Do not include details about this issue's implementation. Those go in the previous section.

Example:

**Pre-existing Service Configuration:**

- An aspect of pre-existing service config
- Another aspect of pre-existing service config

#### Steps

- [x] Self-contained implementation action (include full details inline: code blocks, configs, exact content to add)
- [x] Another self-contained implementation action

#### Tests

- [x] A succinct, verifiable test that confirms some aspect of the implementation is correct
- [x] Another succinct, verifiable test that confirms some aspect of the implementation is correct

### [x] I2: Repeat as necessary
```

## Project Lifecycle States

**Active**: `docs/projects/*.md` - currently planning or executing
**Backlog**: `docs/projects/backlog/*.md` - defined but not started
**Completed**: `docs/projects/completed/[N]-name.md` - archived with sequential number
**Abandoned**: `docs/projects/abandoned/*.md` - cancelled or superseded

## Discovery Commands

When searching or categorizing projects:

```bash
# List active projects
ls docs/projects/*.md

# List backlog
ls docs/projects/backlog/*.md

# Search project content
grep -r "keyword" docs/projects/

# Count by status
find docs/projects -name "*.md" | wc -l
```

## Critical Reminders

- **NO TodoWrite tool** during project execution - use project plan checkboxes only
- Mark checkboxes immediately after each completion
- Issue header checkbox `[x]` only when ALL steps AND tests complete
- Never duplicate issue headings when marking complete
