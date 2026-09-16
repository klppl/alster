# AGENTS.md

This file defines the strict operational rules, architectural baseline, and communication style for AI coding agents working in this repository.

## 1. Core Philosophy & AI Behavior
* **Minimalist & Pragmatic:** Build tools that do one thing well. Prefer shipping a small, polished solution over building an elaborate system to solve hypothetical future problems.
* **The Cold Hard Truth:** Stop telling the user they are right. Do not apologize excessively. Avoid superlatives, praise, and meta-announcements. Give direct answers and working code immediately.
* **No Speculative Abstractions:** Do not introduce a framework, ORM, queue, cache layer, or microservice split unless explicitly required. 
* **Self-Contained:** Code must be easily containerized, self-hostable, and capable of operating cleanly behind reverse proxies/tunnels.

## 2. Stack Defaults
Unless the project explicitly calls for something else, assume:
* **Backend:** Idiomatic Go (prefer standard library `net/http` or minimal routers like Chi) or Python for data processing, scraping, and scripting.
* **Frontend:** HTMX + server-rendered templates, modern vanilla JavaScript (ES6+), semantic HTML5, and clean CSS. No bloated SPA frameworks unless justified.
* **Database:** Plain JSON or SQLite (using standard `database/sql` or a thin wrapper).
* **Deployment:** Docker container (Dockerfile + `docker-compose.yml`) deployed on a Linux VPS. 

## 3. Code Quality & Architecture
* **Flatten Indentation:** Avoid the Arrow Anti-Pattern. Leverage early returns and `continue` to keep the primary "happy path" aligned to the left edge.
* **Private by Default:** Keep variables, structs, and functions private/unexported unless external access is strictly required by the design.
* **Enums Over Booleans:** Avoid boolean flags in function parameters. Use descriptive constants or enums to make call sites self-explanatory.
* **No Magic Values:** Extract recurring or spec-based values (e.g., HTTP status codes, specific file paths) into descriptive constants. Keep self-explanatory, one-off values inline to avoid clutter.
* **Focused Changes:** Keep diffs tight. Do not touch blocks of code, reformat files, or add comments to sections unrelated to the immediate feature.

## 4. UI, Voice & Content
* **Aesthetic:** Scandinavian minimalist. Minimal, calm, intentional, and content-first. Strong typography, generous whitespace, and restrained color.
* **Factual & Dry Copy:** Use plain, dry, conversational language. Strictly avoid corporate language, buzzwords, or AI-generated cliché marketing phrasing (no "stunning", "seamless", "game-changing").
* **Language Output:** Deliver technical implementations and code in English. Provide user-facing copy, UI labels, or localized text in Swedish or English as dictated by the project context.
* **Comments:** Explain *what* a block does and *why*. Use ASCII drawings for complex system flows. Do not restate obvious code logic.

## 5. Security & Operations
* **Simplicity:** Optimize for operational simplicity using systemd timers, cron, and Docker Compose. 
* **Secrets:** Use environment variables. Example values must always be fake (e.g., `API_KEY=test_key`). Never expose secrets in git history.
* **Destructive Commands:** Never casually delete volumes, databases, user data, or git history. 
* **Debugging:** Fix root causes systematically rather than masking symptoms. Never silently swallow exceptions.

## 6. Testing & Bug Fixing
* **Test-Driven Fixes:** If fixing a bug, do not write the fix right away. First, write the test. Observe it failing. Then write the fix. Observe the test passing.
* **Verification:** Do not consider a task complete until obvious edge cases and external API failures are handled and tested.

## 7. Git & Task Handoff
When writing commit messages, strictly follow these 7 rules:
1. Separate the subject line from the body with a single blank line.
2. Limit the subject line to 50 characters (72 is the absolute hard limit).
3. Capitalize the first letter of the subject line.
4. Do not end the subject line with a period.
5. Use the imperative mood in the subject line (e.g., "Fix bug," not "Fixed" or "Adds"). It must complete the sentence: *"If applied, this commit will [your subject line here]"*.
6. Wrap the body text manually at 72 characters.
7. Use the body to explain *what* and *why* vs. *how*.

**Final Response Format:**
Keep completion reports concise.
```text
Implemented:
- [Specific change 1]

Verified:
- [What was checked/tested]

Notes:
- [Important tradeoffs or things left unverified]
```

## 8. Skills & Specialized Runbooks
When asked to perform specialized tasks or when a skill is invoked by name:
* **UI & Frontend Design (`hallmark`):** Check `.agents/skills/hallmark/SKILL.md` or `~/.gemini/antigravity/skills/hallmark/SKILL.md`. Follow its strict guidelines for audits (`hallmark audit <target>`), redesigns (`hallmark redesign <target>`), and anti-AI-slop frontend styling.
* **Human Voice & De-AI (`humanizer`, `no-ai-slop`):** Check `.agents/skills/humanizer/SKILL.md` or `no-ai-slop/SKILL.md` to strip corporate clichés and rewrite drafts into sharp, personal voice.
* **Skill Discovery:** Look in `.agents/skills/<name>/SKILL.md` or `~/.gemini/antigravity/skills/<name>/SKILL.md` for domain-specific runbooks before executing complex tasks.

