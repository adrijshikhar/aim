import {
  Toast,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from '@/components/ui/toast';
import { useToast } from '@/components/ui/use-toast';
import { CheckCircle2, AlertCircle, AlertTriangle, Info } from 'lucide-react';

export function Toaster() {
  const { toasts } = useToast();

  const getVariantIcon = (variant?: string | null) => {
    switch (variant) {
      case 'success':
        return <CheckCircle2 className="h-4 w-4 text-[#98c379] shrink-0 mt-0.5" />;
      case 'destructive':
        return <AlertCircle className="h-4 w-4 text-[#e06c75] shrink-0 mt-0.5" />;
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-[#e5c07b] shrink-0 mt-0.5" />;
      default:
        return <Info className="h-4 w-4 text-[#52a8ff] shrink-0 mt-0.5" />;
    }
  };

  return (
    <ToastProvider>
      {toasts.map(function ({ id, title, description, action, variant, ...props }) {
        return (
          <Toast key={id} variant={variant} {...props}>
            <div className="flex items-start gap-2.5 min-w-0 flex-1">
              {getVariantIcon(variant)}
              <div className="grid gap-0.5 min-w-0 flex-1">
                {title && <ToastTitle>{title}</ToastTitle>}
                {description && (
                  <ToastDescription>{description}</ToastDescription>
                )}
              </div>
            </div>
            {action}
            <ToastClose />
          </Toast>
        );
      })}
      <ToastViewport />
    </ToastProvider>
  );
}
