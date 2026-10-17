#!/usr/bin/env python3
"""
Simple AI Agent Example for Tertic
A lightweight AI agent that runs on Android Termux
"""

import os
import sys
from rich.console import Console
from rich.panel import Panel
from rich.prompt import Prompt

console = Console()

def simple_openai_agent():
    """A simple OpenAI-based agent example"""
    try:
        from openai import OpenAI
    except ImportError:
        console.print("[red]OpenAI library not installed. Run: uv pip install openai[/red]")
        return

    # Check for API key
    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        console.print("[yellow]OpenAI API key not found![/yellow]")
        console.print("Set it with: export OPENAI_API_KEY='your-key-here'")
        api_key = Prompt.ask("Or enter it now", password=True)
        if not api_key:
            console.print("[red]No API key provided. Exiting.[/red]")
            return

    client = OpenAI(api_key=api_key)

    console.print(Panel.fit(
        "[bold green]Tertic Simple Agent[/bold green]\n"
        "Type your message, or 'quit' to exit",
        title="Welcome"
    ))

    while True:
        user_input = Prompt.ask("\n[bold cyan]You[/bold cyan]")

        if user_input.lower() in ['quit', 'exit', 'q']:
            console.print("[yellow]Goodbye! 👋[/yellow]")
            break

        if not user_input.strip():
            continue

        console.print("[dim]Thinking...[/dim]")

        try:
            response = client.chat.completions.create(
                model="gpt-3.5-turbo",
                messages=[
                    {"role": "system", "content": "You are a helpful AI assistant running on an Android phone via Tertic. Be concise and helpful."},
                    {"role": "user", "content": user_input}
                ],
                max_tokens=500
            )

            assistant_message = response.choices[0].message.content
            console.print(Panel.fit(
                assistant_message,
                title="[bold magenta]Assistant[/bold magenta]",
                border_style="magenta"
            ))

        except Exception as e:
            console.print(f"[red]Error: {e}[/red]")

def simple_local_agent():
    """A simple rule-based agent (no API key needed)"""
    console.print(Panel.fit(
        "[bold green]Tertic Local Agent[/bold green]\n"
        "A simple rule-based assistant (no API key needed)",
        title="Welcome"
    ))

    responses = {
        "hello": "Hello! I'm Tertic, your AI assistant on Android!",
        "help": "I can help with basic questions. Try asking about Tertic, Python, or general topics!",
        "tertic": "Tertic is an optimized Termux setup for running AI agents on Android phones!",
        "python": "Python is great! You can run it right here in Termux. Try: python your_script.py",
        "battery": "To save battery: close other apps, lower screen brightness, and use lighter AI models.",
    }

    while True:
        user_input = Prompt.ask("\n[bold cyan]You[/bold cyan]").lower()

        if user_input in ['quit', 'exit', 'q']:
            console.print("[yellow]Goodbye! 👋[/yellow]")
            break

        # Find matching response
        response = None
        for key, value in responses.items():
            if key in user_input:
                response = value
                break

        if response:
            console.print(Panel.fit(
                response,
                title="[bold magenta]Assistant[/bold magenta]",
                border_style="magenta"
            ))
        else:
            console.print("[dim]I'm a simple local agent. For smarter responses, set up an API key and use the OpenAI agent![/dim]")

if __name__ == "__main__":
    console.print("\n[bold cyan]Tertic Simple Agent[/bold cyan]")
    console.print("Choose agent type:")
    console.print("1. OpenAI Agent (requires API key)")
    console.print("2. Local Agent (no API needed)")

    choice = Prompt.ask("Select", choices=["1", "2"], default="2")

    if choice == "1":
        simple_openai_agent()
    else:
        simple_local_agent()
