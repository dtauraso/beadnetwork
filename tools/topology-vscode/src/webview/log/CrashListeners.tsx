import { useEffect } from "react";
import { postLog } from "./post";

export function CrashListeners(): null {
  useEffect(() => {
    const onError = (e: ErrorEvent) => {
      postLog("window-error", {
        message: e.message,
        filename: e.filename,
        lineno: e.lineno,
        colno: e.colno,
        stack: (e.error as Error | undefined)?.stack ?? "",
      });
    };
    const onRejection = (e: PromiseRejectionEvent) => {
      const reason = e.reason as { message?: string; stack?: string } | undefined;
      postLog("unhandled-rejection", {
        message: reason?.message ?? String(e.reason),
        stack: reason?.stack ?? "",
      });
    };
    window.addEventListener("error", onError);
    window.addEventListener("unhandledrejection", onRejection);
    return () => {
      window.removeEventListener("error", onError);
      window.removeEventListener("unhandledrejection", onRejection);
    };
  }, []);
  return null;
}
