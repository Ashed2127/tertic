# Tertic Project Structure

```
tertic/
├── README.md                      # Main project documentation
├── LICENSE                        # MIT License
├── CONTRIBUTING.md                # Contribution guidelines
├── .gitignore                     # Git ignore rules
├── .env.example                   # Environment variables template
│
├── setup.sh                       # Main installation script (executable)
├── quick-start.sh                 # Quick start helper (executable)
├── advanced-setup.sh              # Advanced features setup (executable)
│
├── termux.properties              # Optimized Termux configuration
├── optimized-requirements.txt     # Lightweight Python dependencies
│
├── examples/                      # Example AI agents
│   ├── simple_agent.py           # Basic agent (with/without API)
│   └── langchain_agent.py        # LangChain-based agent
│
└── docs/                          # Documentation
    ├── STUDENT_GUIDE.md          # Step-by-step student guide
    ├── PERFORMANCE_TIPS.md       # Performance optimization guide
    └── QUICK_REFERENCE.md       # Quick reference card
```

## File Descriptions

### Core Files

- **README.md**: Project overview, features, and quick start instructions
- **setup.sh**: Automated installation script that sets up Termux for AI agents
- **termux.properties**: Optimized Termux settings for mobile AI agent usage
- **optimized-requirements.txt**: Curated list of lightweight Python packages

### Documentation

- **STUDENT_GUIDE.md**: Comprehensive guide for students to get started
- **PERFORMANCE_TIPS.md**: Advanced optimization techniques for better performance
- **QUICK_REFERENCE.md**: Quick reference card with common commands
- **CONTRIBUTING.md**: Guidelines for contributors

### Examples

- **simple_agent.py**: Basic AI agent with two modes:
  - Local mode (no API key needed)
  - OpenAI mode (requires API key)
- **langchain_agent.py**: More advanced agent using LangChain framework

### Setup Scripts

- **setup.sh**: Main installation - updates Termux, installs Python, UV, and dependencies
- **quick-start.sh**: Quick helper to activate environment and show available commands
- **advanced-setup.sh**: Optional - installs LangChain, additional tools (tmux, ripgrep, fzf)

## Installation Flow

1. User downloads Termux from GitHub
2. User clones tertic repository
3. User runs `bash setup.sh`
4. Setup script:
   - Updates Termux packages
   - Installs Python and tools
   - Installs UV package manager
   - Creates Python virtual environment
   - Installs optimized dependencies
   - Applies mobile optimizations
5. User activates environment with `source ~/tertic-env/bin/activate`
6. User runs example agents

## Key Features

### Mobile Optimizations

- Lightweight package selection
- UV package manager (faster than pip)
- Optimized environment variables
- Termux configuration for better mobile experience
- Memory and CPU tuning

### Student-Friendly

- Step-by-step documentation
- No local agent example (works without API key)
- Quick reference card
- Clear error messages
- Performance tips

### Extensible

- Easy to add new examples
- Modular setup scripts
- Clear contribution guidelines
- Multiple AI framework support

## Dependencies

### Core (installed by setup.sh)
- Python 3.x
- UV (package manager)
- aiohttp (async HTTP)
- aiosqlite (async SQLite)
- pyyaml (YAML parsing)
- python-dotenv (environment variables)
- rich (terminal formatting)
- openai (OpenAI API)
- anthropic (Anthropic API)

### Optional (advanced-setup.sh)
- langchain (AI framework)
- langchain-community (LangChain integrations)
- tmux (terminal multiplexer)
- ripgrep (fast search)
- fzf (fuzzy finder)

## System Requirements

- Android 7.0 or higher
- 2GB RAM minimum (4GB+ recommended)
- 1GB free storage
- Internet connection (for API calls and package downloads)

## Target Users

- Students without laptops
- Learners interested in AI
- Developers on the go
- Anyone wanting to run AI agents on mobile

## Future Enhancements

Potential additions:
- More example agents
- Local model support (Ollama, etc.)
- Web UI option
- More framework integrations
- Performance monitoring tools
- Battery usage statistics
