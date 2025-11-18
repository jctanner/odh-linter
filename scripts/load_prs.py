#!/usr/bin/env python3
"""
PR Data Loader - Load PR data from go-github-scraper cache.

Phase 1: NO AI - Pure data loading and parsing
"""

import json
import os
from pathlib import Path
from typing import List, Dict, Optional
from dataclasses import dataclass, asdict
from datetime import datetime


@dataclass
class User:
    """GitHub user info."""
    login: str
    id: int
    type: str  # User, Bot
    avatar_url: Optional[str] = None


@dataclass
class ReviewComment:
    """Inline code review comment with code context."""
    # GitHub IDs
    comment_id: int
    pr_number: int
    repo: str
    
    # Comment details
    body: str
    reviewer: User
    created_at: datetime
    updated_at: datetime
    
    # Code context
    path: str
    line: Optional[int]
    original_line: Optional[int]
    diff_hunk: str
    commit_id: Optional[str]
    original_commit_id: Optional[str]
    
    # Additional context
    side: Optional[str] = None  # LEFT or RIGHT
    start_side: Optional[str] = None
    position: Optional[int] = None
    original_position: Optional[int] = None
    
    # Derived fields
    language: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        d = asdict(self)
        # Convert datetime to ISO string
        d['created_at'] = self.created_at.isoformat()
        d['updated_at'] = self.updated_at.isoformat()
        # Convert nested User to dict
        d['reviewer'] = asdict(self.reviewer)
        return d


@dataclass
class FileChange:
    """File change information."""
    filename: str
    status: str
    additions: int
    deletions: int
    changes: int
    patch: Optional[str] = None


@dataclass
class PullRequest:
    """Pull request with review comments."""
    # PR metadata
    repo: str
    number: int
    title: str
    state: str  # open, closed
    merged: bool
    
    # User info
    author: User
    
    # Timestamps
    created_at: datetime
    updated_at: datetime
    closed_at: Optional[datetime]
    merged_at: Optional[datetime]
    
    # Changes
    additions: int
    deletions: int
    changed_files: int
    files: List[FileChange]
    
    # Comments
    comments: List[ReviewComment]
    
    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        d = asdict(self)
        # Convert datetimes
        d['created_at'] = self.created_at.isoformat()
        d['updated_at'] = self.updated_at.isoformat()
        d['closed_at'] = self.closed_at.isoformat() if self.closed_at else None
        d['merged_at'] = self.merged_at.isoformat() if self.merged_at else None
        # Convert nested objects
        d['author'] = asdict(self.author)
        d['comments'] = [c.to_dict() for c in self.comments]
        d['files'] = [asdict(f) for f in self.files]
        return d


class PRLoader:
    """Load PR data from go-github-scraper cache."""
    
    def __init__(self, cache_dir: str):
        self.cache_dir = Path(cache_dir)
    
    def detect_language(self, filepath: str) -> Optional[str]:
        """Detect programming language from file extension."""
        ext_to_lang = {
            '.py': 'python',
            '.go': 'go',
            '.js': 'javascript',
            '.ts': 'typescript',
            '.tsx': 'typescript',
            '.jsx': 'javascript',
            '.java': 'java',
            '.rb': 'ruby',
            '.rs': 'rust',
            '.c': 'c',
            '.cpp': 'cpp',
            '.cc': 'cpp',
            '.h': 'c',
            '.hpp': 'cpp',
            '.cs': 'csharp',
            '.sh': 'shell',
            '.yaml': 'yaml',
            '.yml': 'yaml',
            '.json': 'json',
            '.xml': 'xml',
            '.md': 'markdown',
        }
        
        ext = os.path.splitext(filepath)[1].lower()
        return ext_to_lang.get(ext)
    
    def parse_datetime(self, dt_str: str) -> Optional[datetime]:
        """Parse GitHub datetime string."""
        if not dt_str:
            return None
        try:
            return datetime.fromisoformat(dt_str.replace('Z', '+00:00'))
        except:
            return None
    
    def load_pr(self, pr_file: Path) -> Optional[PullRequest]:
        """Load a single PR from JSON file."""
        try:
            with open(pr_file) as f:
                data = json.load(f)
            
            github_data = data.get('github_data', {})
            
            # Extract repo from file path
            # Path format: .data/api.github.com/repos/owner/repo/pulls/123.json
            parts = pr_file.parts
            owner_idx = parts.index('repos') + 1
            repo = f"{parts[owner_idx]}/{parts[owner_idx + 1]}"
            
            # Parse author
            author_data = github_data.get('user', {})
            author = User(
                login=author_data.get('login', 'unknown'),
                id=author_data.get('id', 0),
                type=author_data.get('type', 'User'),
                avatar_url=author_data.get('avatar_url')
            )
            
            # Parse file changes
            files = []
            for f in github_data.get('files', []):
                files.append(FileChange(
                    filename=f.get('filename', ''),
                    status=f.get('status', ''),
                    additions=f.get('additions', 0),
                    deletions=f.get('deletions', 0),
                    changes=f.get('changes', 0),
                    patch=f.get('patch')
                ))
            
            # Parse inline review comments
            comments = []
            for c in github_data.get('comments', []):
                # Only process inline comments with diff_hunk
                if not c.get('diff_hunk'):
                    continue
                
                reviewer_data = c.get('user', {})
                reviewer = User(
                    login=reviewer_data.get('login', 'unknown'),
                    id=reviewer_data.get('id', 0),
                    type=reviewer_data.get('type', 'User'),
                    avatar_url=reviewer_data.get('avatar_url')
                )
                
                path = c.get('path', '')
                comment = ReviewComment(
                    comment_id=c.get('id', 0),
                    pr_number=github_data.get('number', 0),
                    repo=repo,
                    body=c.get('body', ''),
                    reviewer=reviewer,
                    created_at=self.parse_datetime(c.get('created_at')) or datetime.now(),
                    updated_at=self.parse_datetime(c.get('updated_at')) or datetime.now(),
                    path=path,
                    line=c.get('line'),
                    original_line=c.get('original_line'),
                    diff_hunk=c.get('diff_hunk', ''),
                    commit_id=c.get('commit_id'),
                    original_commit_id=c.get('original_commit_id'),
                    side=c.get('side'),
                    start_side=c.get('start_side'),
                    position=c.get('position'),
                    original_position=c.get('original_position'),
                    language=self.detect_language(path)
                )
                comments.append(comment)
            
            pr = PullRequest(
                repo=repo,
                number=github_data.get('number', 0),
                title=github_data.get('title', ''),
                state=github_data.get('state', ''),
                merged=github_data.get('merged', False),
                author=author,
                created_at=self.parse_datetime(github_data.get('created_at')) or datetime.now(),
                updated_at=self.parse_datetime(github_data.get('updated_at')) or datetime.now(),
                closed_at=self.parse_datetime(github_data.get('closed_at')),
                merged_at=self.parse_datetime(github_data.get('merged_at')),
                additions=github_data.get('additions', 0),
                deletions=github_data.get('deletions', 0),
                changed_files=github_data.get('changed_files', 0),
                files=files,
                comments=comments
            )
            
            return pr
            
        except Exception as e:
            print(f"Error loading PR from {pr_file}: {e}")
            return None
    
    def load_repository(self, owner: str, repo: str) -> List[PullRequest]:
        """Load all PRs for a repository."""
        repo_path = self.cache_dir / "api.github.com" / "repos" / owner / repo / "pulls"
        
        if not repo_path.exists():
            print(f"Repository path not found: {repo_path}")
            return []
        
        prs = []
        pr_files = sorted(repo_path.glob("*.json"))
        
        # Filter out supplementary files (_reviews, _comments, etc.)
        pr_files = [f for f in pr_files if '_' not in f.stem]
        
        print(f"Loading {len(pr_files)} PRs from {owner}/{repo}...")
        
        for pr_file in pr_files:
            pr = self.load_pr(pr_file)
            if pr:
                prs.append(pr)
        
        print(f"Loaded {len(prs)} PRs with {sum(len(pr.comments) for pr in prs)} inline comments")
        
        return prs
    
    def extract_all_comments(self, prs: List[PullRequest]) -> List[ReviewComment]:
        """Extract all inline review comments from PRs."""
        comments = []
        for pr in prs:
            comments.extend(pr.comments)
        return comments


if __name__ == "__main__":
    # Test with opendatahub-operator data
    loader = PRLoader("/home/jtanner/workspace/github/jctanner.redhat/2025_11_10_rhoai_merge_stats/.data")
    prs = loader.load_repository("opendatahub-io", "opendatahub-operator")
    
    print(f"\n=== Summary ===")
    print(f"Total PRs: {len(prs)}")
    print(f"PRs with comments: {sum(1 for pr in prs if pr.comments)}")
    
    comments = loader.extract_all_comments(prs)
    print(f"Total inline comments: {len(comments)}")
    
    # Language breakdown
    lang_counts = {}
    for c in comments:
        lang = c.language or 'unknown'
        lang_counts[lang] = lang_counts.get(lang, 0) + 1
    
    print(f"\n=== Comments by Language ===")
    for lang, count in sorted(lang_counts.items(), key=lambda x: -x[1]):
        print(f"{lang}: {count}")
    
    # Reviewer breakdown
    reviewer_counts = {}
    for c in comments:
        reviewer = c.reviewer.login
        reviewer_counts[reviewer] = reviewer_counts.get(reviewer, 0) + 1
    
    print(f"\n=== Top 10 Reviewers ===")
    for reviewer, count in sorted(reviewer_counts.items(), key=lambda x: -x[1])[:10]:
        print(f"{reviewer}: {count}")

