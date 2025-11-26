import { describe, it, expect } from 'vitest';
import {
  parseIssueLinks,
  hasIssueLinks,
  extractIssueLinks,
  type ParsedIssueLink,
  type ParsedTextSegment,
} from '../issueLinkParser';

describe('issueLinkParser', () => {
  describe('parseIssueLinks', () => {
    it('should parse #123 format', () => {
      const result = parseIssueLinks('Check issue #123 for details');

      expect(result).toHaveLength(3);
      expect(result[0]).toEqual({
        type: 'text',
        text: 'Check issue ',
        start: 0,
        end: 12,
      });
      expect(result[1]).toEqual({
        type: 'issue-link',
        text: '#123',
        issueNumber: 123,
        start: 12,
        end: 16,
      });
      expect(result[2]).toEqual({
        type: 'text',
        text: ' for details',
        start: 16,
        end: 28,
      });
    });

    it('should parse TPP-123 format (project key + number)', () => {
      const result = parseIssueLinks('See TPP-456 for more info');

      expect(result).toHaveLength(3);
      expect(result[1]).toEqual({
        type: 'issue-link',
        text: 'TPP-456',
        projectKey: 'TPP',
        issueNumber: 456,
        start: 4,
        end: 11,
      });
    });

    it('should parse multiple issue links', () => {
      const result = parseIssueLinks('Related: #1, #2, and TPP-3');

      const links = result.filter((s): s is ParsedIssueLink => s.type === 'issue-link');
      expect(links).toHaveLength(3);
      expect(links[0].issueNumber).toBe(1);
      expect(links[1].issueNumber).toBe(2);
      expect(links[2].issueNumber).toBe(3);
      expect(links[2].projectKey).toBe('TPP');
    });

    it('should return single text segment for text without links', () => {
      const result = parseIssueLinks('No links here');

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        type: 'text',
        text: 'No links here',
        start: 0,
        end: 13,
      });
    });

    it('should handle text starting with link', () => {
      const result = parseIssueLinks('#1 is the first issue');

      expect(result).toHaveLength(2);
      expect(result[0].type).toBe('issue-link');
      expect((result[0] as ParsedIssueLink).issueNumber).toBe(1);
    });

    it('should handle text ending with link', () => {
      const result = parseIssueLinks('The last one is #999');

      expect(result).toHaveLength(2);
      expect(result[1].type).toBe('issue-link');
      expect((result[1] as ParsedIssueLink).issueNumber).toBe(999);
    });

    it('should handle only link', () => {
      const result = parseIssueLinks('#42');

      expect(result).toHaveLength(1);
      expect(result[0].type).toBe('issue-link');
      expect((result[0] as ParsedIssueLink).issueNumber).toBe(42);
    });

    it('should handle empty string', () => {
      const result = parseIssueLinks('');

      expect(result).toHaveLength(1);
      expect(result[0]).toEqual({
        type: 'text',
        text: '',
        start: 0,
        end: 0,
      });
    });

    it('should support project keys of various lengths (2-10 chars)', () => {
      const result1 = parseIssueLinks('AB-1');
      const result2 = parseIssueLinks('ABCDEFGHIJ-999');

      expect((result1[0] as ParsedIssueLink).projectKey).toBe('AB');
      expect((result2[0] as ParsedIssueLink).projectKey).toBe('ABCDEFGHIJ');
    });

    it('should not match lowercase project keys', () => {
      const result = parseIssueLinks('tpp-123 is lowercase');

      // Should not be parsed as a link
      const links = result.filter((s): s is ParsedIssueLink => s.type === 'issue-link');
      expect(links).toHaveLength(0);
    });

    it('should handle multiline text', () => {
      const result = parseIssueLinks('Line 1: #1\nLine 2: #2');

      const links = result.filter((s): s is ParsedIssueLink => s.type === 'issue-link');
      expect(links).toHaveLength(2);
    });
  });

  describe('hasIssueLinks', () => {
    it('should return true when text contains #N format', () => {
      expect(hasIssueLinks('See #123')).toBe(true);
    });

    it('should return true when text contains KEY-N format', () => {
      expect(hasIssueLinks('See TPP-456')).toBe(true);
    });

    it('should return false for plain text', () => {
      expect(hasIssueLinks('No links here')).toBe(false);
    });

    it('should return false for lowercase project key', () => {
      expect(hasIssueLinks('tpp-123')).toBe(false);
    });
  });

  describe('extractIssueLinks', () => {
    it('should extract only issue links', () => {
      const links = extractIssueLinks('Check #1 and TPP-2 for details');

      expect(links).toHaveLength(2);
      expect(links.every(l => l.type === 'issue-link')).toBe(true);
      expect(links[0].issueNumber).toBe(1);
      expect(links[1].issueNumber).toBe(2);
      expect(links[1].projectKey).toBe('TPP');
    });

    it('should return empty array for text without links', () => {
      const links = extractIssueLinks('No links');

      expect(links).toHaveLength(0);
    });
  });
});
