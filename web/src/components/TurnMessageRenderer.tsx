import React, { useState, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  Copy,
  Check,
  Terminal,
  Clock,
  FileEdit,
  FileCode2,
} from 'lucide-react';
import { DiffViewer } from './DiffViewer';
import { isDiffText, extractDiffFilePath } from '@/lib/diff-parser';
import { Button } from './ui/button';
import { cn } from '@/lib/utils';

interface TurnMessageRendererProps {
  content: string;
}

interface FencedCodeBlockProps {
  text: string;
  language: string;
}

function formatBytes(bytes?: number): string {
  if (!bytes || bytes <= 0) return '';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function getLanguageFromPath(filePath?: string): { lang: string; badge: string } {
  if (!filePath) return { lang: 'text', badge: 'TEXT' };
  const clean = filePath.replace(/[`"']/g, '');
  const ext = clean.split('.').pop()?.toLowerCase() || '';
  switch (ext) {
    case 'ts':
      return { lang: 'typescript', badge: 'TypeScript' };
    case 'tsx':
      return { lang: 'typescript', badge: 'TSX' };
    case 'js':
      return { lang: 'javascript', badge: 'JavaScript' };
    case 'jsx':
      return { lang: 'javascript', badge: 'JSX' };
    case 'go':
      return { lang: 'go', badge: 'Go' };
    case 'py':
      return { lang: 'python', badge: 'Python' };
    case 'json':
      return { lang: 'json', badge: 'JSON' };
    case 'css':
      return { lang: 'css', badge: 'CSS' };
    case 'html':
      return { lang: 'html', badge: 'HTML' };
    case 'sh':
    case 'bash':
    case 'zsh':
      return { lang: 'shell', badge: 'Shell' };
    case 'rs':
      return { lang: 'rust', badge: 'Rust' };
    case 'yaml':
    case 'yml':
      return { lang: 'yaml', badge: 'YAML' };
    case 'md':
      return { lang: 'markdown', badge: 'Markdown' };
    case 'sql':
      return { lang: 'sql', badge: 'SQL' };
    default:
      return { lang: ext, badge: ext ? ext.toUpperCase() : 'CODE' };
  }
}

export interface FileCodeLine {
  lineNum: number | string;
  code: string;
}

export interface FileCodeBlockProps {
  filePath?: string;
  totalLines?: number;
  totalBytes?: number;
  startLine?: number;
  endLine?: number;
  lines: FileCodeLine[];
  rawCleanCode: string;
  createdAt?: string;
  completedAt?: string;
  truncatedNote?: string;
}

export function FileCodeBlock({
  filePath,
  totalLines,
  totalBytes,
  startLine,
  endLine,
  lines,
  rawCleanCode,
  completedAt,
  truncatedNote,
}: FileCodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);

  const fileName = useMemo(() => {
    if (!filePath) return 'File View';
    const clean = filePath.replace(/^file:\/\//, '').replace(/`/g, '');
    const parts = clean.split(/[/\\]/);
    return parts[parts.length - 1] || 'File View';
  }, [filePath]);

  const langInfo = useMemo(() => getLanguageFromPath(filePath), [filePath]);

  const displayPath = useMemo(() => {
    if (!filePath) return undefined;
    const clean = filePath.replace(/^file:\/\//, '').replace(/`/g, '');
    const parts = clean.split(/[/\\]/);
    if (parts.length > 3) {
      return '…/' + parts.slice(-3).join('/');
    }
    return clean;
  }, [filePath]);

  const lineRangeStr = useMemo(() => {
    if (startLine && endLine) {
      return `L${startLine}–L${endLine}`;
    }
    if (lines.length > 0) {
      const first = lines.find((l) => typeof l.lineNum === 'number');
      const last = [...lines].reverse().find((l) => typeof l.lineNum === 'number');
      if (first && last) {
        return `L${first.lineNum}–L${last.lineNum}`;
      }
    }
    return null;
  }, [startLine, endLine, lines]);

  const statsStr = useMemo(() => {
    const parts: string[] = [];
    if (totalBytes) parts.push(formatBytes(totalBytes));
    if (totalLines) parts.push(`${totalLines.toLocaleString()} total lines`);
    return parts.join(' · ');
  }, [totalBytes, totalLines]);

  const handleCopyCleanCode = () => {
    navigator.clipboard.writeText(rawCleanCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const shouldShowExpandToggle = lines.length > 60;

  return (
    <div className="rounded-[6px] border border-[#2a2a2a] bg-[#0c0c0c] overflow-hidden my-3 shadow-md">
      {/* File Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2 bg-[#121212] border-b border-[#222222]">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <FileCode2 className="h-3.5 w-3.5 text-sky-400 shrink-0" />
          <span className="font-mono text-xs font-semibold text-[#ededed] tracking-tight">
            {fileName}
          </span>
          {lineRangeStr && (
            <span className="px-1.5 py-0.5 rounded bg-[#1c1c1c] border border-[#2a2a2a] text-[10px] font-mono text-[#a1a1a1]">
              {lineRangeStr}
            </span>
          )}
          <span className="px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 text-[10px] font-mono font-medium">
            {langInfo.badge}
          </span>
          {displayPath && (
            <span
              className="text-[10px] font-mono text-[#777777] truncate max-w-[280px] hidden sm:inline"
              title={filePath}
            >
              {displayPath}
            </span>
          )}
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {statsStr && (
            <span className="text-[10px] font-mono text-[#666666] hidden md:inline">
              {statsStr}
            </span>
          )}
          {completedAt && (
            <span className="inline-flex items-center gap-1 text-[10px] font-mono text-[#666666] hidden lg:inline">
              <Clock className="h-2.5 w-2.5 text-[#555555]" />
              {completedAt}
            </span>
          )}
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleCopyCleanCode}
            className="h-6 px-2 text-[10px] font-sans text-muted-foreground hover:text-foreground hover:bg-[#222222] rounded-[3px] gap-1 cursor-pointer"
            title="Copy clean code (without line numbers)"
          >
            {copied ? (
              <>
                <Check className="h-3 w-3 text-status-success" />
                <span className="text-status-success font-medium">Copied Code</span>
              </>
            ) : (
              <>
                <Copy className="h-3 w-3" />
                <span>Copy Code</span>
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Code Viewer Body */}
      <div
        className={cn(
          'overflow-auto bg-[#080808] transition-all',
          isExpanded ? 'max-h-none' : 'max-h-[560px]'
        )}
      >
        <table className="w-full border-collapse text-left font-mono text-[11px] leading-relaxed">
          <tbody>
            {lines.map((l, idx) => (
              <tr key={idx} className="hover:bg-white/[0.03] group transition-colors">
                <td className="sticky left-0 z-10 w-12 sm:w-16 select-none text-right pr-3.5 pl-2 text-[#555555] group-hover:text-[#888888] border-r border-[#1c1c1c] bg-[#0c0c0c] font-mono text-[11px] align-top tabular-nums">
                  {l.lineNum !== '' ? l.lineNum : ''}
                </td>
                <td className="pl-3.5 pr-4 py-0 font-mono text-[11px] text-[#ededed] whitespace-pre align-top select-text">
                  {l.code || '\n'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Footer / Stats bar */}
      {(shouldShowExpandToggle || truncatedNote) && (
        <div className="flex items-center justify-between px-3 py-1.5 bg-[#101010] border-t border-[#1f1f1f] text-[10px] font-mono text-[#777777]">
          <div className="flex items-center gap-1.5 truncate">
            {truncatedNote ? (
              <span className="truncate">{truncatedNote}</span>
            ) : (
              <span>Showing {lines.length} lines</span>
            )}
          </div>
          {shouldShowExpandToggle && (
            <button
              type="button"
              onClick={() => setIsExpanded(!isExpanded)}
              className="text-muted-foreground hover:text-foreground text-[10px] font-sans hover:underline cursor-pointer"
            >
              {isExpanded ? 'Collapse height' : 'Expand full height'}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

export function parseFileViewContent(text: string): FileCodeBlockProps | null {
  if (!text || typeof text !== 'string') return null;

  const filePathMatch = text.match(/File Path:\s*(?:`file:\/\/|`|file:\/\/)?([^`\r\n]+)`?/i);
  const showingLinesMatch = text.match(/Showing lines\s*(\d+)\s+to\s+(\d+)/i);
  const totalLinesMatch = text.match(/Total Lines:\s*(\d+)/i);
  const totalBytesMatch = text.match(/Total Bytes:\s*(\d+)/i);
  const createdAtMatch = text.match(/Created At:\s*([^\r\n]+)/i);
  const completedAtMatch = text.match(/Completed At:\s*([^\r\n]+)/i);

  const rawLines = text.split('\n');
  const codeLines: FileCodeLine[] = [];
  const cleanCodeArr: string[] = [];
  let foundCode = false;
  let truncatedNote: string | undefined = undefined;

  for (let i = 0; i < rawLines.length; i++) {
    const rawLine = rawLines[i];
    const lineMatch = rawLine.match(/^\s*(\d+):\s?(.*)$/);

    if (lineMatch) {
      foundCode = true;
      const num = parseInt(lineMatch[1], 10);
      const code = lineMatch[2];
      codeLines.push({ lineNum: num, code });
      cleanCodeArr.push(code);
    } else if (foundCode) {
      if (
        rawLine.includes('The following code has been modified') ||
        rawLine.includes('Please note that any changes targeting')
      ) {
        continue;
      }
      if (rawLine.includes('The above content does NOT show the entire file contents')) {
        truncatedNote = 'Content truncated to limit · View with offset or slice for more lines';
        continue;
      }

      // Check if there are any numbered lines remaining ahead
      const hasMoreNumberedLines = rawLines.slice(i).some((l) => /^\s*\d+:\s?/.test(l));
      if (!hasMoreNumberedLines) {
        // We reached the end of the numbered code snippet, do not swallow trailing chat prose
        break;
      }

      if (rawLine.trim() === '') {
        codeLines.push({ lineNum: '', code: '' });
        cleanCodeArr.push('');
      } else {
        codeLines.push({ lineNum: '', code: rawLine });
        cleanCodeArr.push(rawLine);
      }
    }
  }

  if (codeLines.length > 0 && (filePathMatch || showingLinesMatch || codeLines.length >= 3)) {
    const rawPath = filePathMatch ? filePathMatch[1].trim() : undefined;
    const cleanPath = rawPath ? rawPath.replace(/^file:\/\//, '').replace(/`/g, '') : undefined;
    const totalLines = totalLinesMatch ? parseInt(totalLinesMatch[1], 10) : undefined;
    const totalBytes = totalBytesMatch ? parseInt(totalBytesMatch[1], 10) : undefined;
    const startLine = showingLinesMatch
      ? parseInt(showingLinesMatch[1], 10)
      : typeof codeLines[0]?.lineNum === 'number'
      ? codeLines[0].lineNum
      : undefined;
    const lastNumLine = [...codeLines].reverse().find((l) => typeof l.lineNum === 'number');
    const endLine = showingLinesMatch
      ? parseInt(showingLinesMatch[2], 10)
      : typeof lastNumLine?.lineNum === 'number'
      ? lastNumLine.lineNum
      : undefined;

    return {
      filePath: cleanPath,
      totalLines,
      totalBytes,
      startLine,
      endLine,
      lines: codeLines,
      rawCleanCode: cleanCodeArr.join('\n'),
      createdAt: createdAtMatch ? createdAtMatch[1].trim() : undefined,
      completedAt: completedAtMatch ? completedAtMatch[1].trim() : undefined,
      truncatedNote,
    };
  }

  return null;
}

function FencedCodeBlock({ text, language }: FencedCodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const isDiff = isDiffText(text, language);

  if (isDiff) {
    const detectedPath = extractDiffFilePath(text);
    return <DiffViewer diffText={text} filePath={detectedPath} />;
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="rounded-[6px] border border-border bg-[#0c0c0c] overflow-hidden my-2.5 shadow-sm">
      <div className="flex items-center justify-between px-3 py-1.5 bg-[#141414] border-b border-border">
        <span className="text-[10px] font-mono font-medium text-muted-foreground uppercase tracking-wider">
          {language || 'code'}
        </span>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleCopy}
          className="h-5 px-1.5 text-[10px] font-sans text-muted-foreground hover:text-foreground hover:bg-[#222222] rounded-[3px] gap-1"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3 text-status-success" />
              <span className="text-status-success">Copied</span>
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" />
              <span>Copy</span>
            </>
          )}
        </Button>
      </div>
      <pre className="p-3 overflow-x-auto text-[11px] font-mono text-[#d4d4d4] leading-relaxed">
        <code>{text}</code>
      </pre>
    </div>
  );
}

export function TurnMessageRenderer({ content }: TurnMessageRendererProps) {
  // 1. Check for diff blocks formatted like:
  // [diff_block_start] ... [diff_block_end]
  const diffBlockRegex = /\[diff_block_start\]([\s\S]*?)\[diff_block_end\]/g;

  if (diffBlockRegex.test(content)) {
    diffBlockRegex.lastIndex = 0;
    const segments: React.ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;

    let segmentKey = 0;
    while ((match = diffBlockRegex.exec(content)) !== null) {
      const precedingText = content.substring(lastIndex, match.index);
      const diffContent = match[1];
      lastIndex = match.index + match[0].length;

      let detectedPath: string | undefined;
      const pathMatch = /to:?\s+([^\s\n]+?)(?:\.\s|\.$|\:|\n|$)/i.exec(precedingText);
      if (pathMatch) {
        detectedPath = pathMatch[1].replace(/[.:,;]$/, '');
      }

      if (precedingText.trim()) {
        segments.push(
          <div key={`text-${segmentKey++}`}>
            <StructuredTextBlock text={precedingText} />
          </div>
        );
      }

      segments.push(
        <DiffViewer
          key={`diff-${segmentKey++}`}
          diffText={diffContent}
          filePath={detectedPath}
        />
      );
    }

    const remainingText = content.substring(lastIndex);
    if (remainingText.trim()) {
      segments.push(
        <div key={`text-${segmentKey++}`}>
          <StructuredTextBlock text={remainingText} />
        </div>
      );
    }

    return <div className="space-y-2">{segments}</div>;
  }

  // 2. Check for unified diff / git diff block
  if (isDiffText(content)) {
    const detectedPath = extractDiffFilePath(content);
    return <DiffViewer diffText={content} filePath={detectedPath} />;
  }

  // 3. Check for File View Block
  const fileBlockStartIndex = content.search(/(?:Created At:[^\n]+\n(?:Completed At:[^\n]+\n)?)?File Path:\s*(?:`file:\/\/|`|file:\/\/)?/i);
  if (fileBlockStartIndex !== -1) {
    const precedingText = content.substring(0, fileBlockStartIndex).trim();
    const fileContent = content.substring(fileBlockStartIndex);
    const fileView = parseFileViewContent(fileContent);

    if (fileView) {
      return (
        <div className="space-y-2">
          {precedingText && <StructuredTextBlock text={precedingText} />}
          <FileCodeBlock {...fileView} />
        </div>
      );
    }
  }

  const directFileView = parseFileViewContent(content);
  if (directFileView) {
    return <FileCodeBlock {...directFileView} />;
  }

  // 4. Render standard structured text
  return <StructuredTextBlock text={content} />;
}

/**
 * Parses tool output preambles, command exit codes, and renders markdown.
 */
function StructuredTextBlock({ text }: { text: string }) {
  const lines = text.split('\n');
  const metadataLines: string[] = [];
  let contentStartIndex = 0;

  for (let i = 0; i < Math.min(lines.length, 6); i++) {
    const l = lines[i].trim();
    if (
      l.startsWith('Created At:') ||
      l.startsWith('Completed At:') ||
      l.startsWith('The command exited with code') ||
      l.startsWith('Output:') ||
      l.startsWith('The following changes were made by the')
    ) {
      if (l) metadataLines.push(l);
      contentStartIndex = i + 1;
    } else if (l === '') {
      contentStartIndex = i + 1;
    } else {
      break;
    }
  }

  const hasMetadata = metadataLines.length > 0;
  const mainContent = hasMetadata ? lines.slice(contentStartIndex).join('\n').trim() : text;

  // Extract exit code or tool action if present
  let exitCode: number | null = null;
  let toolAction: string | null = null;

  for (const meta of metadataLines) {
    const exitMatch = meta.match(/exited with code (\d+)/);
    if (exitMatch) {
      exitCode = parseInt(exitMatch[1], 10);
    }
    const toolMatch = meta.match(/made by the ([a-zA-Z0-9_-]+) tool/);
    if (toolMatch) {
      toolAction = toolMatch[1];
    }
  }

  // Check if mainContent itself is a file view block
  const fileView = useMemo(() => parseFileViewContent(mainContent), [mainContent]);
  if (fileView) {
    return (
      <div className="space-y-2">
        {hasMetadata && (
          <div className="flex flex-wrap items-center gap-2 p-2 rounded-[5px] bg-[#111111] border border-[#222222] text-[11px] font-sans text-[#a1a1a1]">
            {toolAction && (
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 font-mono text-[10px]">
                <FileEdit className="h-3 w-3" />
                {toolAction}
              </span>
            )}
            {exitCode !== null && (
              <span
                className={`inline-flex items-center gap-1 font-mono text-[10px] font-medium ${
                  exitCode === 0 ? 'badge-success' : 'badge-danger'
                }`}
              >
                <Terminal className="h-2.5 w-2.5" />
                exit code {exitCode}
              </span>
            )}
            {metadataLines.some((m) => m.startsWith('Completed At:')) && (
              <span className="inline-flex items-center gap-1 text-[10px] text-[#777777]">
                <Clock className="h-2.5 w-2.5 text-[#555555]" />
                {metadataLines.find((m) => m.startsWith('Completed At:'))?.replace('Completed At: ', '')}
              </span>
            )}
          </div>
        )}
        <FileCodeBlock {...fileView} />
      </div>
    );
  }

  // Check if mainContent itself is a diff
  const isDiff = useMemo(() => isDiffText(mainContent), [mainContent]);
  if (isDiff) {
    const detectedPath = extractDiffFilePath(mainContent);
    return (
      <div className="space-y-2">
        {hasMetadata && (
          <div className="flex flex-wrap items-center gap-2 p-2 rounded-[5px] bg-[#111111] border border-[#222222] text-[11px] font-sans text-[#a1a1a1]">
            {toolAction && (
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 font-mono text-[10px]">
                <FileEdit className="h-3 w-3" />
                {toolAction}
              </span>
            )}
            {exitCode !== null && (
              <span
                className={`inline-flex items-center gap-1 font-mono text-[10px] font-medium ${
                  exitCode === 0 ? 'badge-success' : 'badge-danger'
                }`}
              >
                <Terminal className="h-2.5 w-2.5" />
                exit code {exitCode}
              </span>
            )}
            {metadataLines.some((m) => m.startsWith('Completed At:')) && (
              <span className="inline-flex items-center gap-1 text-[10px] text-[#777777]">
                <Clock className="h-2.5 w-2.5 text-[#555555]" />
                {metadataLines.find((m) => m.startsWith('Completed At:'))?.replace('Completed At: ', '')}
              </span>
            )}
          </div>
        )}
        <DiffViewer diffText={mainContent} filePath={detectedPath} />
      </div>
    );
  }

  // Detect whether output is raw preformatted terminal stdout
  const isPreformattedStdout = useMemo(() => {
    if (!mainContent) return false;
    if (exitCode !== null) {
      const hasMarkdownHeader = /^(#{1,6}\s|[*-]\s\*\*)/m.test(mainContent);
      return !hasMarkdownHeader;
    }
    if (/^[-d][rwx-]{9}/m.test(mainContent)) return true;
    const contentLines = mainContent.split('\n').filter((l) => l.trim().length > 0);
    if (contentLines.length >= 3) {
      const lineNumMatches = contentLines.filter((l) => /^\s*\d+:\s+/.test(l));
      if (lineNumMatches.length / contentLines.length > 0.4) {
        return true;
      }
    }
    return false;
  }, [exitCode, mainContent]);

  return (
    <div className="space-y-2">
      {hasMetadata && (
        <div className="flex flex-wrap items-center gap-2 p-2 rounded-[5px] bg-[#111111] border border-[#222222] text-[11px] font-sans text-[#a1a1a1]">
          {toolAction && (
            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20 font-mono text-[10px]">
              <FileEdit className="h-3 w-3" />
              {toolAction}
            </span>
          )}
          {exitCode !== null && (
            <span
              className={`inline-flex items-center gap-1 font-mono text-[10px] font-medium ${
                exitCode === 0 ? 'badge-success' : 'badge-danger'
              }`}
            >
              <Terminal className="h-2.5 w-2.5" />
              exit code {exitCode}
            </span>
          )}
          {metadataLines.some((m) => m.startsWith('Completed At:')) && (
            <span className="inline-flex items-center gap-1 text-[10px] text-[#777777]">
              <Clock className="h-2.5 w-2.5 text-[#555555]" />
              {metadataLines.find((m) => m.startsWith('Completed At:'))?.replace('Completed At: ', '')}
            </span>
          )}
        </div>
      )}

      {mainContent && (
        isPreformattedStdout ? (
          <pre className="p-3 rounded-[4px] bg-[#0c0c0c] border border-[#222222] font-mono text-[11px] text-[#d4d4d4] overflow-x-auto whitespace-pre leading-relaxed">
            {mainContent}
          </pre>
        ) : (
          <div className="prose-subtle text-xs text-[#ededed] font-sans leading-relaxed selection:bg-[#333333]">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                h1: ({ children }) => (
                  <h1 className="text-base font-semibold text-white mt-4 mb-2 pb-1 border-b border-[#262626]">
                    {children}
                  </h1>
                ),
                h2: ({ children }) => (
                  <h2 className="text-sm font-semibold text-white mt-3 mb-1.5">
                    {children}
                  </h2>
                ),
                h3: ({ children }) => (
                  <h3 className="text-xs font-semibold text-[#ededed] mt-2 mb-1">
                    {children}
                  </h3>
                ),
                p: ({ children }) => (
                  <div className="mb-2 last:mb-0 leading-relaxed text-[#d4d4d4] whitespace-pre-wrap">
                    {children}
                  </div>
                ),
                ul: ({ children }) => (
                  <ul className="list-disc pl-5 mb-2 space-y-1 text-[#d4d4d4]">
                    {children}
                  </ul>
                ),
                ol: ({ children }) => (
                  <ol className="list-decimal pl-5 mb-2 space-y-1 text-[#d4d4d4]">
                    {children}
                  </ol>
                ),
                li: ({ children }) => <li className="leading-relaxed">{children}</li>,
                blockquote: ({ children }) => (
                  <blockquote className="border-l-2 border-[#333333] pl-3 py-0.5 text-[#a1a1a1] italic my-2">
                    {children}
                  </blockquote>
                ),
                table: ({ children }) => (
                  <div className="overflow-x-auto my-3 border border-[#262626] rounded-[4px]">
                    <table className="w-full text-left text-xs divide-y divide-[#262626]">
                      {children}
                    </table>
                  </div>
                ),
                thead: ({ children }) => (
                  <thead className="bg-[#141414] text-[#ededed] font-medium">{children}</thead>
                ),
                tbody: ({ children }) => (
                  <tbody className="divide-y divide-[#1f1f1f] bg-[#0c0c0c]">{children}</tbody>
                ),
                tr: ({ children }) => <tr className="hover:bg-white/[0.02]">{children}</tr>,
                th: ({ children }) => <th className="px-3 py-1.5 font-medium">{children}</th>,
                td: ({ children }) => <td className="px-3 py-1.5 text-[#a1a1a1]">{children}</td>,
                a: ({ href, children }) => (
                  <a
                    href={href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sky-400 hover:text-sky-300 underline underline-offset-2 transition-colors"
                  >
                    {children}
                  </a>
                ),
                pre: ({ children }: any) => {
                  const codeEl = React.isValidElement(children) ? children : null;
                  const codeProps = (codeEl?.props || {}) as { className?: string; children?: React.ReactNode };
                  const className = codeProps.className || '';
                  const match = /language-(\w+)/.exec(className);
                  const language = match ? match[1] : '';
                  const raw = codeProps.children;
                  const codeText = Array.isArray(raw) ? raw.join('') : String(raw || '').replace(/\n$/, '');

                  return <FencedCodeBlock text={codeText} language={language} />;
                },
                code: ({ children, ...props }: any) => {
                  return (
                    <code className="inline-code" {...props}>
                      {children}
                    </code>
                  );
                },
              }}
            >
              {mainContent}
            </ReactMarkdown>
          </div>
        )
      )}
    </div>
  );
}
