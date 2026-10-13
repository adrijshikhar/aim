import React, { useState } from 'react';
import { FileCode, Copy, Check, ChevronDown, ChevronRight, PlusCircle, MinusCircle, AlertTriangle } from 'lucide-react';
import { parseRobustDiff, type DiffFile } from '@/lib/diff-parser';
import { Button } from './ui/button';
import { cn } from '@/lib/utils';

interface DiffViewerProps {
  diffText: string;
  filePath?: string;
  defaultCollapsed?: boolean;
}

export function DiffViewer({ diffText, filePath, defaultCollapsed = false }: DiffViewerProps) {
  const [isCollapsed, setIsCollapsed] = useState(defaultCollapsed);
  const [copied, setCopied] = useState(false);

  // Clean diff markers if present
  const cleanedDiff = React.useMemo(() => {
    return diffText
      .replace(/\[diff_block_start\]\s*/g, '')
      .replace(/\[diff_block_end\]\s*/g, '')
      .trim();
  }, [diffText]);

  // Extract leading truncation notice if present
  const topTruncationNotice = React.useMemo(() => {
    const match = cleanedDiff.match(/^<truncated\s+([^>]+)>/im);
    return match ? match[1] : null;
  }, [cleanedDiff]);

  // Parse diff with robust AST parser
  const parsedFiles: DiffFile[] = React.useMemo(() => {
    if (!cleanedDiff) return [];
    return parseRobustDiff(cleanedDiff);
  }, [cleanedDiff]);

  // Compute total additions / deletions across all files
  const { totalAdditions, totalDeletions } = React.useMemo(() => {
    if (!parsedFiles.length) {
      let addCount = 0;
      let delCount = 0;
      for (const line of cleanedDiff.split('\n')) {
        if (line.startsWith('+') && !line.startsWith('+++')) addCount++;
        if (line.startsWith('-') && !line.startsWith('---')) delCount++;
      }
      return { totalAdditions: addCount, totalDeletions: delCount };
    }

    const adds = parsedFiles.reduce((acc, f) => acc + (f.additions || 0), 0);
    const dels = parsedFiles.reduce((acc, f) => acc + (f.deletions || 0), 0);
    return { totalAdditions: adds, totalDeletions: dels };
  }, [parsedFiles, cleanedDiff]);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(cleanedDiff);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Derive file title
  const displayFileName = React.useMemo(() => {
    if (filePath) {
      return filePath.split('/').pop() || filePath;
    }
    const isNullPath = (p?: string) => !p || p === '/dev/null' || p === 'dev/null' || p === 'Diff Changes' || p === 'file';
    const first = parsedFiles[0];
    if (first?.to && !isNullPath(first.to)) {
      return first.to.replace(/^b\//, '').split('/').pop() || first.to;
    }
    if (first?.from && !isNullPath(first.from)) {
      return first.from.replace(/^a\//, '').split('/').pop() || first.from;
    }
    // Check if any chunk header contains function/class context
    const firstHunk = first?.chunks?.find((c) => /@@.*?@@\s+(.+)/.test(c.content))?.content;
    if (firstHunk) {
      const match = /@@.*?@@\s+(?:export\s+(?:function\s+|const\s+|class\s+)?|function\s+|class\s+)?([a-zA-Z0-9_$]+)/.exec(firstHunk);
      if (match) {
        return `Diff (${match[1]})`;
      }
    }
    return 'Diff Changes';
  }, [filePath, parsedFiles]);

  const fullDisplayPath = React.useMemo(() => {
    if (filePath) return filePath;
    const isNullPath = (p?: string) => !p || p === '/dev/null' || p === 'dev/null' || p === 'Diff Changes' || p === 'file';
    const first = parsedFiles[0];
    if (first?.to && !isNullPath(first.to)) {
      return first.to.replace(/^b\//, '');
    }
    if (first?.from && !isNullPath(first.from)) {
      return first.from.replace(/^a\//, '');
    }
    return undefined;
  }, [filePath, parsedFiles]);

  return (
    <div className="rounded-[6px] border border-[#262626] bg-[#0d0d0d] overflow-hidden my-2.5 font-mono text-xs shadow-sm">
      {/* Header bar */}
      <div
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="flex items-center justify-between px-3 py-2 bg-[#141414] border-b border-[#262626] cursor-pointer select-none hover:bg-[#181818] transition-colors"
      >
        <div className="flex items-center gap-2 min-w-0">
          <button
            type="button"
            className="text-[#888888] hover:text-[#ededed] p-0.5"
            aria-label={isCollapsed ? 'Expand diff' : 'Collapse diff'}
          >
            {isCollapsed ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
          </button>
          <FileCode className="h-3.5 w-3.5 text-[#52a8ff] shrink-0" />
          <span className="text-[#ededed] font-medium text-[11px] truncate" title={fullDisplayPath || displayFileName}>
            {displayFileName}
          </span>
          {fullDisplayPath && fullDisplayPath !== displayFileName && (
            <span className="text-[#666666] text-[10px] truncate hidden sm:inline" title={fullDisplayPath}>
              {fullDisplayPath}
            </span>
          )}

          {/* Semantic Badges from Theme */}
          <div className="flex items-center gap-1.5 ml-1">
            {totalAdditions > 0 && (
              <span className="badge-success text-[10px] px-1.5 py-0.5">
                +{totalAdditions}
              </span>
            )}
            {totalDeletions > 0 && (
              <span className="badge-danger text-[10px] px-1.5 py-0.5">
                -{totalDeletions}
              </span>
            )}
            {parsedFiles.length === 1 && parsedFiles[0].isRename && (
              <span className="badge-warning text-[10px] px-1.5 py-0.5">
                renamed {parsedFiles[0].similarityIndex ? `(${parsedFiles[0].similarityIndex}%)` : ''}
              </span>
            )}
            {parsedFiles.length === 1 && parsedFiles[0].isBinary && (
              <span className="text-[10px] px-1.5 py-0.5 rounded font-mono bg-purple-500/10 text-purple-400 border border-purple-500/20">
                binary
              </span>
            )}
            {topTruncationNotice && (
              <span className="badge-warning text-[10px] px-1.5 py-0.5 inline-flex items-center gap-1">
                <AlertTriangle className="h-2.5 w-2.5" />
                truncated {topTruncationNotice}
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleCopy}
            className="h-6 px-2 text-[10px] font-sans text-[#888888] hover:text-[#ededed] hover:bg-[#222222] rounded-[4px] gap-1"
          >
            {copied ? (
              <>
                <Check className="h-3 w-3 text-emerald-400" />
                <span className="text-emerald-400">Copied</span>
              </>
            ) : (
              <>
                <Copy className="h-3 w-3" />
                <span>Copy diff</span>
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Diff body */}
      {!isCollapsed && (
        <div className="overflow-x-auto max-h-[500px] overflow-y-auto text-[11px] leading-5 divide-y divide-[#181818]/60">
          <div className="min-w-fit w-full">
            {parsedFiles.length > 0 ? (
              parsedFiles.map((file, fileIdx) => (
                <div key={fileIdx} className="divide-y divide-[#181818]/40">
                  {/* File Sub-header if multi-file */}
                  {parsedFiles.length > 1 && (
                    <div className="flex items-center gap-2 px-3 py-1.5 bg-[#111111] border-b border-[#222222] text-[#888888] text-[10px]">
                      <FileCode className="h-3 w-3 text-[#52a8ff]" />
                      <span className="text-[#cccccc] font-medium">
                        {file.to?.replace(/^b\//, '') || file.from?.replace(/^a\//, '') || `File ${fileIdx + 1}`}
                      </span>
                      {file.new && (
                        <span className="badge-success text-[9px] px-1 py-0.5 rounded">
                          <PlusCircle className="h-2.5 w-2.5" /> new file
                        </span>
                      )}
                      {file.deleted && (
                        <span className="badge-danger text-[9px] px-1 py-0.5 rounded">
                          <MinusCircle className="h-2.5 w-2.5" /> deleted
                        </span>
                      )}
                      {file.isRename && (
                        <span className="badge-warning text-[9px] px-1 py-0.5 rounded">
                          renamed {file.similarityIndex ? `(${file.similarityIndex}%)` : ''}
                        </span>
                      )}
                      {file.isBinary && (
                        <span className="text-[9px] px-1 py-0.5 rounded font-mono bg-purple-500/10 text-purple-400 border border-purple-500/20">
                          binary
                        </span>
                      )}
                      <span className="ml-auto font-sans">
                        <span className="text-status-success">+{file.additions}</span>{' '}
                        <span className="text-status-danger">-{file.deletions}</span>
                      </span>
                    </div>
                  )}

                  {/* Empty chunk banner for binary or rename-only files */}
                  {file.isBinary && (
                    <div className="p-3 text-center font-mono text-[11px] text-[#888888] bg-[#0c0c0c]">
                      Binary file (not shown in text diff)
                    </div>
                  )}
                  {file.isRename && file.chunks.length === 0 && (
                    <div className="p-3 text-center font-mono text-[11px] text-[#888888] bg-[#0c0c0c]">
                      File renamed from <span className="text-[#ededed] font-medium">{file.from}</span> to{' '}
                      <span className="text-[#ededed] font-medium">{file.to}</span> with no content changes
                    </div>
                  )}

                  {/* Chunks */}
                  {file.chunks.map((chunk, chunkIdx) => (
                    <div key={chunkIdx}>
                      {/* Hunk Header */}
                      <div className="diff-hunk-header px-3 py-1 font-mono text-[10px] tracking-tight">
                        {chunk.content}
                      </div>

                      {/* Chunk Changes */}
                      <div>
                        {chunk.changes.map((change, changeIdx) => {
                          const isAdd = change.type === 'add';
                          const isDel = change.type === 'del';
                          const isNormal = change.type === 'normal';

                          // Line numbers
                          let oldNum: number | undefined;
                          let newNum: number | undefined;

                          if (isNormal) {
                            oldNum = change.ln1;
                            newNum = change.ln2;
                          } else if (isDel) {
                            oldNum = change.ln;
                          } else if (isAdd) {
                            newNum = change.ln;
                          }

                          const isTruncation = /^<truncated\s+.*>$/i.test(change.content.trim());
                          if (isTruncation) {
                            return (
                              <div
                                key={changeIdx}
                                className="flex items-center gap-2 px-3 py-1 bg-amber-500/10 border-y border-amber-500/20 text-amber-400 font-mono text-[10px]"
                              >
                                <AlertTriangle className="h-3 w-3 shrink-0" />
                                <span>{change.content.trim()}</span>
                              </div>
                            );
                          }

                          // Strip leading '+', '-', or ' ' from change content
                          let lineText = change.content;
                          if (lineText.startsWith('+') || lineText.startsWith('-') || lineText.startsWith(' ')) {
                            lineText = lineText.slice(1);
                          }

                          return (
                            <div
                              key={changeIdx}
                              className={`flex items-start transition-colors min-h-[20px] ${
                                isAdd
                                  ? 'diff-line-add'
                                  : isDel
                                  ? 'diff-line-del'
                                  : 'text-[#a1a1a1] hover:bg-white/[0.02] bg-[#0d0d0d]'
                              }`}
                            >
                              {/* Line numbers gutter (sticky left edge) */}
                              <div
                                className={cn(
                                  'sticky left-0 flex items-center z-10 shrink-0 select-none font-mono border-r border-[#222222]',
                                  isAdd ? 'bg-[#0f1a14]' : isDel ? 'bg-[#1c0f12]' : 'bg-[#0d0d0d]'
                                )}
                              >
                                <div className="w-10 shrink-0 text-right pr-2 py-0.5 text-[10px] text-[#555555]">
                                  {oldNum ?? ''}
                                </div>
                                <div className="w-10 shrink-0 text-right pr-2 py-0.5 text-[10px] text-[#555555]">
                                  {newNum ?? ''}
                                </div>
                                {/* Sign indicator */}
                                <div className="w-5 shrink-0 text-center py-0.5 font-bold text-[11px]">
                                  {isAdd ? '+' : isDel ? '-' : ' '}
                                </div>
                              </div>

                              {/* Line content */}
                              <div className="flex-1 py-0.5 px-3 whitespace-pre font-mono">
                                {lineText || '\u00A0'}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  ))}
                </div>
              ))
            ) : (
              // Fallback for unparseable diffs: render raw lines
              <div className="p-3 font-mono text-[11px] text-[#a1a1a1] whitespace-pre overflow-x-auto">
                {cleanedDiff}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
