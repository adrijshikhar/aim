import React, { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Copy, Check, Terminal, Clock, FileEdit } from 'lucide-react';
import { DiffViewer } from './DiffViewer';
import { Button } from './ui/button';

interface TurnMessageRendererProps {
  content: string;
}

interface CodeBlockProps {
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
}

function CodeBlock({ inline, className, children, ...props }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const match = /language-(\w+)/.exec(className || '');
  const language = match ? match[1] : '';
  const text = String(children).replace(/\n$/, '');

  if (inline) {
    return (
      <code
        className="px-1.5 py-0.5 rounded bg-[#181818] border border-[#262626] font-mono text-[11px] text-[#ededed] break-all"
        {...props}
      >
        {children}
      </code>
    );
  }

  // If this code block is specifically marked as diff
  if (language === 'diff' || text.startsWith('@@') || (text.includes('\n+') && text.includes('\n-'))) {
    return <DiffViewer diffText={text} />;
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="rounded-[6px] border border-[#262626] bg-[#0c0c0c] overflow-hidden my-3 shadow-sm">
      <div className="flex items-center justify-between px-3 py-1.5 bg-[#141414] border-b border-[#262626]">
        <span className="text-[10px] font-mono font-medium text-[#888888] uppercase tracking-wider">
          {language || 'code'}
        </span>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={handleCopy}
          className="h-5 px-1.5 text-[10px] font-sans text-[#888888] hover:text-[#ededed] hover:bg-[#222222] rounded-[3px] gap-1"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3 text-emerald-400" />
              <span className="text-emerald-400">Copied</span>
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

  // Check if content has diff blocks
  if (diffBlockRegex.test(content)) {
    // Reset regex index
    diffBlockRegex.lastIndex = 0;
    const segments: React.ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;

    let segmentKey = 0;
    while ((match = diffBlockRegex.exec(content)) !== null) {
      const precedingText = content.substring(lastIndex, match.index);
      const diffContent = match[1];
      lastIndex = match.index + match[0].length;

      // Detect file path from preceding text if available:
      // "to /Users/nemesis/.../foo.ts:"
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

  // 2. Check for unified diff if starting with diff markers
  if (
    content.includes('--- a/') &&
    content.includes('+++ b/') &&
    content.includes('@@')
  ) {
    const filePathMatch = /\+\+\+ b\/([^\s\n]+)/.exec(content);
    return <DiffViewer diffText={content} filePath={filePathMatch ? filePathMatch[1] : undefined} />;
  }

  // 3. Render standard structured text
  return <StructuredTextBlock text={content} />;
}

/**
 * Parses tool output preambles, command exit codes, and renders markdown.
 */
function StructuredTextBlock({ text }: { text: string }) {
  // Check for common tool execution preamble:
  // Created At: ...
  // Completed At: ...
  // The command exited with code ...
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
      // Empty line after preamble
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
              className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-mono font-medium border ${
                exitCode === 0
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                  : 'bg-[#e5484d]/10 text-[#f87171] border-[#e5484d]/20'
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
        <div className="prose-subtle text-xs text-[#ededed] font-sans leading-relaxed selection:bg-[#333333]">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              // Headings
              h1: ({ children }) => (
                <h1 className="text-sm font-semibold text-white mt-4 mb-2 pb-1 border-b border-[#262626]">
                  {children}
                </h1>
              ),
              h2: ({ children }) => (
                <h2 className="text-xs font-semibold text-white mt-3 mb-1.5">
                  {children}
                </h2>
              ),
              h3: ({ children }) => (
                <h3 className="text-xs font-medium text-[#ededed] mt-2 mb-1">
                  {children}
                </h3>
              ),
              // Paragraphs and Blocks
              p: ({ children }) => (
                <div className="mb-2 last:mb-0 leading-relaxed text-[#d4d4d4]">
                  {children}
                </div>
              ),
              pre: ({ children }) => <>{children}</>,
              // Lists
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
              // Blockquote
              blockquote: ({ children }) => (
                <blockquote className="border-l-2 border-[#333333] pl-3 py-0.5 text-[#a1a1a1] italic my-2">
                  {children}
                </blockquote>
              ),
              // Table
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
              // Links
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
              // Code
              code: CodeBlock as any,
            }}
          >
            {mainContent}
          </ReactMarkdown>
        </div>
      )}
    </div>
  );
}
