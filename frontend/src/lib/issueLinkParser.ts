/**
 * Issue Link Parser
 *
 * Parses text content and identifies issue references in the following formats:
 * - #123 (issue number only - requires project context)
 * - TPP-123 (project key + issue number)
 */

export interface ParsedIssueLink {
  type: 'issue-link';
  text: string;          // Original matched text (e.g., "#123" or "TPP-123")
  projectKey?: string;   // Project key if present (e.g., "TPP")
  issueNumber: number;   // Issue number (e.g., 123)
  start: number;         // Start index in original text
  end: number;           // End index in original text
}

export interface ParsedTextSegment {
  type: 'text';
  text: string;
  start: number;
  end: number;
}

export type ParsedSegment = ParsedIssueLink | ParsedTextSegment;

// Regex patterns for issue references
// Matches: #123, TPP-123, TPP-1, etc.
const ISSUE_LINK_PATTERN = /(?:([A-Z]{2,10})-)?#?(\d+)/g;

// More strict pattern that requires either project key or # prefix
const STRICT_ISSUE_LINK_PATTERN = /(?:([A-Z]{2,10})-(\d+))|(?:#(\d+))/g;

/**
 * Parse text and extract issue links
 * @param text The text content to parse
 * @returns Array of parsed segments (text and issue links)
 */
export function parseIssueLinks(text: string): ParsedSegment[] {
  const segments: ParsedSegment[] = [];
  let lastIndex = 0;

  // Reset regex state
  STRICT_ISSUE_LINK_PATTERN.lastIndex = 0;

  let match;
  while ((match = STRICT_ISSUE_LINK_PATTERN.exec(text)) !== null) {
    // Add text before this match
    if (match.index > lastIndex) {
      segments.push({
        type: 'text',
        text: text.slice(lastIndex, match.index),
        start: lastIndex,
        end: match.index,
      });
    }

    // Determine if it's a project-key format or # format
    if (match[1] && match[2]) {
      // Format: TPP-123
      segments.push({
        type: 'issue-link',
        text: match[0],
        projectKey: match[1],
        issueNumber: parseInt(match[2], 10),
        start: match.index,
        end: match.index + match[0].length,
      });
    } else if (match[3]) {
      // Format: #123
      segments.push({
        type: 'issue-link',
        text: match[0],
        issueNumber: parseInt(match[3], 10),
        start: match.index,
        end: match.index + match[0].length,
      });
    }

    lastIndex = match.index + match[0].length;
  }

  // Add remaining text after last match
  if (lastIndex < text.length) {
    segments.push({
      type: 'text',
      text: text.slice(lastIndex),
      start: lastIndex,
      end: text.length,
    });
  }

  // If no segments, return the whole text as a single segment
  if (segments.length === 0) {
    segments.push({
      type: 'text',
      text: text,
      start: 0,
      end: text.length,
    });
  }

  return segments;
}

/**
 * Check if text contains any issue links
 */
export function hasIssueLinks(text: string): boolean {
  STRICT_ISSUE_LINK_PATTERN.lastIndex = 0;
  return STRICT_ISSUE_LINK_PATTERN.test(text);
}

/**
 * Extract all issue links from text
 */
export function extractIssueLinks(text: string): ParsedIssueLink[] {
  return parseIssueLinks(text).filter(
    (segment): segment is ParsedIssueLink => segment.type === 'issue-link'
  );
}

/**
 * Generate URL for an issue link
 */
export function generateIssueLinkUrl(
  link: ParsedIssueLink,
  currentProjectId: number,
  projectKeyToIdMap?: Map<string, number>
): string | null {
  let projectId = currentProjectId;

  if (link.projectKey && projectKeyToIdMap) {
    const mappedId = projectKeyToIdMap.get(link.projectKey);
    if (mappedId) {
      projectId = mappedId;
    } else {
      // Unknown project key
      return null;
    }
  }

  // We need to find the issue ID from the issue number
  // For now, return a placeholder that will be resolved by the component
  return `/projects/${projectId}/issues/by-number/${link.issueNumber}`;
}
