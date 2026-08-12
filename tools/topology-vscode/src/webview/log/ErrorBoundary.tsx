import { Component, type ErrorInfo, type ReactNode } from "react";
import { postLog } from "./post";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    postLog("render-error", {
      message: error.message,
      stack: error.stack ?? "",
      componentStack: info.componentStack ?? "",
    });
  }

  render(): ReactNode {
    if (this.state.error) {
      return (
        <div style={{ padding: 16, fontFamily: "monospace", color: "#c33" }}>
          webview render error: {this.state.error.message}
        </div>
      );
    }
    return this.props.children;
  }
}
