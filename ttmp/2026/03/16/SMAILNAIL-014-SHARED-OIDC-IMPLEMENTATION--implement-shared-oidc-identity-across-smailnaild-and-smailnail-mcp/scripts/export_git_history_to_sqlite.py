#!/usr/bin/env python3
"""
Export git commit and file history to a SQLite database for ad hoc timeline queries.

Example:
  ./export_git_history_to_sqlite.py \
    /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail \
    /tmp/smailnail-history.sqlite

Example queries:
  sqlite3 /tmp/smailnail-history.sqlite \
    "select commit_hash, author_time, subject from commits order by author_time asc limit 10;"

  sqlite3 /tmp/smailnail-history.sqlite \
    "select c.commit_hash, c.author_time, c.subject
       from commit_files f
       join commits c on c.commit_hash = f.commit_hash
      where f.path like 'pkg/mcp/%' or f.path like 'cmd/smailnail-imap-mcp/%'
      order by c.author_time asc;"
"""

from __future__ import annotations

import argparse
import os
import sqlite3
import subprocess
import sys
from typing import Iterable


def run_git(repo: str, *args: str) -> str:
    proc = subprocess.run(
        ["git", "-C", repo, *args],
        check=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return proc.stdout


def iter_commits(repo: str) -> Iterable[str]:
    output = run_git(repo, "rev-list", "--reverse", "--all")
    for line in output.splitlines():
        commit = line.strip()
        if commit:
            yield commit


def init_db(db_path: str) -> sqlite3.Connection:
    if os.path.exists(db_path):
        os.remove(db_path)
    conn = sqlite3.connect(db_path)
    conn.executescript(
        """
        PRAGMA journal_mode = WAL;
        PRAGMA synchronous = NORMAL;

        CREATE TABLE commits (
            commit_hash TEXT PRIMARY KEY,
            parents TEXT NOT NULL,
            author_name TEXT NOT NULL,
            author_email TEXT NOT NULL,
            author_time TEXT NOT NULL,
            subject TEXT NOT NULL,
            body TEXT NOT NULL
        );

        CREATE TABLE commit_files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            commit_hash TEXT NOT NULL,
            status TEXT NOT NULL,
            path TEXT NOT NULL,
            old_path TEXT,
            added_lines INTEGER,
            deleted_lines INTEGER,
            FOREIGN KEY (commit_hash) REFERENCES commits(commit_hash)
        );

        CREATE INDEX idx_commit_files_path ON commit_files(path);
        CREATE INDEX idx_commit_files_commit_hash ON commit_files(commit_hash);
        CREATE UNIQUE INDEX idx_commit_files_unique
            ON commit_files(commit_hash, status, path, old_path);
        CREATE INDEX idx_commits_author_time ON commits(author_time);
        """
    )
    return conn


def export_commit(conn: sqlite3.Connection, repo: str, commit: str) -> None:
    meta = run_git(
        repo,
        "show",
        "--quiet",
        "--format=%H%x1f%P%x1f%an%x1f%ae%x1f%aI%x1f%s%x1f%b",
        commit,
    ).rstrip("\n")
    commit_hash, parents, author_name, author_email, author_time, subject, body = meta.split("\x1f", 6)
    conn.execute(
        """
        INSERT INTO commits (
            commit_hash, parents, author_name, author_email, author_time, subject, body
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
        """,
        (commit_hash, parents, author_name, author_email, author_time, subject, body),
    )

    name_status_lines = run_git(repo, "diff-tree", "--root", "--no-commit-id", "--name-status", "-r", commit).splitlines()
    numstat_lines = run_git(repo, "diff-tree", "--root", "--no-commit-id", "--numstat", "-r", commit).splitlines()

    stats_by_path: dict[str, tuple[int | None, int | None]] = {}
    for line in numstat_lines:
        parts = line.split("\t")
        if len(parts) < 3:
            continue
        added_raw, deleted_raw, path = parts[0], parts[1], parts[2]
        added = None if added_raw == "-" else int(added_raw)
        deleted = None if deleted_raw == "-" else int(deleted_raw)
        stats_by_path[path] = (added, deleted)

    for line in name_status_lines:
        parts = line.split("\t")
        if not parts:
            continue
        status = parts[0]
        old_path = None
        path = ""
        if status.startswith("R") or status.startswith("C"):
            if len(parts) >= 3:
                old_path = parts[1]
                path = parts[2]
        elif len(parts) >= 2:
            path = parts[1]
        else:
            continue

        added, deleted = stats_by_path.get(path, (None, None))
        conn.execute(
            """
            INSERT INTO commit_files (
                commit_hash, status, path, old_path, added_lines, deleted_lines
            ) VALUES (?, ?, ?, ?, ?, ?)
            """,
            (commit_hash, status, path, old_path, added, deleted),
        )


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("repo", help="Path to the git repository")
    parser.add_argument("db", help="Path to the SQLite database to create")
    args = parser.parse_args()

    repo = os.path.abspath(args.repo)
    db_path = os.path.abspath(args.db)

    conn = init_db(db_path)
    try:
        with conn:
            for commit in iter_commits(repo):
                export_commit(conn, repo, commit)
    finally:
        conn.close()

    print(db_path)
    return 0


if __name__ == "__main__":
    sys.exit(main())
