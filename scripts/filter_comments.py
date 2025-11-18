#!/usr/bin/env python3
"""
Comment Filter - Filter and categorize review comments.

Phase 1: NO AI - Use regex, keywords, and heuristics
"""

import re
from typing import List, Set, Dict
from dataclasses import dataclass, asdict

from load_prs import ReviewComment


@dataclass
class FilteredComment:
    """Review comment with filtering metadata."""
    comment: ReviewComment
    is_actionable: bool
    categories: List[str]
    is_bot: bool
    keywords: List[str]
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            'comment': self.comment.to_dict(),
            'is_actionable': self.is_actionable,
            'categories': self.categories,
            'is_bot': self.is_bot,
            'keywords': self.keywords
        }


class CommentFilter:
    """Filter and categorize review comments."""
    
    # Noise patterns to filter out
    NOISE_PATTERNS = [
        r'^lgtm$',
        r'^looks good',
        r'^\+1$',
        r'^thanks?$',
        r'^thank you',
        r'^👍',
        r'^:thumbsup:',
        r'merge conflict',
        r'failing test',
        r'rebase',
        r'can you please',
        r'could you',
        r'^\s*$',  # Empty comments
    ]
    
    # Indicators of actionable feedback
    ACTIONABLE_INDICATORS = [
        'should', 'must', 'need to', 'needs to', 'required',
        'consider', 'suggest', 'recommend',
        'use instead', 'prefer', 'better to',
        'avoid', 'don\'t', 'do not',
        'missing', 'add', 'remove', 'change',
        'incorrect', 'wrong', 'issue', 'problem',
        'bug', 'error', 'warning',
    ]
    
    # Category keywords
    CATEGORY_KEYWORDS = {
        'security': [
            'security', 'vulnerability', 'exploit', 'sanitize', 'escape',
            'injection', 'xss', 'csrf', 'authentication', 'authorization',
            'password', 'secret', 'token', 'leak', 'exposure'
        ],
        'performance': [
            'performance', 'slow', 'optimize', 'cache', 'memory',
            'cpu', 'inefficient', 'bottleneck', 'scale', 'latency'
        ],
        'error_handling': [
            'error', 'exception', 'panic', 'nil check', 'null',
            'validation', 'handle', 'catch', 'try', 'defer'
        ],
        'style': [
            'style', 'idiomatic', 'convention', 'naming', 'format',
            'lint', 'clean', 'readable', 'consistent'
        ],
        'testing': [
            'test', 'coverage', 'mock', 'assertion', 'unit test',
            'integration test', 'e2e', 'testcase'
        ],
        'documentation': [
            'comment', 'document', 'doc', 'explain', 'godoc',
            'docstring', 'readme', 'description'
        ],
        'api_design': [
            'api', 'interface', 'contract', 'signature', 'parameter',
            'return', 'public', 'private', 'exported'
        ],
        'concurrency': [
            'concurrent', 'goroutine', 'mutex', 'lock', 'race',
            'deadlock', 'thread', 'async', 'parallel', 'context'
        ],
        'kubernetes': [
            'kubernetes', 'k8s', 'pod', 'deployment', 'service',
            'configmap', 'secret', 'namespace', 'crd', 'controller',
            'reconcile', 'operator', 'rbac'
        ],
        'best_practices': [
            'best practice', 'pattern', 'antipattern', 'refactor',
            'clean code', 'solid', 'dry', 'kiss'
        ],
    }
    
    # Bot identifiers
    BOT_IDENTIFIERS = [
        '[bot]',
        'bot',
        'automated',
        'ci-bot',
        'dependabot',
        'renovate',
        'coderabbit',
        'copilot',
    ]
    
    def __init__(self, exclude_bots: bool = False):
        self.exclude_bots = exclude_bots
    
    def is_bot_comment(self, comment: ReviewComment) -> bool:
        """Check if comment is from a bot."""
        username = comment.reviewer.login.lower()
        user_type = comment.reviewer.type.lower()
        
        if user_type == 'bot':
            return True
        
        for bot_id in self.BOT_IDENTIFIERS:
            if bot_id in username:
                return True
        
        return False
    
    def is_actionable(self, comment: ReviewComment) -> bool:
        """Determine if comment contains actionable feedback."""
        body_lower = comment.body.lower()
        
        # Filter out noise
        for pattern in self.NOISE_PATTERNS:
            if re.search(pattern, body_lower, re.IGNORECASE):
                return False
        
        # Look for actionable indicators
        for indicator in self.ACTIONABLE_INDICATORS:
            if indicator in body_lower:
                return True
        
        # Check for code suggestions (markdown code blocks)
        if '```' in comment.body:
            return True
        
        # Check for "instead of" pattern
        if 'instead of' in body_lower:
            return True
        
        # Check for questions that suggest changes
        if any(q in body_lower for q in ['why not', 'what about', 'have you considered']):
            return True
        
        # If comment is substantial (>50 chars) and mentions specific patterns
        if len(comment.body) > 50:
            # Look for specific code elements being discussed
            if any(pattern in body_lower for pattern in [
                'function', 'method', 'variable', 'struct', 'class',
                'field', 'parameter', 'return', 'type', 'interface'
            ]):
                return True
        
        return False
    
    def categorize(self, comment: ReviewComment) -> List[str]:
        """Categorize comment by type."""
        body_lower = comment.body.lower()
        categories = []
        
        for category, keywords in self.CATEGORY_KEYWORDS.items():
            for keyword in keywords:
                if keyword in body_lower:
                    if category not in categories:
                        categories.append(category)
                    break  # Move to next category once matched
        
        # Default category if no match
        if not categories:
            categories.append('general')
        
        return categories
    
    def extract_keywords(self, comment: ReviewComment) -> List[str]:
        """Extract key terms from comment."""
        body_lower = comment.body.lower()
        keywords = set()
        
        # Extract from actionable indicators
        for indicator in self.ACTIONABLE_INDICATORS:
            if indicator in body_lower:
                keywords.add(indicator)
        
        # Extract from category keywords
        for category, kw_list in self.CATEGORY_KEYWORDS.items():
            for kw in kw_list:
                if kw in body_lower:
                    keywords.add(kw)
        
        # Extract common technical terms (simple word extraction)
        # Look for capitalized words or words in code context
        words = re.findall(r'\b[A-Z][a-z]+\b|\b[a-z]+\b', comment.body)
        
        # Common important words
        important_words = {
            'context', 'error', 'nil', 'null', 'timeout', 'cancel',
            'mutex', 'lock', 'race', 'goroutine', 'channel',
            'validation', 'check', 'handle', 'return', 'defer',
            'test', 'mock', 'assert', 'expect',
            'refactor', 'simplify', 'extract', 'rename',
        }
        
        keywords.update(w.lower() for w in words if w.lower() in important_words)
        
        return sorted(list(keywords))[:10]  # Limit to top 10
    
    def filter_comment(self, comment: ReviewComment) -> FilteredComment:
        """Apply all filters to a comment."""
        is_bot = self.is_bot_comment(comment)
        is_actionable = self.is_actionable(comment)
        categories = self.categorize(comment)
        keywords = self.extract_keywords(comment)
        
        return FilteredComment(
            comment=comment,
            is_actionable=is_actionable,
            categories=categories,
            is_bot=is_bot,
            keywords=keywords
        )
    
    def filter_comments(self, comments: List[ReviewComment]) -> List[FilteredComment]:
        """Filter a list of comments."""
        filtered = []
        
        for comment in comments:
            fc = self.filter_comment(comment)
            
            # Skip bots if configured
            if self.exclude_bots and fc.is_bot:
                continue
            
            filtered.append(fc)
        
        return filtered
    
    def get_actionable_comments(self, filtered_comments: List[FilteredComment]) -> List[FilteredComment]:
        """Get only actionable comments."""
        return [fc for fc in filtered_comments if fc.is_actionable]


def analyze_comments(comments: List[ReviewComment], exclude_bots: bool = False):
    """Analyze and display comment statistics."""
    filter = CommentFilter(exclude_bots=exclude_bots)
    filtered = filter.filter_comments(comments)
    actionable = filter.get_actionable_comments(filtered)
    
    print(f"\n=== Comment Analysis ===")
    print(f"Total comments: {len(comments)}")
    print(f"After filtering: {len(filtered)}")
    print(f"Actionable comments: {len(actionable)}")
    print(f"Actionable rate: {len(actionable)/len(filtered)*100:.1f}%")
    
    # Bot vs human breakdown
    bot_comments = sum(1 for fc in filtered if fc.is_bot)
    human_comments = len(filtered) - bot_comments
    print(f"\nBot comments: {bot_comments}")
    print(f"Human comments: {human_comments}")
    
    # Actionable by source
    actionable_bots = sum(1 for fc in actionable if fc.is_bot)
    actionable_humans = len(actionable) - actionable_bots
    print(f"\nActionable bot comments: {actionable_bots}")
    print(f"Actionable human comments: {actionable_humans}")
    
    # Category breakdown
    category_counts = {}
    for fc in actionable:
        for cat in fc.categories:
            category_counts[cat] = category_counts.get(cat, 0) + 1
    
    print(f"\n=== Actionable Comments by Category ===")
    for cat, count in sorted(category_counts.items(), key=lambda x: -x[1]):
        print(f"{cat}: {count}")
    
    # Language breakdown for actionable comments
    lang_counts = {}
    for fc in actionable:
        lang = fc.comment.language or 'unknown'
        lang_counts[lang] = lang_counts.get(lang, 0) + 1
    
    print(f"\n=== Actionable Comments by Language ===")
    for lang, count in sorted(lang_counts.items(), key=lambda x: -x[1]):
        print(f"{lang}: {count}")
    
    # Sample actionable comments
    print(f"\n=== Sample Actionable Comments (first 5) ===")
    for i, fc in enumerate(actionable[:5], 1):
        comment = fc.comment
        print(f"\n{i}. PR #{comment.pr_number} - {comment.path}")
        print(f"   Reviewer: {comment.reviewer.login} ({'BOT' if fc.is_bot else 'HUMAN'})")
        print(f"   Categories: {', '.join(fc.categories)}")
        print(f"   Comment: {comment.body[:150]}...")
        if len(comment.body) > 150:
            print(f"            (...{len(comment.body) - 150} more chars)")
    
    return filtered, actionable


if __name__ == "__main__":
    # Load and analyze comments
    from load_prs import PRLoader
    import os
    
    # Default to .data in parent directory
    cache_dir = os.path.join(os.path.dirname(__file__), "..", ".data")
    loader = PRLoader(cache_dir)
    prs = loader.load_repository("opendatahub-io", "opendatahub-operator")
    comments = loader.extract_all_comments(prs)
    
    print("=" * 70)
    print("ANALYSIS WITH BOTS")
    print("=" * 70)
    filtered_with_bots, actionable_with_bots = analyze_comments(comments, exclude_bots=False)
    
    print("\n\n")
    print("=" * 70)
    print("ANALYSIS WITHOUT BOTS (HUMAN ONLY)")
    print("=" * 70)
    filtered_without_bots, actionable_without_bots = analyze_comments(comments, exclude_bots=True)

