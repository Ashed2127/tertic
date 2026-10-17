#!/usr/bin/env python3
"""
LangChain Agent Example for Tertic
A more advanced agent using LangChain framework
"""

import os
from rich.console import Console
from rich.panel import Panel
from rich.prompt import Prompt

console = Console()

def langchain_agent():
    """LangChain-based agent"""
    try:
        from langchain_openai import ChatOpenAI
        from langchain_core.messages import HumanMessage
    except ImportError:
        console.print("[red]LangChain not installed. Run: pip install langchain langchain-openai[/red]")
        return

    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        console.print("[yellow]OpenAI API key required![/yellow]")
        api_key = Prompt.ask("Enter your OpenAI API key", password=True)
        if not api_key:
            return

    # Initialize with optimized settings for mobile
    llm = ChatOpenAI(
        model="gpt-3.5-turbo",
        api_key=api_key,
        temperature=0.7,
        max_tokens=500,  # Limited for mobile
        timeout=30  # Don't hang
    )

    console.print(Panel.fit(
        "[bold green]Tertic LangChain Agent[/bold green]\n"
        "Powered by LangChain framework",
        title="Welcome"
    ))

    while True:
        user_input = Prompt.ask("\n[bold cyan]You[/bold cyan]")

        if user_input.lower() in ['quit', 'exit', 'q']:
            console.print("[yellow]Goodbye! 👋[/yellow]")
            break

        if not user_input.strip():
            continue

        console.print("[dim]Processing with LangChain...[/dim]")

        try:
            response = llm.invoke([HumanMessage(content=user_input)])
            console.print(Panel.fit(
                response.content,
                title="[bold magenta]LangChain Agent[/bold magenta]",
                border_style="magenta"
            ))
        except Exception as e:
            console.print(f"[red]Error: {e}[/red]")

if __name__ == "__main__":
    langchain_agent()
