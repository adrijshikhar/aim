import * as React from 'react';
import { AlertTriangle, RotateCw } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallbackTitle?: string;
  onReset?: () => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
    if (this.props.onReset) {
      this.props.onReset();
    }
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center min-h-[360px] p-8 text-center rounded-geist border border-status-danger-border/60 bg-status-danger-bg/20 space-y-4">
          <div className="h-12 w-12 rounded-full bg-status-danger-bg border border-status-danger-border flex items-center justify-center text-status-danger">
            <AlertTriangle className="h-6 w-6" />
          </div>
          <div className="space-y-1.5 max-w-md">
            <h3 className="text-base font-semibold text-foreground font-sans">
              {this.props.fallbackTitle || 'Something went wrong in this view'}
            </h3>
            <p className="text-xs text-muted-foreground font-mono leading-relaxed break-words">
              {this.state.error?.message || 'An unexpected rendering error occurred.'}
            </p>
          </div>
          <div className="flex items-center gap-2 pt-2">
            <Button
              size="sm"
              variant="outline"
              onClick={this.handleRetry}
              className="h-8 text-xs font-sans rounded-geist border-border bg-secondary hover:bg-muted text-foreground cursor-pointer flex items-center gap-1.5"
            >
              <RotateCw className="h-3.5 w-3.5" />
              <span>Try Again</span>
            </Button>
            <Button
              size="sm"
              onClick={() => window.location.reload()}
              className="h-8 text-xs font-sans rounded-geist bg-foreground text-background hover:bg-foreground/90 font-medium cursor-pointer"
            >
              Reload Page
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
