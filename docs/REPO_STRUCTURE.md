# Repository Structure

## What's in Git vs What's Excluded

### ✅ TRACKED IN GIT (Source Code & Essential Docs)

```
/
├── todo-cmdline/               # CLI application (Go)
├── todo-server/                # REST API (Go)
├── todo-mobile/                # Mobile app (React Native)
├── shared-types/               # Shared type definitions
├── integration-tests/          # Cross-component tests
│
├── CLAUDE.md                   # AI guidelines for development
├── PLAN.md                     # Project improvement roadmap
├── TESTING.md                  # API testing guide
├── CONFIG.md                   # CLI configuration reference
├── README.md                   # (when created)
├── .gitignore                  # Git exclusions
└── .gitattributes              # (when created)
```

### ❌ EXCLUDED FROM GIT (Analysis & Planning Docs)

```
docs/claude-ai-guides/         # (in .gitignore)
├── COMPREHENSIVE_PROJECT_STATUS.md
├── CI_CD_SETUP_GUIDE.md
├── GITHUB_ACTIONS_QUICKSTART.md
├── GIT_STRUCTURE_PLAN.md
├── GIT_MONOREPO_SETUP_CHECKLIST.md
├── GIT_SETUP_SUMMARY.md
├── ARCHITECTURAL_REFACTORING_GUIDE.md
├── CODE_AUDIT_SUMMARY.md
├── PHASE_*.md (multiple phase reports)
├── SESSION_SUMMARY.md
├── FINAL_SESSION_SUMMARY.md
└── README.md (index of all Claude docs)
```

---

## Why This Structure?

### ✅ In Git
- **Source Code** - What you need to build/run the app
- **Core Docs** - CLAUDE.md, PLAN.md, TESTING.md
- **Configuration** - How to configure the application

### ❌ Not in Git
- **Analysis Documents** - Decision analysis, alternative approaches
- **Phase Reports** - Internal tracking of what was accomplished
- **Implementation Guides** - How to set things up (not needed after setup)
- **Session Summaries** - Development process notes
- **Planning Notes** - Thinking behind architecture decisions

---

## What Developers Need

### To Clone & Run
```bash
git clone https://github.com/your-org/commandlinetodo.git
cd commandlinetodo
# Has everything needed!
```

They get:
- ✅ All source code (3 components)
- ✅ Tests (335+ tests)
- ✅ Core documentation (CLAUDE.md, TESTING.md)
- ✅ Configuration guides

They DON'T need:
- ❌ Analysis docs (already decided)
- ❌ Phase reports (already completed)
- ❌ Planning guides (already implemented)
- ❌ Setup guides (already setup)

---

## Finding Documentation

### For Understanding the Project
Read these **in the repository**:
1. CLAUDE.md - AI development guidelines
2. PLAN.md - Project roadmap
3. README.md (root) - Project overview
4. TESTING.md - API testing guide

### For Deep Analysis (Reference Only)
Available in **docs/claude-ai-guides/** (not in Git):
- COMPREHENSIVE_PROJECT_STATUS.md - Complete status report
- Phase reports - What was done and why
- Implementation guides - How it was set up

---

## Repository Size Impact

### Before Organization
- Git repo size: ~X MB (if all docs included)
- Clone time: Slower
- Repository noise: High (many analysis files)

### After Organization
- Git repo size: ~Y MB (just source)
- Clone time: Faster
- Repository noise: Low (focused on code)

**Result:** Cleaner, faster, more professional repository

---

## Adding New Files

### New Source Code?
→ Place in appropriate component directory
→ Will be tracked automatically

### New Documentation for Developers?
→ Place in root or appropriate docs/ subdirectory
→ Add to Git if it's essential for understanding/running

### New Analysis/Planning Notes?
→ Place in `docs/claude-ai-guides/`
→ Automatically excluded from Git
→ Available for reference offline

---

## .gitignore Details

Key exclusions:
```
docs/claude-ai-guides/     # Claude AI analysis docs
*/node_modules/            # NPM dependencies
*/vendor/                  # Go dependencies
.env files                 # Sensitive configuration
*.log                      # Temporary logs
*/coverage/                # Test coverage reports
```

---

## Example: Adding a New Feature

### Developer Workflow
```bash
# Clone repo
git clone https://github.com/your-org/commandlinetodo.git

# Make changes
cd todo-server
# ... edit code ...

# Test locally
go test ./...

# Commit & push
git add .
git commit -m "feat: new feature"
git push origin feature/my-feature

# GitHub automatically:
# ✅ Runs all tests
# ✅ Checks code quality
# ✅ Builds all components
```

No need to:
- ❌ Read the phase reports
- ❌ Check the analysis docs
- ❌ Look at planning notes
→ Just write code!

---

## When to Look at claude-ai-guides/

**Use these for:**
- Understanding past decisions
- Learning the development process
- Referencing analysis when making architectural changes
- Understanding why certain decisions were made

**Don't use these for:**
- Basic setup (use CLAUDE.md)
- Running the app (use component READMEs)
- Testing (use TESTING.md)
- Contributing (use CONTRIBUTING.md)

---

## Summary

| Type | Location | In Git | Purpose |
|------|----------|--------|---------|
| Source Code | Components | ✅ | Application code |
| Tests | Each component | ✅ | Quality assurance |
| Core Docs | Root | ✅ | Essential reference |
| Analysis | `docs/claude-ai-guides/` | ❌ | Historical reference |
| Config | Each component | ✅ | Configuration |

This structure balances:
- **Completeness** - All code and essential docs tracked
- **Cleanliness** - Analysis/planning docs separate
- **Speed** - Faster clones, focused repos
- **Reference** - All documentation accessible locally

