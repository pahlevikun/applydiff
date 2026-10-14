# ApplyDiff

Git-style approval for job applications.

Scan job boards, match each listing against your structured CV, let AI draft the tailoring edits, then review every edit as a diff and approve it hunk by hunk before a LaTeX PDF is built.

## Why

Most tools either auto-rewrite your CV or make you copy-paste. ApplyDiff treats your CV like a codebase: the AI proposes a change, you see exactly what changed and why, and you accept or reject each one.

## Principles

- **Local-first.** Runs on your machine. No account.
- **Human-in-the-loop.** It never submits an application. You click.
- **Structured CV is the source of truth.** LaTeX is an export target.
- **No fabricated numbers.** A metric that is not in your facts file cannot enter a diff.

## Status

Planning. See `docs/plans/2026-10-14-applydiff.md` and `docs/wireframes/`.

## Stack

Go + PostgreSQL + Nuxt 3 + LaTeX (latexmk).
