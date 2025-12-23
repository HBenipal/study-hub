import React from "react";
import styles from "./MarkdownPreview.module.css";

interface MarkdownPreviewProps {}

export const MarkdownPreview = React.forwardRef<
  HTMLDivElement,
  MarkdownPreviewProps
>((_props, ref) => {
  return (
    <div
      ref={ref}
      className={`${styles.preview} flex-1 p-5 overflow-y-auto leading-relaxed`}
    ></div>
  );
});

MarkdownPreview.displayName = "MarkdownPreview";
