export interface DiffChange {
  type: 'add' | 'del' | 'normal';
  add?: boolean;
  del?: boolean;
  normal?: boolean;
  ln?: number;
  ln1?: number;
  ln2?: number;
  content: string;
}

export interface DiffChunk {
  content: string;
  changes: DiffChange[];
  oldStart?: number;
  oldLines?: number;
  newStart?: number;
  newLines?: number;
}

export interface DiffFile {
  from?: string;
  to?: string;
  chunks: DiffChunk[];
  deletions: number;
  additions: number;
  new?: boolean;
  deleted?: boolean;
}

/**
 * Robust detection of whether a given text contains diff content,
 * including truncated diffs, unified diffs, git patches, or isolated hunks.
 */
export function isDiffText(text: string, language?: string): boolean {
  if (!text || typeof text !== 'string') return false;
  if (language === 'diff') return true;

  const trimmed = text.trim();
  if (!trimmed) return false;

  // 1. Explicit git diff markers
  if (trimmed.includes('diff --git ')) return true;
  if (trimmed.includes('--- a/') && trimmed.includes('+++ b/')) return true;

  // 2. Standard unified diff hunk header accompanied by change lines
  const hasHunkHeader = /^@@\s+-\d+(?:,\d+)?\s+\+\d+(?:,\d+)?\s+@@/m.test(trimmed);
  const hasChanges = /^[+-][^+-]/m.test(trimmed);
  if (hasHunkHeader && hasChanges) return true;

  // 3. Truncated diff output (e.g. <truncated 305 lines> followed by diff changes)
  if (/<truncated\s+\d+\s+(?:lines|bytes)>/i.test(trimmed)) {
    if (hasHunkHeader || hasChanges) {
      const sampleLines = trimmed
        .split(/\r?\n/)
        .filter((l) => l.trim().length > 0 && !/^<truncated/i.test(l.trim()))
        .slice(0, 30);
      const diffLineCount = sampleLines.filter((l) => /^[+\- @]/.test(l)).length;
      if (sampleLines.length >= 3 && diffLineCount / sampleLines.length >= 0.6) {
        return true;
      }
    }
  }

  return false;
}

/**
 * Extracts the file path from git diff headers or hunk symbol references.
 */
export function extractDiffFilePath(text: string): string | undefined {
  if (!text) return undefined;

  const gitMatch = /diff --git a\/[^\s\n]+\s+b\/([^\s\n]+)/.exec(text);
  if (gitMatch) return gitMatch[1];

  const plusMatch = /\+\+\+ b\/([^\s\n]+)/.exec(text);
  if (plusMatch) return plusMatch[1];

  const minusMatch = /--- a\/([^\s\n]+)/.exec(text);
  if (minusMatch) return minusMatch[1];

  return undefined;
}

/**
 * Robust diff parser that handles standard unified diffs as well as
 * truncated diffs, bare changes, orphan leading lines, and mid-diff byte truncations.
 */
export function parseRobustDiff(input: string): DiffFile[] {
  if (!input || typeof input !== 'string' || !input.trim()) return [];

  const lines = input.split(/\r?\n/);
  const files: DiffFile[] = [];
  let currentFile: DiffFile | null = null;
  let currentChunk: DiffChunk | null = null;
  let deletedLineCounter = 0;
  let addedLineCounter = 0;

  const startNewFile = (name = 'Diff Changes') => {
    currentFile = {
      chunks: [],
      deletions: 0,
      additions: 0,
      from: name,
      to: name,
    };
    files.push(currentFile);
    currentChunk = null;
    return currentFile;
  };

  const ensureFile = (name = 'Diff Changes') => {
    if (!currentFile) {
      startNewFile(name);
    }
    return currentFile!;
  };

  const ensureChunk = (header = '@@ -1,1 +1,1 @@') => {
    ensureFile();
    if (!currentChunk) {
      currentChunk = {
        content: header,
        changes: [],
        oldStart: 1,
        oldLines: 1,
        newStart: 1,
        newLines: 1,
      };
      deletedLineCounter = 1;
      addedLineCounter = 1;
      currentFile!.chunks.push(currentChunk);
    }
    return currentChunk;
  };

  for (let i = 0; i < lines.length; i++) {
    const rawLine = lines[i];
    const line = rawLine.replace(/\r$/, '');

    // Git file diff header
    if (line.startsWith('diff --git ')) {
      const parts = line.split(' ');
      const from = parts[2]?.replace(/^a\//, '');
      const to = parts[3]?.replace(/^b\//, '');
      startNewFile(to || from || 'file');
      currentFile!.from = from || 'file';
      currentFile!.to = to || 'file';
      continue;
    }

    // Old file indicator
    if (line.startsWith('--- ')) {
      const from = line.substring(4).trim().replace(/^a\//, '');
      if (currentChunk !== null) {
        // Multi-file diff without diff --git header
        startNewFile(from);
      } else {
        ensureFile(from);
      }
      currentFile!.from = from;
      continue;
    }

    // New file indicator
    if (line.startsWith('+++ ')) {
      const to = line.substring(4).trim().replace(/^b\//, '');
      ensureFile(to);
      currentFile!.to = to;
      continue;
    }

    // File mode indicators
    if (line.startsWith('new file mode ')) {
      ensureFile();
      currentFile!.new = true;
      continue;
    }

    if (line.startsWith('deleted file mode ')) {
      ensureFile();
      currentFile!.deleted = true;
      continue;
    }

    // Chunk (hunk) header
    const chunkMatch = line.match(/^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@(.*)/);
    if (chunkMatch) {
      ensureFile();
      const oldStart = parseInt(chunkMatch[1], 10);
      const oldLines = chunkMatch[2] ? parseInt(chunkMatch[2], 10) : 1;
      const newStart = parseInt(chunkMatch[3], 10);
      const newLines = chunkMatch[4] ? parseInt(chunkMatch[4], 10) : 1;

      deletedLineCounter = oldStart;
      addedLineCounter = newStart;

      currentChunk = {
        content: line,
        changes: [],
        oldStart,
        oldLines,
        newStart,
        newLines,
      };
      currentFile!.chunks.push(currentChunk);
      continue;
    }

    // Metadata lines before first chunk
    if (!currentChunk) {
      if (
        line.startsWith('index ') ||
        line.startsWith('similarity index ') ||
        line.startsWith('old mode ') ||
        line.startsWith('new mode ') ||
        /^<truncated\s+.*>$/i.test(line.trim())
      ) {
        continue;
      }
    }

    // Additions
    if (line.startsWith('+')) {
      ensureChunk();
      currentChunk!.changes.push({
        type: 'add',
        add: true,
        ln: addedLineCounter++,
        content: line,
      });
      currentFile!.additions++;
    } else if (line.startsWith('-')) {
      // Deletions
      ensureChunk();
      currentChunk!.changes.push({
        type: 'del',
        del: true,
        ln: deletedLineCounter++,
        content: line,
      });
      currentFile!.deletions++;
    } else if (line.startsWith(' ') || line === '') {
      // Normal context line
      if (currentChunk) {
        currentChunk.changes.push({
          type: 'normal',
          normal: true,
          ln1: deletedLineCounter++,
          ln2: addedLineCounter++,
          content: line,
        });
      }
    } else if (/^<truncated\s+.*>$/i.test(line.trim())) {
      // Embedded truncation marker inside diff
      if (currentChunk) {
        currentChunk.changes.push({
          type: 'normal',
          normal: true,
          content: line,
        });
      }
    } else if (line.startsWith('\\ No newline at end of file')) {
      continue;
    } else {
      // Unindented context line
      if (currentChunk) {
        currentChunk.changes.push({
          type: 'normal',
          normal: true,
          ln1: deletedLineCounter++,
          ln2: addedLineCounter++,
          content: line,
        });
      }
    }
  }

  return files.filter((f) => f.chunks.length > 0);
}
