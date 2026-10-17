# Tertic Performance Optimization Guide

This guide helps you get the best performance from AI agents on your Android phone.

## System-Level Optimizations

### 1. Enable Developer Options

```
Settings → About Phone → Tap "Build Number" 7 times
```

Then enable:
- **High performance mode** (if available)
- **Disable HW overlays**
- **Force GPU rendering**

### 2. Battery Optimization

Exclude Termux from battery optimization:
```
Settings → Apps → Termux → Battery → Don't optimize
```

### 3. Termux Configuration

Use the provided `termux.properties`:
```bash
mkdir -p ~/.termux
cp ~/tertic/termux.properties ~/.termux/
```

Key settings:
- `wake-lock = true` - Keeps phone awake
- `extra-keys` - Better keyboard shortcuts
- `font-size = 10` - Smaller font = more content visible

## Python/Agent Optimizations

### 1. Use UV Package Manager

UV is much faster than pip:
```bash
# Install packages with UV
uv pip install package-name
```

### 2. Choose Lightweight Models

For OpenAI:
- Use `gpt-3.5-turbo` instead of `gpt-4` (faster, cheaper)
- Set lower `max_tokens` (e.g., 500 instead of 2000)

Example:
```python
response = client.chat.completions.create(
    model="gpt-3.5-turbo",  # Faster than gpt-4
    max_tokens=500,         # Limit response length
    messages=[...]
)
```

### 3. Environment Variables

Add to `~/.tertic-config`:
```bash
export PYTHONUNBUFFERED=1          # Faster output
export PYTHONDONTWRITEBYTECODE=1   # Skip .pyc files
export OMP_NUM_THREADS=4           # Limit CPU threads
export MALLOC_TRIM_THRESHOLD_=65536  # Reduce memory
```

### 4. Use Async Code

Async is more efficient on mobile:
```python
import asyncio
import aiohttp

async def fetch_data():
    async with aiohttp.ClientSession() as session:
        async with session.get('https://api.example.com') as resp:
            return await resp.text()

asyncio.run(fetch_data())
```

## Memory Management

### 1. Monitor Memory Usage

```bash
# Install htop
pkg install htop

# Run it
htop
```

### 2. Clear Caches

```bash
# Clear pip cache
pip cache purge

# Clear Termux package cache
pkg clean

# Clear UV cache
rm -rf ~/.local/share/uv/cache
```

### 3. Use Streaming Responses

For large outputs, stream instead of waiting:
```python
# Stream OpenAI responses
stream = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[...],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Network Optimization

### 1. Use Wi-Fi When Possible

Mobile data can be slower and use more battery.

### 2. Set Short Timeouts

```python
import aiohttp

timeout = aiohttp.ClientTimeout(total=30)  # 30 second timeout
async with aiohttp.ClientSession(timeout=timeout) as session:
    ...
```

### 3. Batch Requests

Combine multiple requests into one when possible.

## Specific Framework Tips

### OpenAI

```python
# Use smaller models
model="gpt-3.5-turbo"

# Limit context
max_tokens=500

# Use temperature 0 for deterministic outputs
temperature=0
```

### LangChain

```python
# Use lightweight chains
from langchain.chains import LLMChain

# Avoid heavy tools
# Don't use: SerpAPI, Wolfram Alpha (unless needed)
```

### Anthropic

```python
# Use Claude Haiku (fastest)
model="claude-3-haiku-20240307"

# Limit tokens
max_tokens=1024
```

## Troubleshooting Performance Issues

### Agent is slow

1. Check if other apps are running
2. Switch to Wi-Fi
3. Use a lighter model
4. Reduce `max_tokens`
5. Clear caches

### Out of memory errors

1. Close other apps
2. Reduce batch size
3. Use streaming responses
4. Restart Termux

### Phone gets hot

1. Take breaks between long tasks
2. Use a cooling pad or fan
3. Reduce model complexity
4. Keep phone charging

## Advanced: Use Local Models

For better privacy and speed, try local models:

### Ollama on Termux (Experimental)

```bash
# Install Ollama (may require workarounds on Android)
# Check: https://github.com/ollama/ollama

# Run tiny models
ollama run phi
```

### LiteLLM with Local Models

```bash
uv pip install litellm
```

```python
from litellm import completion

response = completion(
    model="ollama/phi",  # Lightweight local model
    messages=[{"role": "user", "content": "Hello"}]
)
```

## Benchmark Your Setup

Test your setup speed:

```python
import time
from openai import OpenAI

client = OpenAI()

start = time.time()
response = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "Say 'test'"}],
    max_tokens=10
)
end = time.time()

print(f"Response time: {end - start:.2f} seconds")
```

Aim for under 5 seconds for simple queries.

---

**Remember**: Optimization is about trade-offs. Find what works best for your phone and use case!
