import React, { useState } from 'react';
import { FileCode, Copy, Check, ChevronDown, ChevronRight } from 'lucide-react';
import { Button } from './ui/button';

interface DiffViewerProps {
  diffText: string;
  filePath?: string;
  defaultCollapsed?: boolean;
}

interface ParsedDiffLine {
  type: 'add' | 'del' | 'hunk' | 'meta' | 'context';
  text: string;
  oldNum?: number;
  newNum?: number;
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

  // Parse lines and compute stats
  const { lines, adds, dels } = React.useMemo(() => {
    const rawLines = cleanedDiff.split('\n');
    const parsed: ParsedDiffLine[] = [];
    let addCount = 0;
    let delCount = 0;

    let curOld = 0;
    let curNew = 0;

    for (const line of rawLines) {
      if (line.startsWith('@@')) {
        // Hunk header: @@ -oldStart,oldCount +newStart,newCount @@
        const match = line.match(/@@\s*-(\d+)(?:,\d+)?\s*\+(\d+)(?:,\d+)?\s*@@/);
        if (match) {
          curOld = parseInt(match[1], 10);
          curNew = parseInt(match[2], 10);
        }
        parsed.push({ type: 'hunk', text: line });
      } else if (line.startsWith('---') || line.startsWith('+++')) {
        parsed.push({ type: 'meta', text: line });
      } else if (line.startsWith('+')) {
        addCount++;
        parsed.push({ type: 'add', text: line, newNum: curNew++ });
      } else if (line.startsWith('-')) {
        delCount++;
        parsed.push({ type: 'del', text: line, oldNum: curOld++ });
      } else {
        parsed.push({
          type: 'context',
          text: line,
          oldNum: curOld ? curOld++ : undefined,
          newNum: curNew ? curNew++ : undefined,
        });
      }
    }

    return { lines: parsed, adds: addCount, dels: delCount };
  }, [cleanedDiff]);

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigator.clipboard.writeText(cleanedDiff);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const fileName = filePath ? filePath.split('/').pop() : undefined;

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
          <FileCode className="h-3.5 w-3.5 text-sky-400 shrink-0" />
          <span className="text-[#ededed] font-medium text-[11px] truncate" title={filePath || 'Diff Changes'}>
            {fileName || filePath || 'File Changes'}
          </span>
          <div className="flex items-center gap-1.5 ml-1">
            {adds > 0 && (
              <span className="text-[10px] font-sans font-medium px-1.5 py-0.2 rounded bg-[#00250a] text-[#00ca52] border border-[#003814]">
                +{adds}
              </span>
            )}
            {dels > 0 && (
              <span className="text-[10px] font-sans font-medium px-1.5 py-0.2 rounded bg-[#330a11] text-[#ff5e63] border border-[#5d0e17]">
                -{dels}
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
        <div className="overflow-x-auto max-h-[500px] overflow-y-auto divide-y divide-[#181818]/60 text-[11px] leading-5">
          {lines.map((line, idx) => {
            if (line.type === 'hunk') {
              return (
                <div
                  key={idx}
                  className="px-3 py-1 bg-[#121620] text-sky-400/90 font-mono text-[10px] tracking-tight border-y border-sky-500/20 select-none"
                >
                  {line.text}
                </div>
              );
            }

            if (line.type === 'meta') {
              return (
                <div key={idx} className="px-3 py-0.5 text-[#666666] font-mono text-[10px] bg-[#0f0f0f]">
                  {line.text}
                </div>
              );
            }

            const isAdd = line.type === 'add';
            const isDel = line.type === 'del';

            return (
              <div
                key={idx}
                className={`flex items-start hover:bg-white/[0.02] transition-colors ${
                  isAdd
                    ? 'bg-[#00250a]/70 text-[#00ca52]'
                    : isDel
                    ? 'bg-[#330a11]/70 text-[#ff5e63]'
                    : 'text-[#a1a1a1]'
                }`}
              >
                {/* Line numbers */}
                <div className="w-10 shrink-0 text-right pr-2 py-0.5 text-[10px] text-[#555555] select-none font-mono">
                  {line.oldNum ?? ''}
                </div>
                <div className="w-10 shrink-0 text-right pr-2 py-0.5 text-[10px] text-[#555555] select-none font-mono border-r border-[#222222]">
                  {line.newNum ?? ''}
                </div>

                {/* Sign indicator */}
                <div className="w-5 shrink-0 text-center py-0.5 font-bold select-none text-[11px]">
                  {isAdd ? '+' : isDel ? '-' : ' '}
                </div>

                {/* Line content */}
                <div className="flex-1 py-0.5 pr-3 whitespace-pre overflow-x-auto break-all">
                  {line.text.slice(line.type === 'add' || line.type === 'del' ? 1 : 0)}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
