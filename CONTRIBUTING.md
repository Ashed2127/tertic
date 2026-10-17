# Contributing to Tertic

Thank you for your interest in contributing to Tertic! We welcome contributions from everyone.

## How to Contribute

### Reporting Bugs

- Open an issue on GitHub
- Describe the bug clearly
- Include steps to reproduce
- Mention your Android version and Termux version

### Suggesting Features

- Open an issue with the "enhancement" label
- Describe the feature and why it would help students
- Consider if it's mobile-friendly

### Adding Examples

We love new examples! To add one:

1. Create a new file in `examples/`
2. Make it lightweight (mobile-optimized)
3. Add comments explaining the code
4. Update the README if it's a major example

Example structure:
```python
#!/usr/bin/env python3
"""
Brief description of what this example does
"""

# Your code here
```

### Improving Documentation

- Documentation is in the `docs/` folder
- Keep language simple for students
- Include code examples
- Test instructions on real devices

## Development Guidelines

### Code Style

- Follow PEP 8 for Python
- Keep code simple and readable
- Add comments for complex logic
- Optimize for mobile (low memory, low CPU)

### Testing

- Test on real Android devices when possible
- Test on different Android versions (7, 10, 13+)
- Test with different RAM amounts (2GB, 4GB, 8GB)

### Performance

- Avoid heavy dependencies
- Use async I/O when possible
- Provide fallbacks for advanced features
- Document performance trade-offs

## Submitting Changes

1. Fork the repository
2. Create a branch for your changes
3. Make your changes
4. Test thoroughly
5. Submit a pull request

In your PR, describe:
- What you changed
- Why you changed it
- How you tested it
- Any known limitations

## Community Guidelines

- Be respectful and welcoming
- Assume good intentions
- Help others learn
- Focus on what's best for students

## Areas That Need Help

- More example agents
- Better error handling
- Local model support
- Android-specific optimizations
- Translations for documentation

---

Thank you for contributing! Every contribution helps students learn AI on their phones. 🎓
