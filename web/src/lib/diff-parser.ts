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
  oldStart: number;
  oldLines: number;
  newStart: number;
  newLines: number;
}

export interface DiffFile {
  from?: string;
  to?: string;
  chunks: DiffChunk[];
  deletions: number;
  additions: number;
  new?: boolean;
  deleted?: boolean;
  oldMode?: string;
  newMode?: string;
  index?: string[];
  isRename?: boolean;
  isCopy?: boolean;
  isBinary?: boolean;
  similarityIndex?: number;
}

const basePrefixes = ['a/', 'b/', 'i/', 'w/', 'c/', 'o/', '1/', '2/'];

/**
 * Normalizes file paths by stripping quotes, GNU diff timestamps,
 * and git mnemonic prefixes (a/, b/, i/, w/, c/, o/, 1/, 2/).
 * Inspired by parse-diff and diff2html path cleaners.
 */
export function cleanFilename(raw?: string): string {
  if (!raw) return '';
  let name = raw.trim();

  // Strip GNU diff timestamp (tab or space followed by date/time/timezone)
  // e.g. "--- lao\t2002-02-21 23:30:39.942229878 -0800"
  name = name.replace(/\t.*$/, '').replace(/\s+\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}(?:\.\d+)?\s+[+-]\d{4}.*$/, '');

  // Strip surrounding quotes
  if (
    (name.startsWith('"') && name.endsWith('"')) ||
    (name.startsWith("'") && name.endsWith("'"))
  ) {
    name = name.slice(1, -1);
  }

  // Strip git mnemonic prefixes
  for (const prefix of basePrefixes) {
    if (name.startsWith(prefix)) {
      name = name.slice(prefix.length);
      break;
    }
  }

  // Remove any remaining escaped quotes
  return name.replace(/^\\?["']|\\?["']$/g, '');
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

  // 1. Explicit git / hg / svn diff markers
  if (trimmed.includes('diff --git ') || trimmed.includes('diff -r ') || trimmed.startsWith('Index: ')) return true;
  if (trimmed.includes('--- a/') && trimmed.includes('+++ b/')) return true;
  if (trimmed.includes('Binary files ') && trimmed.includes(' differ')) return true;

  // 2. Standard unified diff hunk header accompanied by change lines
  const hasHunkHeader = /^@@@?\s+-(\d+)(?:,\d+)?(?:\s+-\d+(?:,\d+)?)?\s+\+(\d+)(?:,\d+)?\s+@@@?/m.test(trimmed);
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

  // 1. diff --git header (handles quoted paths with spaces)
  const gitMatch = /^diff --git\s+(?:"a\/(.+?)"|a\/(.+?))\s+(?:"b\/(.+?)"|b\/(.+?))$/m.exec(text);
  if (gitMatch) return gitMatch[3] || gitMatch[4] || gitMatch[1] || gitMatch[2];

  // 2. +++ or --- header
  const plusMatch = /^\+\+\+\s+(?:b\/)?([^\s\n\t]+)/m.exec(text);
  if (plusMatch && plusMatch[1] !== '/dev/null') return cleanFilename(plusMatch[1]);

  const minusMatch = /^---\s+(?:a\/)?([^\s\n\t]+)/m.exec(text);
  if (minusMatch && minusMatch[1] !== '/dev/null') return cleanFilename(minusMatch[1]);

  // 3. Rename target
  const renameMatch = /^rename to "?(.+?)"?$/m.exec(text);
  if (renameMatch) return cleanFilename(renameMatch[1]);

  return undefined;
}

/**
 * Fault-tolerant diff parser inspired by parse-diff and diff2html.
 * Handles:
 * - Standard Git unified diffs
 * - GNU unified diffs with timestamps
 * - Combined 3-way merge diffs (@@@)
 * - File renames, copies, mode changes, and binary files
 * - Escaped quoted paths with spaces
 * - Missing/truncated headers and mid-stream byte truncations
 * - Orphan changes before first @@ hunk
 * - diff.suppressBlankEmpty blank context lines
 * - "\ No newline at end of file" EOF markers
 */
export function parseRobustDiff(input: string): DiffFile[] {
  if (!input || typeof input !== 'string' || !input.trim()) return [];

  const lines = input.split(/\r?\n/);
  const files: DiffFile[] = [];
  let currentFile: DiffFile | null = null;
  let currentChunk: DiffChunk | null = null;
  let deletedLineCounter = 0;
  let addedLineCounter = 0;

  const startNewFile = (from?: string, to?: string) => {
    currentFile = {
      from: from || undefined,
      to: to || undefined,
      chunks: [],
      deletions: 0,
      additions: 0,
    };
    files.push(currentFile);
    currentChunk = null;
    return currentFile;
  };

  const ensureFile = (name = 'Diff Changes') => {
    if (!currentFile) {
      startNewFile(name, name);
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

    // 1. Git / Mercurial / SVN File Headers
    if (line.startsWith('diff --git ') || line.startsWith('diff --combined ')) {
      const match = /^diff --(?:git|combined)\s+(?:"[a-ciow12]\/(.+?)"|[a-ciow12]\/(.+?))\s+(?:"[a-ciow12]\/(.+?)"|[a-ciow12]\/(.+?))$/.exec(line);
      let from: string | undefined;
      let to: string | undefined;
      if (match) {
        from = match[1] || match[2];
        to = match[3] || match[4];
      } else {
        const parts = line.split(' ');
        from = cleanFilename(parts[2]);
        to = cleanFilename(parts[3]);
      }
      startNewFile(from, to);
      continue;
    }

    if (line.startsWith('Index: ')) {
      const name = cleanFilename(line.substring(7));
      startNewFile(name, name);
      continue;
    }

    // 2. Git Extended Metadata Headers
    if (line.startsWith('new file mode ')) {
      ensureFile();
      currentFile!.new = true;
      currentFile!.newMode = line.substring(14).trim();
      currentFile!.from = '/dev/null';
      continue;
    }

    if (line.startsWith('deleted file mode ')) {
      ensureFile();
      currentFile!.deleted = true;
      currentFile!.oldMode = line.substring(18).trim();
      currentFile!.to = '/dev/null';
      continue;
    }

    if (line.startsWith('old mode ')) {
      ensureFile();
      currentFile!.oldMode = line.substring(9).trim();
      continue;
    }

    if (line.startsWith('new mode ')) {
      ensureFile();
      currentFile!.newMode = line.substring(9).trim();
      continue;
    }

    if (line.startsWith('similarity index ')) {
      ensureFile();
      currentFile!.similarityIndex = parseInt(line.substring(17), 10);
      continue;
    }

    if (line.startsWith('rename from ')) {
      ensureFile();
      currentFile!.from = cleanFilename(line.substring(12));
      currentFile!.isRename = true;
      continue;
    }

    if (line.startsWith('rename to ')) {
      ensureFile();
      currentFile!.to = cleanFilename(line.substring(10));
      currentFile!.isRename = true;
      continue;
    }

    if (line.startsWith('copy from ')) {
      ensureFile();
      currentFile!.from = cleanFilename(line.substring(10));
      currentFile!.isCopy = true;
      continue;
    }

    if (line.startsWith('copy to ')) {
      ensureFile();
      currentFile!.to = cleanFilename(line.substring(8));
      currentFile!.isCopy = true;
      continue;
    }

    if (line.startsWith('index ')) {
      ensureFile();
      const idxParts = line.substring(6).split(' ');
      currentFile!.index = idxParts[0]?.split('..');
      if (idxParts[1]) {
        currentFile!.oldMode = currentFile!.newMode = idxParts[1].trim();
      }
      continue;
    }

    if (line.startsWith('Binary files ') || line.startsWith('GIT binary patch')) {
      ensureFile();
      currentFile!.isBinary = true;
      const binMatch = /^Binary files\s+(.+?)\s+and\s+(.+?)\s+differ/.exec(line);
      if (binMatch) {
        currentFile!.from = cleanFilename(binMatch[1]);
        currentFile!.to = cleanFilename(binMatch[2]);
      }
      continue;
    }

    // 3. Old / New file indicators (--- and +++)
    if (line.startsWith('--- ')) {
      const from = cleanFilename(line.substring(4));
      if (currentChunk !== null) {
        // Multi-file diff without diff --git header
        startNewFile(from, undefined);
      } else {
        ensureFile(from);
      }
      currentFile!.from = from === '/dev/null' ? '/dev/null' : from;
      if (from === '/dev/null') currentFile!.new = true;
      continue;
    }

    if (line.startsWith('+++ ')) {
      const to = cleanFilename(line.substring(4));
      ensureFile(to);
      currentFile!.to = to === '/dev/null' ? '/dev/null' : to;
      if (to === '/dev/null') currentFile!.deleted = true;
      continue;
    }

    // 4. Hunk Header: Standard @@ or 3-way Combined @@@
    const chunkMatch = line.match(/^@@@?\s+-(\d+)(?:,(\d+))?(?:\s+-\d+(?:,\d+)?)?\s+\+(\d+)(?:,(\d+))?\s+@@@?(.*)/);
    if (chunkMatch) {
      ensureFile();
      const oldStart = parseInt(chunkMatch[1], 10);
      const oldLines = chunkMatch[2] !== undefined ? parseInt(chunkMatch[2], 10) : (oldStart === 0 ? 0 : 1);
      const newStart = parseInt(chunkMatch[3], 10);
      const newLines = chunkMatch[4] !== undefined ? parseInt(chunkMatch[4], 10) : (newStart === 0 ? 0 : 1);

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

    // 5. EOF newline marker: \ No newline at end of file
    if (line.startsWith('\\ No newline at end of file')) {
      if (currentChunk && currentChunk.changes.length > 0) {
        const lastChange = currentChunk.changes[currentChunk.changes.length - 1];
        currentChunk.changes.push({
          type: lastChange.type,
          [lastChange.type]: true,
          ln1: lastChange.ln1,
          ln2: lastChange.ln2,
          ln: lastChange.ln,
          content: line,
        });
      }
      continue;
    }

    // Ignore metadata before first chunk
    if (!currentChunk) {
      if (/^<truncated\s+.*>$/i.test(line.trim())) {
        continue;
      }
    }

    // 6. Content Lines (+, -, space, or blank context line)
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
      ensureChunk();
      currentChunk!.changes.push({
        type: 'del',
        del: true,
        ln: deletedLineCounter++,
        content: line,
      });
      currentFile!.deletions++;
    } else if (line.startsWith(' ') || line === '') {
      // Normal context line (including blank lines with diff.suppressBlankEmpty)
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
      // Embedded truncation marker inside diff body
      if (currentChunk) {
        currentChunk.changes.push({
          type: 'normal',
          normal: true,
          content: line,
        });
      }
    } else {
      // Unindented or reformatted context line
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

  return files.filter(
    (f) =>
      f.chunks.length > 0 ||
      f.isRename ||
      f.isCopy ||
      f.isBinary ||
      f.new ||
      f.deleted ||
      (f.oldMode !== undefined && f.newMode !== undefined)
  );
}
