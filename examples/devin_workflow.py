#!/usr/bin/env python3
"""
Example: Setting up a project for Devin CLI on Android
This script creates a simple project structure to test Devin CLI
"""

import os
from rich.console import Console
from rich.panel import Panel

console = Console()

def create_sample_project():
    """Create a sample project for Devin to work with"""

    project_name = "devin-test-project"
    project_path = os.path.expanduser(f"~/{project_name}")

    if os.path.exists(project_path):
        console.print(f"[yellow]Project {project_name} already exists![/yellow]")
        return project_path

    os.makedirs(project_path, exist_ok=True)

    # Create a simple Python file
    with open(os.path.join(project_path, "main.py"), "w") as f:
        f.write("""# Simple Python script
def greet(name):
    return f"Hello, {name}!"

if __name__ == "__main__":
    print(greet("World"))
""")

    # Create a README
    with open(os.path.join(project_path, "README.md"), "w") as f:
        f.write(f"""# {project_name}

A simple test project for Devin CLI on Android.

## Instructions

1. Make sure Devin CLI is installed: `devin --version`
2. Login to Devin: `devin auth login`
3. Run Devin in this directory: `devin`
4. Ask Devin to: "Add a function to calculate the sum of two numbers"
""")

    # Create a requirements.txt
    with open(os.path.join(project_path, "requirements.txt"), "w") as f:
        f.write("# No dependencies needed for this simple example\n")

    console.print(Panel.fit(
        f"[bold green]Project created at: {project_path}[/bold green]\n\n"
        f"To use Devin CLI:\n"
        f"1. cd {project_name}\n"
        f"2. devin auth login (if not logged in)\n"
        f"3. devin\n"
        f"4. Ask Devin to help with the code!",
        title="Project Ready"
    ))

    return project_path

if __name__ == "__main__":
    console.print("\n[bold cyan]Creating sample project for Devin CLI...[/bold cyan]\n")
    project_path = create_sample_project()
    console.print(f"\n[yellow]Project path: {project_path}[/yellow]")
